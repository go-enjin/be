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

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/net/headers"
	"github.com/go-enjin/be/pkg/net/ip/deny"
	"github.com/go-enjin/be/pkg/request/argv"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

func (e *Enjin) setupRouter(router *chi.Mux) (err error) {
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			r = r.Clone(context.WithValue(r.Context(), "enjin-id", e.String()))

			if e.debug {
				w.Header().Set("Server", fmt.Sprintf("%v/%v-%v", globals.BinName, globals.Version, globals.BinHash))
			} else {
				w.Header().Set("Server", globals.BinName)
			}

			path := forms.SanitizeRequestPath(r.URL.Path)
			if reqArgv := argv.DecodeHttpRequest(r); reqArgv != nil {
				r = reqArgv.Set(r)
				path = reqArgv.Path
				log.TraceF("parsed request argv: %v", reqArgv)
			}
			r.URL.Path = path

			next.ServeHTTP(w, r)
			return
		})
	})

	router.Use(middleware.RequestID)
	router.Use(e.panicMiddleware)

	for _, f := range e.Features() {
		tag := f.Tag()
		if rm, ok := f.(feature.RequestModifier); ok {
			log.DebugF("including %v request modifier middleware", tag)
			router.Use(func(next http.Handler) http.Handler {
				log.DebugF("using %v request modifier middleware", tag)
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					rm.ModifyRequest(w, r)
					next.ServeHTTP(w, r)
				})
			})
		}
	}

	for _, f := range e.Features() {
		tag := f.Tag()
		if rm, ok := f.(feature.RequestRewriter); ok {
			log.DebugF("including %v request rewriter middleware", tag)
			router.Use(func(next http.Handler) http.Handler {
				log.DebugF("using %v request rewriter middleware", tag)
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					r = rm.RewriteRequest(w, r)
					next.ServeHTTP(w, r)
				})
			})
		}
	}

	// logging after requests modified so proxy and populate ip
	router.Use(middleware.Logger)

	// these should be request modifiers instead of enjin middleware
	router.Use(e.langMiddleware)
	router.Use(e.redirectionMiddleware)
	router.Use(e.headersMiddleware)

	// operational security measures
	router.Use(deny.Middleware)
	router.Use(e.domainsMiddleware)
	router.Use(e.permissionsPolicy.PrepareRequestMiddleware)
	router.Use(e.contentSecurityPolicy.PrepareRequestMiddleware)
	router.Use(e.requestFiltersMiddleware)

	for _, f := range e.Features() {
		if ppm, ok := f.(feature.PermissionsPolicyModifier); ok {
			log.DebugF("including %v modify permissions policy middleware", f.Tag())
			router.Use(e.permissionsPolicy.ModifyPolicyMiddleware(ppm.ModifyPermissionsPolicy))
		}
		if cspm, ok := f.(feature.ContentSecurityPolicyModifier); ok {
			log.DebugF("including %v modify content security policy middleware", f.Tag())
			router.Use(e.contentSecurityPolicy.ModifyPolicyMiddleware(cspm.ModifyContentSecurityPolicy))
		}
	}

	// theme static files
	if t, ee := e.GetTheme(); ee != nil {
		log.WarnF("not including any theme middleware: %v", ee)
	} else {
		router.Use(t.Middleware)
		if tp := t.GetParent(); tp != nil {
			router.Use(tp.Middleware)
		}
	}

	for _, f := range e.Features() {
		if mf, ok := f.(feature.Middleware); ok {
			if mw := mf.Use(e); mw != nil {
				router.Use(mw)
			}
		}
	}

	for _, f := range e.Features() {
		if hm, ok := f.(feature.HeadersModifier); ok {
			log.DebugF("including %v use-after modify headers middleware", f.Tag())
			router.Use(headers.ModifyAfterUseMiddleware(hm.ModifyHeaders))
		}
	}

	for _, f := range e.Features() {
		if proc, ok := f.(feature.Processor); ok {
			log.DebugF("including %v processor middleware", f.Tag())
			router.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					proc.Process(e, next, w, r)
				})
			})
		}
	}

	sortedRoutes := beStrings.SortByLengthDesc(maps.Keys(e.eb.processors))
	for _, route := range sortedRoutes {
		log.DebugF("including enjin %v route processor middleware", route)
		processor := e.eb.processors[route]
		router.Use(func(next http.Handler) http.Handler {
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

	router.Use(e.pagesMiddleware)

	for _, f := range e.Features() {
		if mf, ok := f.(feature.Middleware); ok {
			if err = mf.Apply(e); err != nil {
				return
			}
		}
	}

	router.NotFound(e.ServeNotFound)
	router.MethodNotAllowed(e.Serve405)

	// chi needs this for whatever reason, pages can catch before this
	// so that it's really just a nop and chi actually does something with
	// the middleware set
	router.Get("/", e.Serve404)
	router.Put("/", e.Serve404)
	router.Post("/", e.Serve404)

	return
}