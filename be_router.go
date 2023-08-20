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
	"sort"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/net/headers"
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

	if e.eb.hotReload {
		log.DebugF("including hot-reload middleware")
		router.Use(e.hotReloadMiddleware)
	}

	// request modifier features are expected to modify the request object
	// in-place, before any further feature processing
	if count := len(e.eb.fRequestModifiers); count > 0 {
		var tags feature.Tags
		for _, rm := range e.eb.fRequestModifiers {
			tags = append(tags, rm.Tag())
		}
		log.DebugF("including %d request modifier middlewares: %+v", count, tags)
		router.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				for _, rm := range e.eb.fRequestModifiers {
					rm.ModifyRequest(w, r)
				}
				next.ServeHTTP(w, r)
			})
		})
	}

	// request rewriter features are expected to return a modified request
	// object, potentially dropping data if requests are modified WithContext
	// and not Clone
	if count := len(e.eb.fRequestRewriters); count > 0 {
		var tags feature.Tags
		for _, rr := range e.eb.fRequestRewriters {
			tags = append(tags, rr.Tag())
		}
		log.DebugF("including %d request rewriter middlewares: %+v", count, tags)
		router.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				for _, rr := range e.eb.fRequestRewriters {
					r = rr.RewriteRequest(w, r)
				}
				next.ServeHTTP(w, r)
			})
		})
	}

	// logging after requests modified so proxy has a chance to populate ip
	router.Use(middleware.Logger)

	// gzip compression for default compressible content types
	router.Use(middleware.Compress(5))

	// these should be request modifiers instead of enjin middleware
	router.Use(e.langMiddleware)
	router.Use(e.redirectionMiddleware)
	router.Use(e.headersMiddleware)

	// operational security measures
	router.Use(e.domainsMiddleware)
	router.Use(e.permissionsPolicy.PrepareRequestMiddleware)
	router.Use(e.contentSecurityPolicy.PrepareRequestMiddleware)
	router.Use(e.requestFiltersMiddleware)

	// header policy modifier features do not block next.ServeHTTP calls and
	// must happen before blocking middleware features (ones that may not call
	// next.ServeHTTP having already served the response)
	for _, ppm := range e.eb.fPermissionsPolicyModifiers {
		log.DebugF("including %v modify permissions policy middleware", ppm.Tag())
		router.Use(e.permissionsPolicy.ModifyPolicyMiddleware(ppm.ModifyPermissionsPolicy))
	}
	for _, cspm := range e.eb.fContentSecurityPolicyModifiers {
		log.DebugF("including %v modify content security policy middleware", cspm.Tag())
		router.Use(e.contentSecurityPolicy.ModifyPolicyMiddleware(cspm.ModifyContentSecurityPolicy))
	}

	// theme static files [blocking middleware]
	if t, ee := e.GetTheme(); ee != nil {
		log.WarnF("not including any theme middleware: %v", ee)
	} else {
		if t.StaticFS() != nil {
			router.Use(t.Middleware)
		}
		if tp := t.GetParent(); tp != nil {
			if tp.StaticFS() != nil {
				router.Use(tp.Middleware)
			}
		}
	}

	router.Use(e.userAuthMiddleware)

	// potentially blocking middleware features that do not require standard
	// page rendering or data response facilities
	for _, um := range e.eb.fUseMiddlewares {
		log.DebugF("including %v use middleware", um.Tag())
		if mw := um.Use(e); mw != nil {
			router.Use(mw)
		}
	}

	// header modifier features that happen after potentially blocking features
	// that did not actually serve a response
	for _, hm := range e.eb.fHeadersModifiers {
		log.DebugF("including %v use-after modify headers middleware", hm.Tag())
		router.Use(headers.ModifyAfterUseMiddleware(hm.ModifyHeaders))
	}

	// processor middleware features are potentially blocking
	for _, proc := range e.eb.fProcessors {
		log.DebugF("including %v processor middleware", proc.Tag())
		router.Use(func(next http.Handler) http.Handler {
			innerProc := proc // prevent outer scoped proc overwriting
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				innerProc.Process(e, next, w, r)
			})
		})
	}

	// route processor middleware features, in order of longest to shortest
	sortedRoutes := maps.Keys(e.eb.processors)
	sort.Sort(beStrings.SortByLengthDesc(sortedRoutes))
	for _, route := range sortedRoutes {
		log.DebugF("including enjin %v route processor middleware", route)
		router.Use(func(next http.Handler) http.Handler {
			innerRoute := route // prevent outer scoped route overwriting
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == innerRoute {
					if e.eb.processors[innerRoute](e, w, r) {
						return
					}
				}
				next.ServeHTTP(w, r)
			})
		})
	}

	// middleware features have a final chance to apply enjin changes before
	// error handling router changes are made
	for _, am := range e.eb.fApplyMiddlewares {
		log.DebugF("including %v apply middleware", am.Tag())
		if err = am.Apply(e); err != nil {
			return
		}
	}

	// error handling router changes
	router.NotFound(e.ServeNotFound)
	router.MethodNotAllowed(e.Serve405)

	// standard page processing catch-all-not-already-routed
	if e.eb.fRoutePagesHandler != nil {
		log.DebugF("default routing handler: %s", e.eb.fRoutePagesHandler.Tag())
		router.HandleFunc("/*", e.eb.fRoutePagesHandler.RoutePage)
	} else {
		log.DebugF("default routing handler: .ServeNotFound")
		router.HandleFunc("/*", e.ServeNotFound)
	}

	return
}