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

	"github.com/go-corelibs/x-text/message"
	beContext "github.com/go-enjin/be/pkg/context"
	berrs "github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/request"
	"github.com/go-enjin/be/pkg/userbase"
)

func (f *CFeature) ServeSettingsPanelSelectorPage(settingsPath string, w http.ResponseWriter, r *http.Request) {

	var claims *feature.CSiteAuthClaims
	if claims = f.getPrivateClaims(r); claims == nil {
		f.Enjin.ServeNotFound(w, r)
		return
	}

	type deleteOwnUser struct {
		Confirmed    bool
		Confirmation string
	}

	su := f.Site().SiteUsers()
	au := userbase.GetCurrentUser(r)
	printer := message.GetPrinter(r)
	var deleteRequested *deleteOwnUser

	if r.Method == http.MethodPost {
		if nonce := request.SafeQueryFormValue(r, SettingsNonceName); nonce != "" {
			if f.Enjin.VerifyNonce(SettingsNonceKey, nonce) {

				switch request.SafeQueryFormValue(r, "submit") {

				case "reactivate":
					if err := su.SetUserActive(r, claims.EID, true); err != nil {
						panic(err)
					} else {
						msg := printer.Sprintf("Your primary sign-in method has been reactivated.")
						f.Site().PushImportantNotice(claims.EID, true, msg)
						f.Enjin.ServeRedirect(r.URL.Path, w, r)
						return
					}

				case "deactivate":
					if err := su.SetUserActive(r, claims.EID, false); err != nil {
						panic(err)
					} else {
						msg := printer.Sprintf("Your primary sign-in method has been deactivated. You can still sign-in with a backup method.")
						f.Site().PushWarnNotice(claims.EID, true, msg)
						f.Enjin.ServeRedirect(r.URL.Path, w, r)
						return
					}

				case "delete":
					deleteRequested = &deleteOwnUser{}

				case "delete-confirm":
					deleteRequested = &deleteOwnUser{
						Confirmed:    true,
						Confirmation: DefaultDeleteOwnUserConfirmation,
					}

				case "delete-confirmed":

					if confirmation := request.SafeQueryFormValue(r, "delete-confirmation"); confirmation == DefaultDeleteOwnUserConfirmation {
						log.InfoRF(r, "user correctly requested account deletion: %v - %v", au.GetEID(), au.GetEmail())
						if err := su.DeleteUser(r, au.GetEID()); err == nil {

							log.InfoRF(r, "user account deleted: %v", au.GetEID())
							if m := f.resetCurrentUser(w, r); m != nil {
								r = m
							}

							f.Enjin.ServeRedirect("/", w, r)
							return

						} else {
							log.ErrorRF(r, "error deleting user: %v - %v", au.GetEID(), err)
							r = feature.AddErrorNotice(r, true, berrs.UnexpectedError(printer))
						}

					} else {

						msg := printer.Sprintf("Your account was not deleted because you did not enter the correct confirmation phrase!")
						f.Site().PushErrorNotice(claims.EID, true, msg)
						deleteRequested = &deleteOwnUser{
							Confirmed:    true,
							Confirmation: DefaultDeleteOwnUserConfirmation,
						}
					}

				}

			}
		}
	}

	var err error
	ctx := beContext.Context{
		"FeatureInfo": f.SiteFeatureInfo(r),
		"FormAction":  settingsPath,
		"Nonces": feature.Nonces{
			{Name: SettingsNonceName, Key: SettingsNonceKey},
		},
	}

	ctx.SetSpecific("UserActive", au.GetActive())
	ctx.SetSpecific("CanDeleteOwnUser", au.Can(feature.NewAction(su.Tag().Kebab(), "delete-own", "user")))

	if deleteRequested != nil {
		ctx.SetSpecific("DeleteRequested", deleteRequested)
	}

	var order []string
	paths := make(map[string]string)
	infos := make(map[string]*feature.CSiteFeatureInfo)

	for _, saf := range append(f.sap.Features.AsFeatures(), f.mfa.Features.AsFeatures()...) {
		if sasp, ok := saf.This().(feature.SiteAuthSettingsPanel); ok {
			kebab := saf.Tag().Kebab()
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
	if err = f.Site().PrepareAndServePage("site-auth", "settings--selector", r.URL.Path, t, w, r, ctx); err != nil {
		log.ErrorRF(r, "error preparing and serving settings--selector page: %v", err)
		panic(err)
	}

}
