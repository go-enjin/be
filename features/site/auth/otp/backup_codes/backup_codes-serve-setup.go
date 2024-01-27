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

package backup_codes

import (
	"net/http"

	"github.com/go-corelibs/slices"
	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/crypto"
	berrs "github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/hash/sha"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/request"
	"github.com/go-enjin/be/pkg/strings"
)

func (f *CFeature) SiteUserSetupStageReady(eid string, r *http.Request) (ready bool) {
	ready = f.countSecureProvisions(r) > 0
	return
}

func (f *CFeature) SiteUserSetupStageHandler(saf feature.SiteAuthFeature, w http.ResponseWriter, r *http.Request) {
	if handled := f.ProcessSetupPage(saf, w, r); !handled {
		f.Enjin.ServeRedirect(r.URL.Path, w, r)
	}
	return
}

func (f *CFeature) ProcessSetupPage(saf feature.SiteAuthFeature, w http.ResponseWriter, r *http.Request) (handled bool) {

	var err error
	var backupCodes []string
	var provision string

	if submit := request.SafeQueryFormValue(r, "submit"); submit == "cancel" {
		// cancel is just resetting form values with a reload because we don't really know where "back" really is
		f.Enjin.ServeRedirect(r.URL.Path, w, r)
		return
	}

	if provision = r.FormValue("provision"); provision == "" {
		info := f.SiteMultiFactorInfo(r)
		provision = info.Label
	}

	names := f.listSecureProvisions(r)
	for slices.Within(provision, names) {
		provision = strings.IncrementFileName(provision)
	}

	printer := lang.GetPrinterFromRequest(r)

	if r.Method == http.MethodPost {

		if nonce := r.FormValue(SetupNonceName); nonce != "" {
			if !f.Enjin.VerifyNonce(SetupNonceKey, nonce) {
				_ = f.removeNewSecretKey(r) // don't re-use backup codes before setup!
				r = feature.AddErrorNotice(r, true, berrs.FormExpiredError(printer))

			} else {

				submit := r.FormValue("submit")

				switch submit {

				case "setup-confirm":
					// let the user view the backup codes
					for i := 0; i < 10; i++ {
						rv, ee := crypto.RandomValue(32)
						berrs.Must(ee)
						shasum := sha.MustDataHash10(rv)
						backupCodes = append(backupCodes, shasum)
					}
					berrs.Must(f.setNewSecretKey(backupCodes, r))

				case "setup":

					if codes := f.getNewSecretKey(r); len(codes) == 10 {
						berrs.Must(f.removeNewSecretKey(r))
						berrs.Must(f.setSecureProvision(provision, codes, nil, r))
						// finalize and reload same page
						return
					}

				}

			}
		}

	}

	handled = true

	ctx := context.Context{
		"FeatureInfo": f.SiteFeatureInfo(r),
		"FormAction":  r.URL.Path,
		"Nonces": feature.Nonces{
			{Name: SetupNonceName, Key: SetupNonceKey},
		},
		"ProvisionLabel": provision,
		"BackupCodes":    backupCodes,
	}

	t := f.Site().SiteTheme()
	if err = f.Site().PrepareAndServePage("site-auth", "backup-codes--setup", r.URL.Path, t, w, r, ctx); err != nil {
		log.ErrorRF(r, "error preparing and serving totp--setup page: %v", err)
	}
	return
}
