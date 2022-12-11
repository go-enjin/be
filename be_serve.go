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

	"github.com/iancoleman/strcase"

	"github.com/go-enjin/golang-org-x-text/language"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/request/argv"
	"github.com/go-enjin/be/pkg/types/site"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net"
	"github.com/go-enjin/be/pkg/page"
	beStrings "github.com/go-enjin/be/pkg/strings"
	"github.com/go-enjin/be/pkg/theme"
)

const ServeStatusResponseKey beContext.RequestKey = "ServeStatusResponse"

func (e *Enjin) ServeRedirect(destination string, w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, destination, http.StatusSeeOther)
}

func (e *Enjin) Serve204(w http.ResponseWriter, r *http.Request) {
	printer := lang.GetPrinterFromRequest(r)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNoContent)

	// The request was successful and there is no content in the response
	_, _ = w.Write([]byte("204 - " + printer.Sprintf("No Content")))
}

func (e *Enjin) Serve401(w http.ResponseWriter, r *http.Request) {
	printer := lang.GetPrinterFromRequest(r)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)

	// The request was rejected due to being an Unauthorized user (or anonymous guest)
	_, _ = w.Write([]byte("401 - " + printer.Sprintf("Unauthorized")))
}

func (e *Enjin) ServeBasic401(w http.ResponseWriter, r *http.Request) {
	printer := lang.GetPrinterFromRequest(r)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("WWW-Authenticate", "Basic")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte("401 - " + printer.Sprintf("Unauthorized")))
}

func (e *Enjin) Serve403(w http.ResponseWriter, r *http.Request) {
	printer := lang.GetPrinterFromRequest(r)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)

	// The request was rejected due to being Forbidden to the user (or anonymous guest)
	_, _ = w.Write([]byte("403 - " + printer.Sprintf("Forbidden")))
}

func (e *Enjin) Serve404(w http.ResponseWriter, r *http.Request) {
	printer := lang.GetPrinterFromRequest(r)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)

	// The request was rejected due to the page requested not existing
	_, _ = w.Write([]byte("404 - " + printer.Sprintf("Not Found")))
}

func (e *Enjin) Serve405(w http.ResponseWriter, r *http.Request) {
	printer := lang.GetPrinterFromRequest(r)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusMethodNotAllowed)

	// The request was rejected due to the method used bin not allowed
	_, _ = w.Write([]byte("405 - " + printer.Sprintf("Method Not Allowed")))
}

func (e *Enjin) Serve500(w http.ResponseWriter, r *http.Request) {
	printer := lang.GetPrinterFromRequest(r)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)

	// The request resulted in an internal server error
	_, _ = w.Write([]byte("500 - " + printer.Sprintf("Internal Server Error")))
}

func (e *Enjin) ServeNotFound(w http.ResponseWriter, r *http.Request) {
	e.ServeStatusPage(404, w, r)
}

func (e *Enjin) ServeInternalServerError(w http.ResponseWriter, r *http.Request) {
	e.ServeStatusPage(500, w, r)
}

func (e *Enjin) ServeStatusPage(status int, w http.ResponseWriter, r *http.Request) {
	r = r.WithContext(context.WithValue(r.Context(), ServeStatusResponseKey, status))
	reqLangTag := lang.GetTag(r)

	if path, ok := e.eb.statusPages[status]; ok {
		if pg := e.FindPage(reqLangTag, path); pg != nil {
			pg.Context.SetSpecific(argv.RequestArgvIgnoredKey, true)
			if err := e.ServePage(pg, w, r); err != nil {
				log.DebugF("error serving %v (pages) page: %v - %v", status, path, err)
			} else {
				log.DebugF("served %v (pages) page: %v", status, path)
				return
			}
		}
		for _, f := range e.eb.features {
			if mf, ok := f.(feature.Middleware); ok {
				if err := mf.ServePath(path, e, w, r); err != nil {
					log.DebugF("error serving %v (%v middleware) page: %v - %v", status, f.Tag(), path, err)
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

func (e *Enjin) ServePath(urlPath string, w http.ResponseWriter, r *http.Request) (err error) {

	pages := e.Pages()
	if p, ok := pages[urlPath]; ok {
		if err = e.ServePage(p, w, r); err == nil {
			// eb page found
			return
		}
	}

	for _, f := range e.Features() {
		if mw, ok := f.(feature.Middleware); ok {
			if err = mw.ServePath(urlPath, e, w, r); err == nil {
				// middleware found
				return
			}
		}
	}

	return
}

func (e *Enjin) ServePage(p *page.Page, w http.ResponseWriter, r *http.Request) (err error) {
	if p.Url != "" && p.Url[0] == '!' {
		err = fmt.Errorf("cannot serve not-path page: %v", p.Url)
		return
	}

	for _, f := range e.Features() {
		if ptp, ok := f.(feature.PageTypeProcessor); ok {
			var pg *page.Page
			var redirect string
			var processed bool
			if pg, redirect, processed, err = ptp.ProcessRequestPageType(r, p); err != nil {
				return
			} else if redirect != "" {
				e.ServeRedirect(redirect, w, r)
				return
			} else if processed {
				p = pg
				break
			}
		}
	}

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

	ctx.SetSpecific("SiteInfo", site.Info(e))
	ctx.SetSpecific("SiteEnjin", site.Enjin(e))

	reqLangTag := lang.GetTag(r)
	ctx.SetSpecific("ReqLangTag", reqLangTag)
	ctx.Set("Request", map[string]string{
		"URL":        r.URL.String(),
		"Path":       r.URL.Path,
		"Host":       r.Host,
		"Method":     r.Method,
		"RequestURI": r.RequestURI,
		"RemoteAddr": r.RemoteAddr,
		"Referer":    r.Referer(),
		"UserAgent":  r.UserAgent(),
		"Language":   reqLangTag.String(),
	})

	ctx.SetSpecific("BaseUrl", net.BaseURL(r))
	ctx.SetSpecific("LangPrinter", lang.GetPrinterFromRequest(r))
	ctx.SetSpecific(string(argv.RequestArgvKey), argv.GetRequestArgv(r))

	parsedTag := e.eb.defaultLang
	if v := ctx.Get("Language"); v != nil {
		if pageLang, ok := v.(string); ok {
			if pageLang != "" {
				if tag, ee := language.Parse(pageLang); ee == nil {
					parsedTag = tag
				} else {
					log.ErrorF("invalid language tag: %v - %v", pageLang, ee)
				}
			}
		} else {
			log.ErrorF("page language tag not a string: %T %+v", v, v)
		}
	}
	ctx.SetSpecific("Language", parsedTag.String())
	ctx.SetSpecific("LanguageTag", parsedTag)

	fpcPgCtx := p.Context.Copy()
	fpcPgCtx.SetSpecific("Content", p.Content)
	for _, f := range e.eb.features {
		if s, ok := f.(feature.PageContextModifier); ok {
			log.DebugF("filtering page context with: %v", f.Tag())
			ctx = s.FilterPageContext(ctx, fpcPgCtx, r)
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
			for name, m := range mp.GetMenus(reqLangTag) {
				camel := strcase.ToCamel(name)
				allMenus[camel] = m
				log.DebugF("providing [%v] menu: %v (.SiteMenu.%v)", reqLangTag.String(), name, camel)
			}
		}
	}

	if len(allMenus) > 0 {
		ctx.SetSpecific("SiteMenu", allMenus)
	}

	var data []byte
	var redirect string
	if data, redirect, err = t.RenderPage(ctx, p); err != nil {
		log.ErrorF("error rendering page: %v - %v", p.Url, err)
		return
	} else if redirect != "" {
		log.DebugF("redirecting from RenderPage: %v - %v", p.Url, redirect)
		e.ServeRedirect(redirect, w, r)
		return
	}
	if cacheControl := p.Context.String("CacheControl", ""); cacheControl != "" {
		r = r.WithContext(context.WithValue(r.Context(), "Cache-Control", cacheControl))
	}
	mime := ctx.String("ContentType", "text/html; charset=utf-8")
	contentDisposition := ctx.String("ContentDisposition", "inline")
	r = r.WithContext(context.WithValue(r.Context(), "Content-Disposition", contentDisposition))
	e.ServeData(data, fmt.Sprintf("%v", mime), w, r)
	return
}

func (e *Enjin) ServeData(data []byte, mime string, w http.ResponseWriter, r *http.Request) {
	for _, f := range e.eb.features {
		if prh, ok := f.(feature.DataRestrictionHandler); ok {
			// log.DebugF("checking restricted data with: %v", f.Tag())
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

	w.Header().Set("Content-Type", mime)
	if reqArgv := argv.GetRequestArgv(r); len(reqArgv.Argv) > 0 {
		w.Header().Set("Cache-Control", "no-store")
	} else if value, ok := r.Context().Value("Cache-Control").(string); ok {
		w.Header().Set("Cache-Control", value)
	}
	if value, ok := r.Context().Value("Content-Disposition").(string); ok {
		w.Header().Set("Content-Disposition", value)
	} else {
		w.Header().Set("Content-Disposition", "inline")
	}

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

	var status int
	if v, ok := r.Context().Value(ServeStatusResponseKey).(int); ok {
		status = v
	} else {
		status = http.StatusOK
	}
	w.WriteHeader(status)
	_, _ = w.Write(data)
}