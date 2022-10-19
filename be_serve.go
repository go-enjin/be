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
	"fmt"
	"net/http"

	"github.com/iancoleman/strcase"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net"
	"github.com/go-enjin/be/pkg/page"
	beStrings "github.com/go-enjin/be/pkg/strings"
	"github.com/go-enjin/be/pkg/theme"
)

func (e *Enjin) Serve204(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNoContent)
	_, _ = w.Write([]byte("204 - No Content\n"))
}

func (e *Enjin) Serve401(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte("401 - Unauthorized\n"))
}

func (e *Enjin) ServeBasic401(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("WWW-Authenticate", "Basic")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte("401 - Unauthorized\n"))
}

func (e *Enjin) Serve403(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)
	_, _ = w.Write([]byte("403 - Forbidden\n"))
}

func (e *Enjin) Serve404(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte("404 - Not Found\n"))
}

func (e *Enjin) Serve405(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusMethodNotAllowed)
	_, _ = w.Write([]byte("405 - Method Not Allowed\n"))
}

func (e *Enjin) Serve500(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte("500 - Internal Server Error\n"))
}

func (e *Enjin) ServeNotFound(w http.ResponseWriter, r *http.Request) {
	e.ServeStatusPage(404, w, r)
}

func (e *Enjin) ServeInternalServerError(w http.ResponseWriter, r *http.Request) {
	e.ServeStatusPage(500, w, r)
}

func (e *Enjin) ServeStatusPage(status int, w http.ResponseWriter, r *http.Request) {
	if path, ok := e.eb.statusPages[status]; ok {
		if p, ok := e.eb.pages[path]; ok {
			if err := e.ServePage(p, w, r); err != nil {
				log.DebugF("error serving %v (pages) page: %v", status, err)
			} else {
				log.DebugF("served %v (pages) page: %v", status, path)
				return
			}
		}
		for _, f := range e.eb.features {
			if mf, ok := f.(feature.Middleware); ok {
				if err := mf.ServePath(path, e, w, r); err != nil {
					log.DebugF("error serving %v (middleware) page: %v", status, err)
				} else {
					log.DebugF("served %v (middleware) page: %v", status, path)
					return
				}
			}
		}
	}
	switch status {
	case 401:
		e.Serve401(w, r)
	case 403:
		e.Serve403(w, r)
	case 404:
		e.Serve404(w, r)
	case 405:
		e.Serve405(w, r)
	case 500:
		e.Serve500(w, r)
	default:
		log.WarnF("unsupported status page: %v, serving 404 instead", status)
		e.ServeStatusPage(404, w, r)
	}
}

func (e *Enjin) ServePage(p *page.Page, w http.ResponseWriter, r *http.Request) (err error) {
	var t *theme.Theme
	if t, err = e.GetTheme(); err != nil {
		return
	}

	if e.eb.hotReload {
		if err = t.Layouts.Reload(); err != nil {
			err = fmt.Errorf("error refreshing layout template: %v", err)
			return
		}
	}

	ctx := e.Context()
	ctx.Set("Request", map[string]string{
		"URL":        r.URL.String(),
		"Path":       r.URL.Path,
		"Host":       r.Host,
		"Method":     r.Method,
		"RequestURI": r.RequestURI,
		"RemoteAddr": r.RemoteAddr,
		"Referer":    r.Referer(),
		"UserAgent":  r.UserAgent(),
	})
	ctx.Set("BaseUrl", net.BaseURL(r))
	ctx.Apply(p.Context.Copy())

	for _, f := range e.eb.features {
		if s, ok := f.(feature.PageContextModifier); ok {
			log.DebugF("filtering page context with: %v", f.Tag())
			ctx = s.FilterPageContext(ctx, p.Context, r)
		}
	}

	for _, f := range e.eb.features {
		if prh, ok := f.(feature.PageRestrictionHandler); ok {
			log.DebugF("checking restricted pages with: %v", f.Tag())
			if ctx, r, ok = prh.RestrictServePage(ctx, w, r); !ok {
				addr, _ := net.GetIpFromRequest(r)
				log.WarnF("[restricted] permission denied %v for: %v", addr, r.URL.Path)
				e.ServeBasic401(w, r)
				return
			}
		}
	}

	allMenus := make(map[string]interface{})
	for _, f := range e.eb.features {
		if mp, ok := f.(feature.MenuProvider); ok {
			for name, m := range mp.GetMenus() {
				camel := strcase.ToCamel(name)
				allMenus[camel] = m
				log.DebugF("providing menu: %v (.SiteMenu.%v)", name, camel)
			}
		}
	}
	if len(allMenus) > 0 {
		ctx.SetSpecific("SiteMenu", allMenus)
	}

	var data []byte
	if data, err = t.RenderPage(ctx, p); err != nil {
		return
	}
	e.ServeData(data, "text/html; charset=utf-8", w, r)
	return
}

func (e *Enjin) ServeData(data []byte, mime string, w http.ResponseWriter, r *http.Request) {
	for _, f := range e.eb.features {
		if prh, ok := f.(feature.DataRestrictionHandler); ok {
			log.DebugF("checking restricted data with: %v", f.Tag())
			if r, ok = prh.RestrictServeData(data, mime, w, r); !ok {
				addr, _ := net.GetIpFromRequest(r)
				log.WarnF("[restricted] permission denied %v for: %v", addr, r.URL.Path)
				e.ServeBasic401(w, r)
				return
			}
		}
	}

	// only one translation allowed, non-feature translators take precedence
	basicMime := beStrings.GetBasicMime(mime)
	if fn, ok := e.eb.translators[basicMime]; ok {
		data, mime = fn(data)
		basicMime = beStrings.GetBasicMime(mime)
	} else {
		for _, f := range e.eb.features {
			if v, ok := f.(feature.OutputTranslator); ok {
				// log.DebugF("checking output filter: %v", f.Tag())
				if v.CanTranslate(mime) {
					if d, m, err := v.TranslateOutput(e, data, mime); err == nil {
						data, mime = d, m
						basicMime = beStrings.GetBasicMime(mime)
						log.DebugF("filtered output: %v - %v", f.Tag(), mime)
					} else {
						log.DebugF("%v error filtering output: %v", f.Tag(), err)
					}
					break
				}
			}
		}
	}

	w.Header().Set("Content-Disposition", "inline")
	w.Header().Set("Content-Type", mime)

	for _, f := range e.eb.features {
		if tfr, ok := f.(feature.OutputTransformer); ok {
			if tfr.CanTransform(mime, r) {
				data = tfr.TransformOutput(mime, data)
			}
		}
	}

	// non-feature transformers happen after features
	if fn, ok := e.eb.transformers[mime]; ok {
		data = fn(data)
	} else if fn, ok := e.eb.transformers[basicMime]; ok {
		data = fn(data)
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}