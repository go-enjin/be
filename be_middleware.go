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
	"net"
	"net/http"
	"runtime"
	"strings"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	beNet "github.com/go-enjin/be/pkg/net"
	"github.com/go-enjin/be/pkg/net/headers"
	bePath "github.com/go-enjin/be/pkg/path"
	beStrings "github.com/go-enjin/be/pkg/strings"
	"github.com/go-enjin/be/pkg/types/site"
	"github.com/go-enjin/golang-org-x-text/language"
)

func (e *Enjin) requestFiltersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		remoteAddr, _ := beNet.GetIpFromRequest(r)
		for _, f := range e.eb.features {
			if rf, ok := f.(feature.RequestFilter); ok {
				if err := rf.FilterRequest(r); err != nil {
					log.WarnF("filtering request from: %v - %v", remoteAddr, err)
					e.Serve404(w, r)
					return
				} else {
					log.DebugF("allowing request from: %v", remoteAddr)
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (e *Enjin) domainsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(e.eb.domains) > 0 {
			var err error
			var host string
			if host, _, err = net.SplitHostPort(r.Host); err != nil {
				host = r.Host
			}
			if !beStrings.StringInStrings(host, e.eb.domains...) {
				e.Serve404(w, r)
				log.WarnF("ignoring request for unsupported domain: %v", host)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (e *Enjin) headersMiddleware(next http.Handler) http.Handler {
	return headers.ModifyMiddleware(e.modifyHeadersFn)(next)
}

func (e *Enjin) modifyHeadersFn(request *http.Request, headers map[string]string) map[string]string {
	headers["Server"] = e.ServerName()
	for _, fn := range e.eb.headers {
		headers = fn(request, headers)
	}
	return headers
}

func (e *Enjin) pagesMiddleware(next http.Handler) http.Handler {
	log.DebugF("including pages middleware")
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		path := request.URL.Path
		if p, ok := e.eb.pages[path]; ok {
			if err := e.ServePage(p, w, request); err == nil {
				log.DebugF("page served: %v", path)
				return
			} else {
				log.ErrorF("serve page err: %v", err)
			}
		}
		// log.DebugF("not a page: %v, %v", path, e.eb.pages)
		next.ServeHTTP(w, request)
	})
}

func (e *Enjin) themeMiddleware(next http.Handler) http.Handler {
	keys := e.ThemeNames()
	log.DebugF("including theme middleware: %v", keys)
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		path := fmt.Sprintf("static/%v", bePath.TrimSlashes(request.URL.Path))
		var err error
		var data []byte
		var mime string
		var name string
		for _, k := range keys {
			t := e.eb.theming[k]
			if data, err = t.FileSystem.ReadFile(path); err == nil {
				name = t.Config.Name
				mime, _ = t.FileSystem.MimeType(path)
				e.ServeData(data, mime, w, request)
				log.DebugF("%v served: %v (%v)", name, path, mime)
				return
			}
		}
		// log.DebugF("not a theme static: %v", path)
		next.ServeHTTP(w, request)
	})
}

func langMiddleware(e site.Enjin) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requested := e.SiteDefaultLanguage().String()
			switch e.SiteLanguageMode() {
			case "query":
				langValues, ok := r.URL.Query()["lang"]
				if ok && len(langValues) >= 1 {
					if tag, err := language.Parse(langValues[0]); err != nil {
						log.ErrorF("error parsing language tag: %v", langValues[0])
					} else {
						requested = tag.String()
					}
				}
			case "domain":
				log.ErrorF("domain language mode not implemented")
			case "path":
				path := bePath.TrimSlashes(r.URL.Path)
				parts := strings.Split(path, "/")
				if len(parts) >= 1 {
					if pathTag, err := language.Parse(parts[0]); err == nil {
						for _, tag := range e.SiteLocales() {
							if language.Compare(tag, pathTag) {
								requested = tag.String()
								r.URL.Path = strings.TrimPrefix(r.URL.Path, "/"+tag.String())
								if r.URL.Path == "" {
									r.URL.Path = "/"
								}
								break
							}
						}
					}
				}
			}

			tag, printer := lang.NewCatalogPrinter(requested, e.SiteLanguageCatalog())
			ctx := context.WithValue(r.Context(), lang.LanguageTag, tag)
			ctx = context.WithValue(ctx, lang.LanguagePrinter, printer)
			ctx = context.WithValue(ctx, lang.LanguageDefault, e.SiteDefaultLanguage())
			log.TraceF("client requested language: %v", tag)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (e *Enjin) panicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				buf := make([]byte, 1<<16)
				n := runtime.Stack(buf, false)
				buf = buf[:n]

				log.ErrorF("recovering from panic: %v\n(begin stacktrace)\n%s\n(end stacktrace)", err, buf)

				defer func() {
					if ee := recover(); ee != nil {
						log.ErrorF("recovering from secondary panic")
						e.Serve500(w, r)
					}
				}()
				e.ServeInternalServerError(w, r)
			}
		}()
		next.ServeHTTP(w, r)
	})
}