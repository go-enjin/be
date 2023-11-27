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

package email_backup

import (
	"net/http"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/request"
)

func (f *CFeature) ServeManagePage(settingsPath string, saf feature.SiteAuthFeature, w http.ResponseWriter, r *http.Request) (handled bool, redirect string) {

	var err error

	if r.Method == http.MethodPost {

		if nonce := request.SafeQueryFormValue(r, ManageNonceName); nonce != "" {
			if f.Enjin.VerifyNonce(ManageNonceKey, nonce) {
				switch r.FormValue("submit") {
				case "create":
					f.SiteUserSetupStageHandler(saf, w, r)
					handled = true
					return
				case "revoke":
					handled, redirect = f.ServeRevokePage(settingsPath, saf, w, r)
					return
				}
			}
		} else {
			switch {
			case r.Form.Has(SetupNonceName):
				f.SiteUserSetupStageHandler(saf, w, r)
				handled = true
				return
			case r.Form.Has(RevokeNonceName):
				handled, redirect = f.ServeRevokePage(settingsPath, saf, w, r)
				return
			}
		}

	}

	ctx := beContext.Context{
		"FeatureInfo": f.SiteFeatureInfo(r),
		"Provider":    f.Tag().Kebab(),
		"FormAction":  r.URL.Path,
		"Nonces": feature.Nonces{
			{Name: ManageNonceName, Key: ManageNonceKey},
		},
	}

	t := f.Site().SiteTheme()
	if err = f.Site().PrepareAndServePage("site-auth", "email-backup--manage", r.URL.Path, t, w, r, ctx); err != nil {
		log.ErrorRF(r, "error preparing and serving totp--manage page: %v", err)
		panic(err)
	}

	handled = true
	return
}
