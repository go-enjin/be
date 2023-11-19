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

package email_hotp

import (
	"net/http"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	bePath "github.com/go-enjin/be/pkg/path"
)

func (f *CFeature) SiteAuthSettingsPanel(settingsPath string, saf feature.SiteAuthFeature) (serve, handle http.HandlerFunc) {
	// settingsPath is the path to this feature's settings panel
	serve = f.MakeServeSiteSettingsPanel(settingsPath, saf)
	handle = f.MakeHandleSiteSettingsPanel(settingsPath, saf)
	return
}

func (f *CFeature) MakeServeSiteSettingsPanel(settingsPath string, saf feature.SiteAuthFeature) (serve http.HandlerFunc) {
	return func(w http.ResponseWriter, r *http.Request) {

		if _, ok := bePath.MatchCut(r.URL.Path, settingsPath); ok {
			if allowed := f.Site().RequireVerification(settingsPath, w, r); !allowed {
				return
			}

			if factors := f.listSecureProvisions(r); len(factors) > 0 {
				f.ServeManagePage(settingsPath, saf, w, r)
				return
			}

			f.SiteUserSetupStageHandler(saf, w, r)
			return
		}

		log.ErrorRF(r, "bad routing, email-hotp serve-settings panel handler received %q request for: %q", r.Method, r.URL.Path)
		f.Enjin.ServeInternalServerError(w, r)
	}
}

func (f *CFeature) MakeHandleSiteSettingsPanel(settingsPath string, saf feature.SiteAuthFeature) (serve http.HandlerFunc) {
	return func(w http.ResponseWriter, r *http.Request) {

		if _, ok := bePath.MatchCut(r.URL.Path, settingsPath); ok {
			if allowed := f.Site().RequireVerification(settingsPath, w, r); !allowed {
				return
			}
			if factors := f.listSecureProvisions(r); len(factors) > 0 {
				f.ServeManagePage(settingsPath, saf, w, r)
				return
			}
			f.SiteUserSetupStageHandler(saf, w, r)
			return
		}

		log.ErrorRF(r, "bad routing, email-hotp handle-settings panel handler received %q request for: %q", r.Method, r.URL.Path)
		f.Enjin.ServeInternalServerError(w, r)
	}
}