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

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net"
	"github.com/go-enjin/be/pkg/net/serve"
	"github.com/go-enjin/be/pkg/request"
	"github.com/go-enjin/be/pkg/request/argv"
	"github.com/go-enjin/be/pkg/signals"
	beStrings "github.com/go-enjin/be/pkg/strings"
	"github.com/go-enjin/be/pkg/userbase"
)

func (e *Enjin) FinalizeServeRequest(w http.ResponseWriter, r *http.Request) (modified *http.Request) {
	for _, fsp := range e.GetFinalizeServePagesFeatures() {
		if m := fsp.FinalizeServeRequest(w, r); m != nil {
			r = m
		}
	}
}

func (e *Enjin) ServeRedirect(destination string, w http.ResponseWriter, r *http.Request) {
	r = e.FinalizeServeRequest(w, r)
	http.Redirect(w, r, destination, http.StatusSeeOther)
	e.Emit(signals.ServedRedirect, feature.EnjinTag.String(), interface{}(e).(feature.Internals), r)
}

func (e *Enjin) ServeRedirectHomePath(w http.ResponseWriter, r *http.Request) {
	e.ServeRedirect(request.GetHomePath(r), w, r)
}

func (e *Enjin) Serve204(w http.ResponseWriter, r *http.Request) {
	r = e.FinalizeServeRequest(w, r)
	serve.Serve204(w, r)
	e.Emit(signals.Served204, feature.EnjinTag.String(), interface{}(e).(feature.Internals), r)
}

func (e *Enjin) Serve400(w http.ResponseWriter, r *http.Request) {
	r = e.FinalizeServeRequest(w, r)
	serve.Serve400(w, r)
	e.Emit(signals.Served400, feature.EnjinTag.String(), interface{}(e).(feature.Internals), r)
}

func (e *Enjin) Serve401(w http.ResponseWriter, r *http.Request) {
	r = e.FinalizeServeRequest(w, r)
	serve.Serve401(w, r)
	e.Emit(signals.Served401, feature.EnjinTag.String(), interface{}(e).(feature.Internals), r)
}

func (e *Enjin) ServeBasic401(w http.ResponseWriter, r *http.Request) {
	r = e.FinalizeServeRequest(w, r)
	serve.ServeBasic401(w, r)
	e.Emit(signals.ServedBasic401, feature.EnjinTag.String(), interface{}(e).(feature.Internals), r)
}

func (e *Enjin) Serve403(w http.ResponseWriter, r *http.Request) {
	r = e.FinalizeServeRequest(w, r)
	serve.Serve403(w, r)
	e.Emit(signals.Served403, feature.EnjinTag.String(), interface{}(e).(feature.Internals), r)
}

func (e *Enjin) Serve404(w http.ResponseWriter, r *http.Request) {
	r = e.FinalizeServeRequest(w, r)
	serve.Serve404(w, r)
	e.Emit(signals.Served404, feature.EnjinTag.String(), interface{}(e).(feature.Internals), r)
}

func (e *Enjin) Serve405(w http.ResponseWriter, r *http.Request) {
	r = e.FinalizeServeRequest(w, r)
	serve.Serve405(w, r)
	e.Emit(signals.Served405, feature.EnjinTag.String(), interface{}(e).(feature.Internals), r)
}

func (e *Enjin) Serve500(w http.ResponseWriter, r *http.Request) {
	r = e.FinalizeServeRequest(w, r)
	serve.Serve500(w, r)
	e.Emit(signals.Served500, feature.EnjinTag.String(), interface{}(e).(feature.Internals), r)
}

func (e *Enjin) ServeNotFound(w http.ResponseWriter, r *http.Request) {
	e.ServeStatusPage(404, w, r)
}

func (e *Enjin) ServeForbidden(w http.ResponseWriter, r *http.Request) {
	e.ServeStatusPage(403, w, r)
}

func (e *Enjin) ServeInternalServerError(w http.ResponseWriter, r *http.Request) {
	e.ServeStatusPage(500, w, r)
}

func (e *Enjin) ServeStatusPage(status int, w http.ResponseWriter, r *http.Request) {
	e.Emit(signals.PreServeStatusPage, feature.EnjinTag.String(), interface{}(e).(feature.Internals), status, r)

	r = serve.SetServeStatus(status, r)
	reqLangTag := lang.GetTag(r)

	if path, ok := e.eb.statusPages[status]; ok {
		if pg := e.FindPage(r, reqLangTag, path); pg != nil {
			pg.Context().SetSpecific(argv.RequestIgnoredKey, true)
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

	e.Emit(signals.PostServeStatusPage, feature.EnjinTag.String(), interface{}(e).(feature.Internals), status, r)
}

func (e *Enjin) ServePath(urlPath string, w http.ResponseWriter, r *http.Request) (err error) {

	for _, mw := range e.eb.fServePathFeatures {
		if err = mw.ServePath(urlPath, e, w, r); err == nil {
			// serve path middleware handled
			e.Emit(signals.ServedPath, feature.EnjinTag.String(), interface{}(e).(feature.Internals), urlPath, r)
			return
		}
	}

	pages := e.Pages()
	if p, ok := pages[urlPath]; ok {
		if err = e.ServePage(p, w, r); err == nil {
			// eb page found
			e.Emit(signals.ServedPath, feature.EnjinTag.String(), interface{}(e).(feature.Internals), urlPath, r)
			e.Emit(signals.ServedPathPage, feature.EnjinTag.String(), interface{}(e).(feature.Internals), urlPath, p, r)
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

	e.Emit(signals.PreServePage, feature.EnjinTag.String(), interface{}(e).(feature.Internals), p, r)

	if pUrl := p.Url(); pUrl != "" && pUrl[0] == '!' {
		err = fmt.Errorf("cannot serve not-path page: %v", pUrl)
		return
	} else if len(e.eb.fThemeRenderers) == 0 {
		err = fmt.Errorf("enjin has no theme renderers, cannot ServePage")
		return
	} else if e.eb.fServePagesHandler == nil {
		err = fmt.Errorf("enjin has no page serve handlers, cannot ServePage")
		return
	}

	pUrl := p.Url()

	if v, ok := r.Context().Value(gDenyUserAndAllowErrorPage).(bool); ok && v {
		log.DebugRF(r, "bypassing all user access controls to show an error page of some sort: %v", pUrl)
	} else {

		// looking for view_<origin>_page actions... if any found, perform access
		check := feature.NewAction(p.PageMatter().Origin, "view", "page")
		if e.FindAllUserActions().Has(check) {

			// found this page's particular view action
			// this requires that the user must have a group with this action
			log.DebugRF(r, "found access control page action: %v", check)

			eid := userbase.GetCurrentEID(r)
			if !userbase.CurrentUserCan(r, check) {
				log.WarnF("denying user %q access (%v) to: %v", eid, check, pUrl)
				r = r.Clone(context.WithValue(r.Context(), gDenyUserAndAllowErrorPage, true))
				e.ServeNotFound(w, r)
				return
			}
			log.DebugRF(r, "allowing user %q access (%v) to: %v", eid, check, pUrl)

		} else {
			log.ErrorRF(r, `denying access, page action not found: "%v"`, check)
			r = r.Clone(context.WithValue(r.Context(), gDenyUserAndAllowErrorPage, true))
			e.ServeNotFound(w, r)
			return
		}

	}

	err = e.eb.fServePagesHandler.ServePage(p, e.MustGetTheme(), e.Context(r), w, r)
	e.Emit(signals.PostServePage, feature.EnjinTag.String(), interface{}(e).(feature.Internals), p, r)
	return
}

func (e *Enjin) ServeData(data []byte, mime string, w http.ResponseWriter, r *http.Request) {
	e.Emit(signals.PreServeData, feature.EnjinTag.String(), interface{}(e).(feature.Internals), &data, mime, r)

	r = e.FinalizeServeRequest(w, r)

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
	if reqArgv := argv.Get(r); len(reqArgv.Argv) > 0 {
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
	e.Emit(signals.PostServeData, feature.EnjinTag.String(), interface{}(e).(feature.Internals), &data, mime, status, r)
}