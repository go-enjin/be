// Copyright (c) 2023  The Go-Enjin Authors
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

package auth

import (
	"net/http"

	"github.com/go-enjin/be/pkg/feature"
	clPath "github.com/go-corelibs/path"
	"github.com/go-enjin/be/pkg/request"
)

func (f *CFeature) SiteSettingsPanel(settingsPath string) (serve, handle http.HandlerFunc) {
	// find all auth providers and mfa features which support serving and/or handling settings panel pages
	// if none, return nil, else provide top-level handler funcs

	var serveOrder, handleOrder []string
	serveLookup := make(map[string]http.HandlerFunc)
	handleLookup := make(map[string]http.HandlerFunc)

	for _, saf := range append(f.sap.Features.AsFeatures(), f.mfa.Features.AsFeatures()...) {
		if sasp, ok := saf.This().(feature.SiteAuthSettingsPanel); ok {
			tag := saf.Tag().Kebab()
			path := settingsPath + "/" + tag
			if s, h := sasp.SiteAuthSettingsPanel(path, f); s != nil || h != nil {
				if s != nil {
					serveLookup[tag] = s
					serveOrder = append(serveOrder, tag)
				}
				if h != nil {
					handleLookup[tag] = h
					handleOrder = append(handleOrder, tag)
				}
			}
		}
	}

	if len(serveLookup) > 0 {
		serve = f.MakeServeSiteSettingsPanel(settingsPath, serveOrder, serveLookup)
	}
	if len(handleLookup) > 0 {
		handle = f.MakeHandleSiteSettingsPanel(settingsPath, handleOrder, handleLookup)
	}
	return
}

func (f *CFeature) MakeServeSiteSettingsPanel(settingsPath string, order []string, lookup map[string]http.HandlerFunc) (serve http.HandlerFunc) {
	return func(w http.ResponseWriter, r *http.Request) {

		if allowed := f.Site().RequireVerification(settingsPath, w, r); !allowed {
			return
		}

		if clPath.MatchExact(r.URL.Path, settingsPath) {
			// serve page for user to select a specific auth settings panel
			f.ServeSettingsPanelSelectorPage(settingsPath, w, r)
			return
		}

		if suffix, match := clPath.MatchCut(r.URL.Path, settingsPath); match {
			for _, prefix := range order {
				if clPath.MatchExact(suffix, prefix) {
					if h, ok := lookup[prefix]; ok {
						h(w, r)
						return
					}
				}
			}
		}

		f.Enjin.ServeRedirect(settingsPath, w, r)
		return
	}
}

func (f *CFeature) MakeHandleSiteSettingsPanel(settingsPath string, order []string, lookup map[string]http.HandlerFunc) (handle http.HandlerFunc) {
	return func(w http.ResponseWriter, r *http.Request) {

		if allowed := f.Site().RequireVerification(settingsPath, w, r); !allowed {
			return
		}

		if clPath.MatchExact(r.URL.Path, settingsPath) {
			// serve page for user to select a specific auth settings panel
			_ = r.ParseForm()
			if r.Method == http.MethodPost && r.Form.Has(SettingsNonceName) && request.SafeQueryFormValue(r, "submit") != "cancel" {
				f.ServeSettingsPanelSelectorPage(settingsPath, w, r)
				return
			}
			f.Enjin.ServeRedirect(settingsPath, w, r)
			return
		}

		if suffix, match := clPath.MatchCut(r.URL.Path, settingsPath); match {
			for _, prefix := range order {
				if clPath.MatchExact(suffix, prefix) {
					if h, ok := lookup[prefix]; ok {
						h(w, r)
						return
					}
				}
			}
		}

		f.Enjin.ServeRedirect(settingsPath, w, r)
	}
}
