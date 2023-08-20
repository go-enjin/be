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
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iancoleman/strcase"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net"
	"github.com/go-enjin/be/pkg/net/headers/policy/csp"
	"github.com/go-enjin/be/pkg/net/serve"
	"github.com/go-enjin/be/pkg/request/argv"
	beStrings "github.com/go-enjin/be/pkg/strings"
	"github.com/go-enjin/be/pkg/userbase"
)

func (e *Enjin) ServeRedirect(destination string, w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, destination, http.StatusSeeOther)
}

func (e *Enjin) Serve204(w http.ResponseWriter, r *http.Request) {
	serve.Serve204(w, r)
}

func (e *Enjin) Serve400(w http.ResponseWriter, r *http.Request) {
	serve.Serve400(w, r)
}

func (e *Enjin) Serve401(w http.ResponseWriter, r *http.Request) {
	serve.Serve401(w, r)
}

func (e *Enjin) ServeBasic401(w http.ResponseWriter, r *http.Request) {
	serve.ServeBasic401(w, r)
}

func (e *Enjin) Serve403(w http.ResponseWriter, r *http.Request) {
	serve.Serve403(w, r)
}

func (e *Enjin) Serve404(w http.ResponseWriter, r *http.Request) {
	serve.Serve404(w, r)
}

func (e *Enjin) Serve405(w http.ResponseWriter, r *http.Request) {
	serve.Serve405(w, r)
}

func (e *Enjin) Serve500(w http.ResponseWriter, r *http.Request) {
	serve.Serve500(w, r)
}

func (e *Enjin) ServeNotFound(w http.ResponseWriter, r *http.Request) {
	e.ServeStatusPage(404, w, r)
}

func (e *Enjin) ServeInternalServerError(w http.ResponseWriter, r *http.Request) {
	e.ServeStatusPage(500, w, r)
}

func (e *Enjin) ServeStatusPage(status int, w http.ResponseWriter, r *http.Request) {
	r = serve.SetServeStatus(status, r)
	reqLangTag := lang.GetTag(r)

	if path, ok := e.eb.statusPages[status]; ok {

		if pg := e.FindPage(reqLangTag, path); pg != nil {
			pg.Context().SetSpecific(argv.RequestArgvIgnoredKey, true)
			if err := e.ServePage(pg, w, r); err != nil {
				log.ErrorRDF(r, 1, "enjin error serving %v (found) page: %v - %v", status, path, err)
			} else {
				log.DebugRDF(r, 1, "enjin served %v (found) page: %v", status, path)
				return
			}
		}

		for _, spf := range e.eb.fServePathFeatures {
			if err := spf.ServePath(path, e, w, r); err == nil {
				log.DebugRDF(r, 1, "enjin served %v (%v) path: %v", status, spf.Tag(), path)
				return
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
		log.WarnRF(r, "unsupported status page: %v, serving 404 instead", status)
		e.ServeStatusPage(404, w, r)
	}
}

func (e *Enjin) ServePath(urlPath string, w http.ResponseWriter, r *http.Request) (err error) {

	for _, mw := range e.eb.fServePathFeatures {
		if err = mw.ServePath(urlPath, e, w, r); err == nil {
			// serve path middleware handled
			return
		}
	}

	pages := e.Pages()
	if p, ok := pages[urlPath]; ok {
		if err = e.ServePage(p, w, r); err == nil {
			// eb page found
			return
		}
	}

	return
}

func (e *Enjin) ServeJSON(v interface{}, w http.ResponseWriter, r *http.Request) (err error) {
	var data []byte
	if data, err = json.Marshal(v); err != nil {
		return
	}
	// log.DebugRDF(r, 1, "serving json: %v", string(data))
	e.ServeData(data, "application/json", w, r)
	return
}

func (e *Enjin) ServeStatusJSON(status int, v interface{}, w http.ResponseWriter, r *http.Request) (err error) {
	var data []byte
	if data, err = json.Marshal(v); err != nil {
		return
	}
	// log.DebugRDF(r, 1, "serving json: %v", string(data))
	e.ServeData(data, "application/json", w, serve.SetServeStatus(status, r))
	return
}

func (e *Enjin) ServePage(p feature.Page, w http.ResponseWriter, r *http.Request) (err error) {
	pUrl := p.Url()
	if pUrl != "" && pUrl[0] == '!' {
		err = fmt.Errorf("cannot serve not-path page: %v", pUrl)
		return
	} else if len(e.eb.fThemeRenderers) == 0 {
		err = fmt.Errorf("enjin has no theme renderers, cannot ServePage")
		return
	}

	if v, ok := r.Context().Value("userbase-denied-allow-error-page").(bool); ok && v {
		log.DebugRF(r, "bypassing all user access controls to show an error page of some sort: %v", pUrl)
	} else {

		// looking for view_<origin>_page actions... if any found, perform access
		check := feature.NewAction(p.PageMatter().Origin, "view", "page")
		if e.FindAllUserActions().Has(check) {

			// found this page's particular view action
			// this requires that the user must have a group with this action
			log.DebugRF(r, "found access control page action: %v", check)

			if user := userbase.GetCurrentUser(r); user != nil {
				if !user.Can(check) {
					log.WarnF("denying user %v access (%v) to: %v", user.EID, check, pUrl)
					r = r.Clone(context.WithValue(r.Context(), "userbase-denied-allow-error-page", true))
					e.ServeNotFound(w, r)
					return
				} else {
					log.DebugRF(r, "authenticated user allowed to: (%v) %v", check, pUrl)
				}
			} else if e.eb.publicUser.Has(check) {
				log.DebugRF(r, "public user allowed to: (%v) %v", check, pUrl)
			} else {
				log.ErrorRF(r, "denying access, authenticated user not found and public has no access: (%v) %v", check, pUrl)
				r = r.Clone(context.WithValue(r.Context(), "userbase-denied-allow-error-page", true))
				e.ServeNotFound(w, r)
				return
			}

		} else {
			log.ErrorRF(r, `denying access, page action not found: "%v"`, check)
			r = r.Clone(context.WithValue(r.Context(), "userbase-denied-allow-error-page", true))
			e.ServeNotFound(w, r)
			return
		}

	}

	for _, ptp := range e.eb.fPageTypeProcessors {
		var pg feature.Page
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

	var t feature.Theme
	if t, err = e.GetTheme(); err != nil {
		return
	} else if e.eb.hotReload {
		log.DebugF("hot-reloading theme: %v", t.Name())
		if err = t.Reload(); err != nil {
			err = fmt.Errorf("error reloading theme: %v", err)
			return
		}
	}

	ctx := e.Context()

	reqLangTag := lang.GetTag(r)
	ctx.SetSpecific("ReqLangTag", reqLangTag)
	ctx.SetSpecific("Request", map[string]string{
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
	ctx.SetSpecific("RequestContext", r.Context())

	tConfig := t.GetConfig()
	var pccs *csp.PageContextContentSecurity
	pccs, r = e.contentSecurityPolicy.PreparePageContext(tConfig.ContentSecurityPolicy, ctx, r)
	ctx.SetSpecific("RequestPolicy", map[string]interface{}{
		"Permissions":     e.permissionsPolicy.GetRequestPolicy(r),
		"ContentSecurity": pccs,
	})

	ctx.SetSpecific("BaseUrl", net.BaseURL(r))
	ctx.SetSpecific(lang.PrinterKey, lang.GetPrinterFromRequest(r))
	ctx.SetSpecific(string(argv.RequestArgvKey), argv.GetRequestArgv(r))

	parsedTag := e.eb.defaultLang
	if v := ctx.Get("Language"); v != nil {
		if pageLang, ok := v.(string); ok {
			if pageLang != "" {
				if tag, ee := language.Parse(pageLang); ee == nil {
					parsedTag = tag
				} else {
					log.ErrorRF(r, "invalid language tag: %v - %v", pageLang, ee)
				}
			}
		} else {
			log.ErrorRF(r, "page language tag not a string: %T %+v", v, v)
		}
	}
	ctx.SetSpecific("Language", parsedTag.String())
	ctx.SetSpecific("LanguageTag", parsedTag)

	fpcPgCtx := p.Context().Copy()
	fpcPgCtx.SetSpecific("Content", p.Content())
	for _, pcm := range e.eb.fPageContextModifiers {
		log.TraceRF(r, "filtering page context with: %v", pcm.Tag())
		ctx = pcm.FilterPageContext(ctx, fpcPgCtx, r)
	}

	for _, prh := range e.eb.fPageRestrictionHandlers {
		log.TraceRF(r, "checking restricted pages with: %v", prh.Tag())
		var ok bool
		if ctx, r, ok = prh.RestrictServePage(ctx, w, r); !ok {
			addr, _ := net.GetIpFromRequest(r)
			log.WarnRF(r, "%v feature denied %v access to: %v", prh.Tag(), addr, r.URL.Path)
			e.ServeNotFound(w, r)
			return
		}
	}

	allMenus := make(map[string]interface{})
	for _, mp := range e.eb.fMenuProviders {
		for name, m := range mp.GetMenus(reqLangTag) {
			camel := strcase.ToCamel(name)
			allMenus[camel] = m
			log.TraceRF(r, "providing [%v] menu: %v (.SiteMenu.%v)", reqLangTag.String(), name, camel)
		}
	}

	if len(allMenus) > 0 {
		ctx.SetSpecific("SiteMenu", allMenus)
	}

	var data []byte
	var redirect string

	renderer := e.GetThemeRenderer(ctx)

	if data, redirect, err = renderer.RenderPage(ctx, p); err != nil {
		log.ErrorRF(r, "error rendering page: %v - %v", pUrl, err)
		return
	} else if redirect != "" {
		log.DebugRF(r, "redirecting from RenderPage: %v - %v", pUrl, redirect)
		e.ServeRedirect(redirect, w, r)
		return
	}
	if cacheControl := p.Context().String("CacheControl", ""); cacheControl != "" {
		r = serve.SetCacheControl(cacheControl, w, r)
	}
	mime := ctx.String("ContentType", "text/html; charset=utf-8")
	contentDisposition := ctx.String("ContentDisposition", "inline")
	r = r.Clone(context.WithValue(r.Context(), "Content-Disposition", contentDisposition))
	e.permissionsPolicy.FinalizeRequest(w, r)
	e.contentSecurityPolicy.FinalizeRequest(w, r)
	e.ServeData(data, mime, w, r)
	return
}

func (e *Enjin) ServeData(data []byte, mime string, w http.ResponseWriter, r *http.Request) {

	for _, prh := range e.eb.fDataRestrictionHandlers {
		// log.TraceRF(r, "checking restricted data with: %v", f.Tag())
		var ok bool
		if r, ok = prh.RestrictServeData(data, mime, w, r); !ok {
			addr, _ := net.GetIpFromRequest(r)
			log.WarnRF(r, "[restricted] permission denied %v for: %v", addr, r.URL.Path)
			// e.ServeBasic401(w, r)
			return
		}
	}

	w.Header().Set("Content-Type", mime)
	if reqArgv := argv.GetRequestArgv(r); len(reqArgv.Argv) > 0 {
		w.Header().Set("Cache-Control", "no-store")
	} else if value := serve.GetCacheControl(r); value != "" {
		w.Header().Set("Cache-Control", value)
	}
	if value, ok := r.Context().Value("Content-Disposition").(string); ok {
		w.Header().Set("Content-Disposition", value)
	} else {
		w.Header().Set("Content-Disposition", "inline")
	}

	status := serve.GetServeStatus(r)
	if !serve.StatusHasBody(status) {
		w.WriteHeader(status)
		return
	}

	// only one translation allowed, non-feature translators take precedence
	basicMime := beStrings.GetBasicMime(mime)
	if fn, ok := e.eb.translators[basicMime]; ok {
		data, mime = fn(data)
		basicMime = beStrings.GetBasicMime(mime)
	} else {
		for _, ot := range e.eb.fOutputTranslators {
			// log.TraceRF(r, "checking output filter: %v", f.Tag())
			if ot.CanTranslate(mime) {
				if d, m, err := ot.TranslateOutput(e, data, mime); err == nil {
					data, mime = d, m
					basicMime = beStrings.GetBasicMime(mime)
					log.DebugRF(r, "filtered output: %v - %v", ot.Tag(), mime)
				} else {
					log.DebugRF(r, "%v error filtering output: %v", ot.Tag(), err)
				}
				break
			}
		}
	}

	for _, ot := range e.eb.fOutputTransformers {
		if ot.CanTransform(mime, r) {
			data = ot.TransformOutput(mime, data)
		}
	}

	// non-feature transformers happen after features
	if fn, ok := e.eb.transformers[mime]; ok {
		data = fn(data)
	} else if fn, ok := e.eb.transformers[basicMime]; ok {
		data = fn(data)
	}

	w.WriteHeader(status)
	_, _ = w.Write(data)
}