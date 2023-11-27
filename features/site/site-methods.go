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

package site

import (
	"net/http"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/menu"
	"github.com/go-enjin/be/pkg/userbase"
)

func (f *CFeature) SitePath() (path string) {
	path = f.sitePath
	return
}

func (f *CFeature) SiteTheme() (t feature.Theme) {
	return f.theme
}

func (f *CFeature) SiteMenu(r *http.Request) (siteMenu context.Context) {
	mainMenu := menu.Menu{}
	if permissions := userbase.GetCurrentPermissions(r); permissions != nil {
		for _, ef := range f.siteFeatures.Features {
			if permissions.Has(ef.Action("access", "feature")) {
				mainMenu = append(mainMenu, ef.SiteFeatureMenu(r)...)
			}
		}
	}
	return context.Context{
		"MainMenu": mainMenu,
	}
}

func (f *CFeature) SiteUsers() (sup feature.SiteUsersProvider) {
	sup = f.siteUsersProvider
	return
}

func (f *CFeature) SiteAuth() (sup feature.SiteAuthFeature) {
	sup = f.authFeature
	return
}

func (f *CFeature) SiteFeatures() (list feature.SiteFeatures) {
	list = feature.FilterTyped[feature.SiteFeature](f.siteFeatures.Features.AsFeatures())
	return
}

func (f *CFeature) RequireVerification(verifyPath string, w http.ResponseWriter, r *http.Request) (allowed bool) {
	// if the site has no auth or factors ready, allow the request to proceed as-if it had been verified
	// otherwise, mfa challenge wall
	if allowed = f.authFeature == nil || f.authFeature.NumFactorsPresent() == 0; allowed {
	} else if handled, redirect := f.authFeature.RequireVerification(verifyPath, w, r); handled {
	} else if allowed = redirect == ""; !allowed {
		f.Enjin.ServeRedirect(redirect, w, r)
	}
	return
}

func (f *CFeature) MustRequireVerification(verifyPath string, w http.ResponseWriter, r *http.Request) (allowed bool) {
	// serves 404 if there are no auth or factors ready, deny the request to proceed, otherwise, mfa challenge wall
	if f.authFeature == nil || f.authFeature.NumFactorsPresent() == 0 {
		log.DebugRF(r, "no SiteAuthFeature present, must require verification serving not found: %v", verifyPath)
		f.Enjin.ServeNotFound(w, r)
	} else if handled, redirect := f.authFeature.RequireVerification(verifyPath, w, r); handled {
	} else if allowed = redirect == ""; !allowed {
		f.Enjin.ServeRedirect(redirect, w, r)
	}
	return
}
