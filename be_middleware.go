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
	"runtime"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	beNet "github.com/go-enjin/be/pkg/net"
	"github.com/go-enjin/be/pkg/net/headers"
	bePath "github.com/go-enjin/be/pkg/path"
	beStrings "github.com/go-enjin/be/pkg/strings"
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
			if !beStrings.StringInStrings(r.Host, e.eb.domains...) {
				log.WarnF("rejecting unsupported domain: %v", r.Host)
				e.ServeNotFound(w, r)
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := forms.SanitizeRequestPath(r.URL.Path)
		if pg, ok := e.eb.pages[path]; ok {
			if err := e.ServePage(pg, w, r); err != nil {
				log.ErrorF("serve page err: %v", err)
				e.ServeInternalServerError(w, r)
			} else {
				log.DebugF("page served: %v", path)
			}
			return
		}
		// log.DebugF("not a page: %v, %v", path, e.eb.pages)
		next.ServeHTTP(w, r)
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
				log.DebugF("%v theme served: %v (%v)", name, path, mime)
				return
			}
		}
		// log.DebugF("not a theme static: %v", path)
		next.ServeHTTP(w, request)
	})
}

func (e *Enjin) redirectionMiddleware(next http.Handler) http.Handler {
	log.DebugF("including pages middleware")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := forms.SanitizeRequestPath(r.URL.Path)

		if rp := e.FindRedirection(path); rp != nil {
			langMode := e.SiteLanguageMode()
			reqLang := lang.GetTag(r)
			dst := langMode.ToUrl(e.SiteDefaultLanguage(), reqLang, rp.Url)
			// log.DebugF("redirecting from %v to %v", path, dst)
			e.ServeRedirect(dst, w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (e *Enjin) langMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		langMode := e.SiteLanguageMode()
		defaultTag := e.SiteDefaultLanguage()

		var reqPath string
		var requested language.Tag

		if lang.NonPageRequested(r) {
			requested = defaultTag
			reqPath = forms.SanitizeRequestPath(r.URL.Path)
			// log.DebugF("non page requested: %v", reqPath)
		} else {
			var reqOk bool
			if requested, reqPath, reqOk = langMode.FromRequest(defaultTag, r); !reqOk {
				log.WarnF("language mode rejecting request: %#v", r)
				e.Serve404(w, r) // specifically not ServeNotFound()
				return
			}
			// log.DebugF("page requested: %v", reqPath)
		}

		// TODO: determine what to do with Accept-Language request headers
		// if acceptLanguage := r.Header.Get("Accept-Language"); acceptLanguage != "" {
		// 	if parsed, err := language.Parse(acceptLanguage); err == nil {
		// 		if e.SiteSupportsLanguage(parsed) && !language.Compare(requested, parsed) {
		// 			// requested = parsed
		// 			// e.ServeRedirect(langMode.ToUrl(requested, parsed, reqPath), w, r)
		// 			// return
		// 		}
		// 	}
		// }

		if !e.SiteSupportsLanguage(requested) {
			log.WarnF("%v language not supported, using default: %v", requested, defaultTag)
			requested = defaultTag
		}

		// if pg := e.FindPage(requested, reqPath); pg == nil {
		// 	log.DebugF("page not found for: [%v] %v", requested, reqPath)
		// }
		tag, printer := lang.NewCatalogPrinter(requested.String(), e.SiteLanguageCatalog())
		ctx := context.WithValue(r.Context(), lang.LanguageTag, tag)
		ctx = context.WithValue(ctx, lang.LanguagePrinter, printer)
		ctx = context.WithValue(ctx, lang.LanguageDefault, e.SiteDefaultLanguage())
		if reqPath == "" {
			reqPath = "/"
		} else if reqPath[0] != '/' {
			reqPath = "/" + reqPath
		}
		r.URL.Path = reqPath
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
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