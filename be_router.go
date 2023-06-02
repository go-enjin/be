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
	"github.com/go-enjin/be/pkg/feature/signaling"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/lang"
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

	// request modifier features are expected to modify the request object
	// in-place, before any further feature processing
	for _, f := range e.Features() {
		tag := f.Tag()
		if rm, ok := f.(feature.RequestModifier); ok {
			log.DebugF("including %v request modifier middleware", tag)
			router.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					rm.ModifyRequest(w, r)
					next.ServeHTTP(w, r)
				})
			})
		}
	}

	// request rewriter features are expected to return a modified request
	// object, potentially dropping data if requests are modified WithContext
	// and not Clone
	for _, f := range e.Features() {
		tag := f.Tag()
		if rm, ok := f.(feature.RequestRewriter); ok {
			log.DebugF("including %v request rewriter middleware", tag)
			router.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					r = rm.RewriteRequest(w, r)
					next.ServeHTTP(w, r)
				})
			})
		}
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
	router.Use(deny.Middleware)
	router.Use(e.domainsMiddleware)
	router.Use(e.permissionsPolicy.PrepareRequestMiddleware)
	router.Use(e.contentSecurityPolicy.PrepareRequestMiddleware)
	router.Use(e.requestFiltersMiddleware)

	// header policy modifier features do not block next.ServeHTTP calls and
	// must happen before blocking middleware features (ones that may not call
	// next.ServeHTTP having already served the response)
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

	// theme static files [blocking middleware]
	if t, ee := e.GetTheme(); ee != nil {
		log.WarnF("not including any theme middleware: %v", ee)
	} else {
		log.DebugF("including %v theme middleware", t.Name)
		router.Use(t.Middleware)
		if tp := t.GetParent(); tp != nil {
			router.Use(tp.Middleware)
		}
	}

	router.Use(e.userAuthMiddleware)

	// potentially blocking middleware features that do not require standard
	// page rendering or data response facilities
	for _, f := range e.Features() {
		if af, ok := f.(feature.UseMiddleware); ok {
			log.DebugF("including %v use middleware", f.Tag())
			if mw := af.Use(e); mw != nil {
				router.Use(mw)
			}
		}
	}

	// header modifier features that happen after potentially blocking features
	// that did not actually serve a response
	for _, f := range e.Features() {
		if hm, ok := f.(feature.HeadersModifier); ok {
			log.DebugF("including %v use-after modify headers middleware", f.Tag())
			router.Use(headers.ModifyAfterUseMiddleware(hm.ModifyHeaders))
		}
	}

	// processor middleware features are potentially blocking
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

	// route processor middleware features, in order of longest to shortest
	sortedRoutes := maps.Keys(e.eb.processors)
	sort.Sort(beStrings.SortByLengthDesc(sortedRoutes))
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

	// middleware features have a final chance to apply enjin changes before
	// error handling router changes are made
	for _, f := range e.Features() {
		if af, ok := f.(feature.ApplyMiddleware); ok {
			log.DebugF("including %v apply middleware", f.Tag())
			if err = af.Apply(e); err != nil {
				return
			}
		}
	}

	// error handling router changes
	router.NotFound(e.ServeNotFound)
	router.MethodNotAllowed(e.Serve405)

	// standard page processing catch-all-not-already-routed
	router.HandleFunc("/*", e.RoutingHTTP)

	return
}

func (e *Enjin) RoutingHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	tag := lang.GetTag(r)
	allFeatures := e.Features()

	// look for any page provider providing the requested page
	for _, pp := range feature.FindAllTypedFeatures[feature.PageProvider](allFeatures) {
		if pg := pp.FindPage(tag, path); pg != nil {
			if err := e.ServePage(pg, w, r); err == nil {
				log.DebugRF(r, "enjin router served provided page: %v", pg.Url)
				e.Emit(signaling.SignalServePage, pp.(feature.Feature).Tag().String(), pg)
				return
			} else {
				log.ErrorRF(r, "error serving provided page: %v - %v", pg.Url, err)
			}
		}
	}

	// look for any serve-path feature handling the requested page
	for _, spf := range feature.FindAllTypedFeatures[feature.ServePathFeature](allFeatures) {
		if ee := spf.ServePath(path, e, w, r); ee == nil {
			log.DebugRF(r, "%v feature served path: %v", spf.(feature.Feature).Tag(), path)
			e.Emit(signaling.SignalServePath, spf.(feature.Feature).Tag().String(), path)
			return
		}
	}

	// look for any fallback, enjin-built-in, pages
	if pg, ok := e.eb.pages[path]; ok {
		if err := e.ServePage(pg, w, r); err != nil {
			log.ErrorRF(r, "serve page err: %v", err)
			e.ServeInternalServerError(w, r)
		} else {
			log.DebugRF(r, "enjin router served page: %v", path)
			e.Emit(signaling.SignalServePage, EnjinTag.String(), pg)
		}
		return
	}

	log.DebugRF(r, "enjin router did not find any page or path for: %v", path)
	e.ServeNotFound(w, r)
}