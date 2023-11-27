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

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
)

func (f *CFeature) ServeSettingsPanelSetupSelectorPage(settingsPath string, w http.ResponseWriter, r *http.Request) {

	var claims *feature.CSiteAuthClaims
	if claims = f.getPrivateClaims(r); claims == nil {
		f.Enjin.ServeNotFound(w, r)
		return
	}

	var err error
	ctx := beContext.Context{
		"FeatureInfo": f.SiteFeatureInfo(r),
	}

	var order []string
	paths := make(map[string]string)
	infos := make(map[string]*feature.CSiteFeatureInfo)

	for _, mfp := range f.mfa.Features {
		if sasp, ok := mfp.This().(feature.SiteAuthSettingsPanel); ok {
			if mfp.IsMultiFactorBackup() {
				continue
			}
			kebab := mfp.Tag().Kebab()
			path := settingsPath + "/" + kebab
			infos[kebab] = sasp.SiteFeatureInfo(r)
			if s, _ := sasp.SiteAuthSettingsPanel(path, f); s != nil {
				order = append(order, kebab)
				paths[kebab] = path
			}
		}
	}

	ctx.SetSpecific("PanelsOrder", order)
	ctx.SetSpecific("PanelsPaths", paths)
	ctx.SetSpecific("PanelsInfos", infos)

	t := f.Site().SiteTheme()
	if err = f.Site().PrepareAndServePage("site-auth", "settings--setup-selector", r.URL.Path, t, w, r, ctx); err != nil {
		log.ErrorRF(r, "error preparing and serving settings--selector page: %v", err)
		panic(err)
	}

}
