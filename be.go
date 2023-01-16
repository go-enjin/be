// Copyright (c) 2022  The Go-Enjin Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package be

// TODO: build flags to configure database support
// TODO: allow/deny direct connections
// TODO: allow/deny proxy connections (all, any, cloudflare, atlassian)
// TODO: allow/deny requests (atlas-gonnect stuff?)

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/hostrouter"
	"github.com/iancoleman/strcase"
	"github.com/nanmu42/gzip"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net/gorilla-handlers"
	"github.com/go-enjin/be/pkg/net/ip/deny"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

//go:generate _scripts/be-pkg-list.sh

var (
	_ feature.Runner    = &Enjin{}
	_ feature.System    = &Enjin{}
	_ feature.Internals = &Enjin{}
)

func init() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("%s %s\n", globals.BinName, c.App.Version)
	}
}

type Enjin struct {
	port       int
	listen     string
	prefix     string
	production bool

	debug bool

	catalog *lang.Catalog

	eb  *EnjinBuilder
	cli *cli.App

	router *chi.Mux
	enjins []*Enjin
}

func newEnjin(eb *EnjinBuilder) *Enjin {
	e := &Enjin{
		eb: eb,
		cli: &cli.App{
			Name:     globals.BinName,
			Usage:    globals.Summary,
			Version:  globals.BuildVersion(),
			Flags:    eb.flags,
			Commands: eb.commands,
		},
	}
	e.initLocales()
	e.initConsoles()
	e.setupFeatures()
	e.cli.Action = e.action
	return e
}

func newIncludedEnjin(eb *EnjinBuilder, parent *Enjin) *Enjin {
	e := &Enjin{
		eb: eb,
	}
	e.initLocales()
	e.initConsoles()
	e.setupFeatures()
	return e
}

func (e *Enjin) action(ctx *cli.Context) (err error) {
	if err = e.SetupRootEnjin(ctx); err != nil {
		return
	}

	if err = e.startupIntegrityChecks(ctx); err != nil {
		return
	}

	if err = e.startupRootService(ctx); err != nil && err.Error() == "http: Server closed" {
		err = nil
	}
	return
}

func (e *Enjin) Self() (self interface{}) {
	return e
}

func (e *Enjin) SetupRootEnjin(ctx *cli.Context) (err error) {

	e.port = ctx.Int("port")
	e.listen = ctx.String("listen")
	e.debug = ctx.Bool("debug")
	e.prefix = ctx.String("prefix")
	e.prefix = strings.ToLower(e.prefix)
	e.production = e.prefix == "" || e.prefix == "prd"
	if e.production {
		e.prefix = ""
		e.debug = false
	}

	lvl := strings.ToLower(ctx.String("log-level"))
	if v, ok := log.Levels[lvl]; ok {
		log.Config.LogLevel = v
		log.Config.Apply()
	} else {
		if lvl != "" {
			log.FatalF("invalid log-level: %v", lvl)
		}
		if e.debug {
			log.Config.LogLevel = log.LevelDebug
			log.Config.Apply()
		} else if ctx.Bool("quiet") {
			log.Config.LogLevel = log.LevelWarn
			log.Config.Apply()
		}
	}

	if len(e.eb.theming) == 0 {
		err = fmt.Errorf("builder error: at least one theme is required")
		return
	}

	middleware.DefaultLogger = func(next http.Handler) http.Handler {
		return handlers.LoggingHandler(log.InfoWriter(), next)
	}

	deny.DenyDuration = ctx.Int64("deny-duration")
	deny.DenyWordPressPaths()

	if domains := ctx.StringSlice("domain"); domains != nil && len(domains) > 0 {
		for _, domain := range domains {
			if domain != "" && !beStrings.StringInStrings(domain, e.eb.domains...) {
				e.eb.domains = append(e.eb.domains, domain)
			}
		}
	}

	for _, enjin := range e.eb.enjins {
		tag := strcase.ToKebab(enjin.tag)
		if domains := ctx.StringSlice(tag + "-domain"); len(domains) > 0 {
			for _, domain := range domains {
				if !beStrings.StringInSlices(domain, enjin.domains) {
					log.DebugF("adding domain to %v enjin: %v", tag, domain)
					enjin.domains = append(enjin.domains, domain)
				}
			}
		}
	}

	return
}

func (e *Enjin) setupFeatures() {
	for _, f := range e.Features() {
		f.Setup(e)
	}
}

func (e *Enjin) startupFeatures(ctx *cli.Context) (err error) {
	for _, f := range e.Features() {
		if err = f.Startup(ctx); err != nil {
			return
		}
	}
	return
}

func (e *Enjin) startupRootService(ctx *cli.Context) (err error) {

	if err = e.startupFeatures(ctx); err != nil {
		return
	}

	e.router = chi.NewRouter()
	if err = e.setupRouter(e.router); err != nil {
		return
	}

	if len(e.eb.enjins) == 0 {
		return e.startupHttpListener(e.listen, e.port, e.router)
	}

	hr := hostrouter.New()

	for _, eb := range e.eb.enjins {
		if len(eb.domains) == 0 {
			log.FatalDF(5, "%v enjin domains not found", eb.tag)
			continue
		}

		enjin := newIncludedEnjin(eb, e)
		e.enjins = append(e.enjins, enjin)

		if err = enjin.startupFeatures(ctx); err != nil {
			return
		}

		enjin.router = chi.NewRouter()
		if err = enjin.setupRouter(enjin.router); err != nil {
			return
		}

		for _, domain := range eb.domains {
			hr.Map(domain, enjin.router)
		}
	}

	hr.Map("*", e.router)

	root := chi.NewRouter()
	root.Mount("/", hr)
	return e.startupHttpListener(e.listen, e.port, root)
}

func (e *Enjin) startupHttpListener(listen string, port int, router *chi.Mux) (err error) {
	e.Notify("web process startup")
	log.DebugF("web process info:\n%v", e.StartupString())

	var srv http.Server

	idleConnectionsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
		<-sigint
		// We received an interrupt signal, shut down.
		if ee := srv.Shutdown(context.Background()); ee != nil {
			// Error from closing listeners, or context timeout:
			log.ErrorF("error shutting down http.Server: %v", ee)
		}
		e.Shutdown()
		close(idleConnectionsClosed)
	}()

	srv.Addr = fmt.Sprintf("%s:%d", listen, port)
	srv.Handler = gzip.DefaultHandler().WrapHandler(router)
	if err = srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.ErrorF("unexpected error listening and serving http: %v", err)
	}

	<-idleConnectionsClosed
	return
}

func (e *Enjin) Shutdown() {
	for _, f := range e.Features() {
		f.Shutdown()
	}
	e.Notify("web process shutdown")
}