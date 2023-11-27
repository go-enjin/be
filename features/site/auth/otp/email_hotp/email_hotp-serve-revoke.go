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

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/request"
)

func (f *CFeature) ServeRevokePage(settingsPath string, saf feature.SiteAuthFeature, w http.ResponseWriter, r *http.Request) (handled bool, redirect string) {

	var err error
	printer := lang.GetPrinterFromRequest(r)

	if submit := request.SafeQueryFormValue(r, "submit"); submit == "cancel" {
		// cancel is just resetting form values with a reload
		f.Enjin.ServeRedirect(r.URL.Path, w, r)
		return
	}

	ctx := beContext.Context{
		"FeatureInfo": f.SiteFeatureInfo(r),
		"FormAction":  settingsPath,
		"Nonces": feature.Nonces{
			{Name: RevokeNonceName, Key: RevokeNonceKey},
		},
	}

	var provision string

	if r.Method == http.MethodPost {

		if r.URL.Path == settingsPath {

			if provision = r.FormValue("provision"); provision != "" {

				if f.hasSecureProvision(provision, r) {

					switch r.FormValue("submit") {
					case "revoke":
						ctx.SetSpecific("RevokeConfirmation", true)
					case "revoke--confirmation":
						ctx.SetSpecific("RevokeConfirmed", true)
					case "revoke--confirmed":

						if nonce := request.SafeQueryFormValue(r, RevokeNonceName); nonce != "" {
							if !f.Enjin.VerifyNonce(RevokeNonceKey, nonce) {
								r = feature.AddErrorNotice(r, true, errors.FormExpiredError(printer))
							} else {

								if err = f.revokeSecureProvision(provision, r); err != nil {
									panic(err)
								}

								f.Enjin.ServeRedirect(settingsPath, w, r)
								return

							}
						}
					}

				}

			}

		}

	}

	names := f.listSecureProvisions(r)
	ctx.SetSpecific("ProvisionLabels", names)
	ctx.SetSpecific("SelectedProvision", provision)

	t := f.Site().SiteTheme()
	if err = f.Site().PrepareAndServePage("site-auth", "email-hotp--revoke", r.URL.Path, t, w, r, ctx); err != nil {
		log.ErrorRF(r, "error preparing and serving email-hotp--revoke page: %v", err)
		panic(err)
	}
	handled = true
	return
}
