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

// TODO: allow/deny direct connections
// TODO: allow/deny proxy connections (all, any, cloudflare, atlassian)
// TODO: allow/deny requests (atlas-gonnect stuff?)

import (
	"fmt"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/hostrouter"
	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/feature/signaling"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/lang/catalog"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net/headers/policy/csp"
	"github.com/go-enjin/be/pkg/net/headers/policy/permissions"
	"github.com/go-enjin/be/pkg/slices"
)

var (
	_ feature.Runner    = (*Enjin)(nil)
	_ feature.System    = (*Enjin)(nil)
	_ feature.Service   = (*Enjin)(nil)
	_ feature.Internals = (*Enjin)(nil)
)

func init() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("%s %s\n", globals.BinName, c.App.Version)
	}
}

type Enjin struct {
	prefix     string
	production bool

	debug bool

	catalog catalog.Catalog

	contentSecurityPolicy *csp.PolicyHandler
	permissionsPolicy     *permissions.PolicyHandler

	eb  *EnjinBuilder
	cli *cli.App

	router *chi.Mux
	enjins []*Enjin

	signaling     map[signaling.Signal][]*signalListener
	signalingLock *sync.RWMutex

	mutex *sync.RWMutex
}

func newEnjin(eb *EnjinBuilder) *Enjin {
	var description string
	notes := make(map[feature.Tag][]string)
	for _, f := range eb.features.List() {
		if un := f.UsageNotes(); len(un) > 0 {
			notes[f.Tag()] = un
		}
	}
	for _, ie := range eb.enjins {
		for _, f := range ie.features.List() {
			if un := f.UsageNotes(); len(un) > 0 {
				notes[f.Tag()] = un
			}
		}
	}
	if len(notes) > 0 {
		description += "Feature usage notes:\n\n"
		for _, tag := range feature.SortedFeatureTags(notes) {
			description += " * " + tag.String() + ":\n"
			for _, note := range notes[tag] {
				description += "   * " + strings.TrimSpace(note) + "\n"
			}
		}
	}
	e := &Enjin{
		eb: eb,
		cli: &cli.App{
			Name:        globals.BinName,
			Usage:       globals.Summary,
			Version:     globals.BuildVersion(),
			Description: description,
			Flags:       eb.flags,
			Commands:    eb.commands,
		},
		// policies
		contentSecurityPolicy: csp.NewPolicyHandler(),
		permissionsPolicy:     permissions.NewPolicyHandler(),
		// make variables
		catalog:       catalog.New(),
		signaling:     make(map[signaling.Signal][]*signalListener),
		signalingLock: &sync.RWMutex{},
		mutex:         &sync.RWMutex{},
	}
	e.initConsoles()
	e.setupFeatures()
	e.reloadLocales()
	e.cli.Action = e.action
	return e
}

func newIncludedEnjin(eb *EnjinBuilder, parent *Enjin) *Enjin {
	e := &Enjin{
		eb:                    eb,
		contentSecurityPolicy: csp.NewPolicyHandler(),
		permissionsPolicy:     permissions.NewPolicyHandler(),
		catalog:               catalog.New(),
		signaling:             make(map[signaling.Signal][]*signalListener),
		signalingLock:         &sync.RWMutex{},
		mutex:                 &sync.RWMutex{},
	}
	e.initConsoles()
	e.setupFeatures()
	e.reloadLocales()
	return e
}

func (e *Enjin) action(ctx *cli.Context) (err error) {
	if err = e.SetupRootEnjin(ctx); err != nil {
		return
	}

	if err = e.startupIntegrityChecks(ctx); err != nil {
		return
	}

	if err = e.startupFeatures(ctx); err != nil {
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
	} else if e.eb.fServiceListener == nil {
		err = fmt.Errorf("builder error: a feature.ServiceListener is required")
		return
	}

	if e.eb.fPanicHandler == nil {
		err = fmt.Errorf("builder error: a feature.PanicHandler is required")
		return
	}

	if e.eb.fServiceLogHandler == nil {
		err = fmt.Errorf("builder error: a feature.ServiceLogHandler is required")
		return
	} else if list := feature.FilterTyped[feature.ServiceLogger](e.eb.features.List()); len(list) == 0 {
		err = fmt.Errorf("builder error: at least one feature.ServiceLogger is required")
		return
	} else {
		middleware.DefaultLogger = e.eb.fServiceLogHandler.LogHandler
	}

	if domains := ctx.StringSlice("domain"); domains != nil && len(domains) > 0 {
		for _, domain := range domains {
			if domain != "" && !slices.Present(domain, e.eb.domains...) {
				e.eb.domains = append(e.eb.domains, domain)
			}
		}
	}

	for _, enjin := range e.eb.enjins {
		tag := strcase.ToKebab(enjin.tag)
		if domains := ctx.StringSlice(tag + "-domain"); len(domains) > 0 {
			for _, domain := range domains {
				if !slices.Within(domain, enjin.domains) {
					log.DebugF("adding domain to %v enjin: %v", tag, domain)
					enjin.domains = append(enjin.domains, domain)
				}
			}
		}
	}

	return
}

func (e *Enjin) setupFeatures() {
	for _, f := range e.eb.features.List() {
		f.Setup(e)
	}
}

func (e *Enjin) startupFeatures(ctx *cli.Context) (err error) {
	for _, f := range e.eb.features.List() {
		if err = f.Startup(ctx); err != nil {
			return
		}
	}
	return
}

func (e *Enjin) startupRootService(ctx *cli.Context) (err error) {

	e.router = chi.NewRouter()
	if err = e.setupRouter(e.router); err != nil {
		return
	}

	if len(e.eb.enjins) == 0 {
		return e.eb.fServiceListener.StartListening(e.router, e)
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
	return e.eb.fServiceListener.StartListening(root, e)
}

func (e *Enjin) Shutdown() {
	for _, f := range e.eb.features.List() {
		f.Shutdown()
	}
	if err := e.eb.fServiceListener.StopListening(); err != nil {
		log.ErrorF("error stopping http listener: %v - %v", e.eb.fServiceListener.Tag(), err)
	}
	e.Notify("enjin shutdown")
}