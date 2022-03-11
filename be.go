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

// TODO: redirect management
// TODO: output formats
// TODO: language
// TODO: page kinds: single, list
// TODO: output formats: page, home, taxonomy, term, section
// TODO: page layout settings
// TODO: build flags to configure database support
// TODO: allow/deny direct connections
// TODO: allow/deny proxy connections (all, any, cloudflare, atlassian)
// TODO: allow/deny requests (atlassian-connect stuff?)
// DONE: tmpl content format

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nanmu42/gzip"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/hash/sha"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net/gorilla-handlers"
	"github.com/go-enjin/be/pkg/net/headers"
	"github.com/go-enjin/be/pkg/net/ip/deny"
	"github.com/go-enjin/be/pkg/slug"
)

var _ feature.Runner = &Enjin{}
var _ feature.System = &Enjin{}
var _ feature.Service = &Enjin{}

type Enjin struct {
	port       int
	listen     string
	prefix     string
	production bool

	debug bool

	be     *EnjinBuilder
	cli    *cli.App
	router *chi.Mux
}

func newEnjin(be *EnjinBuilder) *Enjin {
	r := &Enjin{
		be: be,
		cli: &cli.App{
			Name:    globals.BinName,
			Usage:   globals.Summary,
			Version: globals.Version,
			Flags:   be.flags,
		},
		router: chi.NewRouter(),
	}
	r.cli.Action = r.Action
	return r
}

func (e *Enjin) Action(ctx *cli.Context) (err error) {
	if len(e.be.theming) == 0 {
		err = fmt.Errorf("builder error: at least one theme is required")
		return
	}
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
	if e.debug {
		log.Config.LogLevel = log.LevelDebug
		log.Config.Apply()
	} else if ctx.Bool("quiet") {
		log.Config.LogLevel = log.LevelWarn
		log.Config.Apply()
	}
	middleware.DefaultLogger = func(next http.Handler) http.Handler {
		return handlers.LoggingHandler(log.InfoWriter(), next)
	}
	deny.DenyDuration = ctx.Int64("deny-duration")
	for _, f := range e.be.features {
		if err = f.Startup(ctx); err != nil {
			return
		}
	}

	if domains := ctx.StringSlice("domain"); domains != nil {
		e.be.domains = append(e.be.domains, domains...)
	}

	eicPrefix := "enjin integrity checks"
	eicLogMsg := func(status, format string, argv ...interface{}) (msg string) {
		msgFmt := fmt.Sprintf("%v %v: %v", eicPrefix, status, format)
		msg = fmt.Sprintf(msgFmt, argv...)
		return
	}
	eicFmtErr := func(format string, argv ...interface{}) (e error) {
		e = fmt.Errorf(eicLogMsg("failed", format, argv...))
		return
	}
	if e.be.slugsums {
		if slug.SlugsumsPresent() {
			var slugMap slug.ShaMap
			var imposters, extraneous, validated []string
			if slugMap, _, imposters, extraneous, validated, err = slug.ValidateSlugsumsComplete(); err != nil {
				log.ErrorF(eicLogMsg("failed", err.Error()))
				return
			}
			il := len(imposters)
			el := len(extraneous)
			if il > 0 || (ctx.Bool("strict") && el > 0) {
				if il > 0 {
					if el > 0 {
						err = eicFmtErr("imposters: %v, extraneous: %v", imposters, extraneous)
						log.ErrorF(eicLogMsg("failed", "summary:\n\timposter files present: %v\n\textraneous files present: %v", imposters, extraneous))
					} else {
						err = eicFmtErr("imposters: %v", imposters)
						log.ErrorF(eicLogMsg("failed", "summary:\n\timposter files present: %v", imposters))
					}
				} else {
					err = eicFmtErr("extraneous: %v", extraneous)
					log.ErrorF(eicLogMsg("failed", "summary:\n\textraneous files present: %v", extraneous))
				}
				e.NotifyF("failed", err.Error())
				return
			}
			if ctx.Bool("strict") {
				if err = slugMap.CheckSlugIntegrity(); err != nil {
					err = eicFmtErr(err.Error())
					return
				}
				log.InfoF(eicLogMsg("partial-pass", "slug integrity validated successfully"))
				if ctx.IsSet("sums-integrity") {
					globals.SumsIntegrity = strings.ToLower(ctx.String("sums-integrity"))
					if !sha.RxShasum64.MatchString(globals.SumsIntegrity) {
						err = eicFmtErr("invalid --sums-integrity value, must be 64 characters of [a-f0-9]")
						return
					}
				} else {
					err = eicFmtErr("missing --sums-integrity value (--strict present)")
					return
				}
				if err = slugMap.CheckSumsIntegrity(); err != nil {
					err = eicFmtErr(err.Error())
					return
				}
				log.InfoF(eicLogMsg("partial-pass", "sums integrity validated successfully"))
			}
			vl := len(validated)
			if el > 0 {
				log.WarnF(eicLogMsg("passed", "ignoring extraneous files present: %v", extraneous))
				e.NotifyF(eicPrefix+" passed", "successfully validated %d files (ignoring %d extraneous)", vl, el)
			} else {
				e.NotifyF(eicPrefix+" passed", "successfully validated %d files", vl)
			}
		} else if ctx.Bool("strict") {
			err = eicFmtErr("missing Slugsums file (--strict present)")
			e.NotifyF(eicPrefix+" failed", "missing Slugsums file (--strict present)")
			return
		} else {
			log.DebugF("Slugsums not found, enjin integrity checks skipped")
		}
	}

	return e.Startup()
}

func (e *Enjin) Startup() (err error) {
	deny.DenyWordPressPaths()

	c := make(chan os.Signal, 10)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		e.Shutdown()
		os.Exit(0)
	}()

	log.DebugF(e.String())

	for _, f := range e.be.features {
		if rm, ok := f.(feature.RequestModifier); ok {
			log.DebugF("including %v request modifier middleware", f.Tag())
			e.router.Use(func(next http.Handler) http.Handler {
				log.DebugF("using %v request modifier middleware", f.Tag())
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					rm.ModifyRequest(w, r)
					next.ServeHTTP(w, r)
				})
			})
		}
	}

	e.router.Use(middleware.Logger)

	e.router.Use(e.headersMiddleware)

	// operational security measures
	e.router.Use(deny.Middleware)
	e.router.Use(e.domainsMiddleware)
	e.router.Use(e.requestFiltersMiddleware)

	// theme static files
	e.router.Use(e.themeMiddleware)

	for _, f := range e.be.features {
		if mf, ok := f.(feature.Middleware); ok {
			if mw := mf.Use(e); mw != nil {
				// log.DebugF("including %v feature middleware", f.Tag())
				e.router.Use(mw)
			}
		}
	}

	for _, f := range e.be.features {
		if hm, ok := f.(feature.HeadersModifier); ok {
			log.DebugF("including %v use-after modify headers middleware", f.Tag())
			e.router.Use(headers.ModifyAfterUseMiddleware(hm.ModifyHeaders))
		}
	}

	for _, f := range e.be.features {
		if proc, ok := f.(feature.Processor); ok {
			log.DebugF("including %v processor middleware", f.Tag())
			e.router.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					proc.Process(e, next, w, r)
				})
			})
		}
	}

	for route, processor := range e.be.processors {
		log.DebugF("including enjin %v route processor middleware", route)
		e.router.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == route {
					if processor(e, w, r) {
						return
					}
				}
				next.ServeHTTP(w, r)
			})
		})
	}

	e.router.Use(e.pagesMiddleware)

	for _, f := range e.be.features {
		if mf, ok := f.(feature.Middleware); ok {
			if err = mf.Apply(e); err != nil {
				return
			}
		}
	}

	// chi needs this for whatever reason, pages can catch before this
	// so it's really just a nop so chi actually does something with
	// the middleware set
	e.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok\n"))
	})

	handler := gzip.DefaultHandler().WrapHandler(e.router)

	e.Notify("web process startup")

	addr := fmt.Sprintf("%s:%d", e.listen, e.port)
	if err = http.ListenAndServe(addr, handler); err != nil {
		e.NotifyF("startup error", "%v", err)
		return
	}
	return
}

func (e *Enjin) Shutdown() {
	for _, f := range e.be.features {
		f.Shutdown()
	}
	e.Notify("web process shutdown")
}