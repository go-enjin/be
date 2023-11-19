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

package app_totp

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/xlzd/gotp"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/request"
	"github.com/go-enjin/be/pkg/slices"
	"github.com/go-enjin/be/pkg/strings"
	"github.com/go-enjin/be/pkg/userbase"
)

func (f *CFeature) SiteUserSetupStageReady(eid string, r *http.Request) (ready bool) {
	ready = f.countSecureProvisions(r) > 0
	return
}

func (f *CFeature) SiteUserSetupStageHandler(saf feature.SiteAuthFeature, w http.ResponseWriter, r *http.Request) {
	if !f.ProcessSetupPage(saf, w, r) {
		f.Enjin.ServeRedirect(r.URL.Path, w, r)
	}
	return
}

func (f *CFeature) ProcessSetupPage(saf feature.SiteAuthFeature, w http.ResponseWriter, r *http.Request) (handled bool) {

	var err error
	var provision, secret string
	printer := lang.GetPrinterFromRequest(r)

	if submit := request.SafeQueryFormValue(r, "submit"); submit == "cancel" {
		// cancel is just resetting form values with a reload
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

	if r.Method == http.MethodPost {

		if r.FormValue("submit") == "setup" {

			if secret = f.getNewSecretKey(r); secret != "" {

				if nonce := request.SafeQueryFormValue(r, SetupNonceName); nonce != "" {
					if !f.Enjin.VerifyNonce(SetupNonceKey, nonce) {
						r = feature.AddErrorNotice(r, true, errors.FormExpiredError(printer))
					} else {

						totp := gotp.NewDefaultTOTP(secret)
						challenge := r.FormValue("challenge")
						if !totp.Verify(challenge, time.Now().Unix()) {
							r = feature.AddErrorNotice(r, true, errors.OtpChallengeFailed(printer))
						} else {

							errors.Must(f.removeNewSecretKey(r))
							errors.Must(f.setSecureProvision(provision, secret, r))

							// finalize and reload same page
							return

						}

					}
				}

			}

		}

	}

	handled = true

	au := userbase.GetCurrentAuthUser(r)

	ctx := context.Context{
		"FeatureInfo": f.SiteFeatureInfo(r),
		"FormAction":  r.URL.Path,
		"Nonces": feature.Nonces{
			{Name: SetupNonceName, Key: SetupNonceKey},
		},
		"ProvisionLabel": provision,
	}

	if secret == "" {
		// re-use existing new-secret, simple page reload clears and generates another new secret
		secret = gotp.RandomSecret(16)
		errors.Must(f.setNewSecretKey(secret, r))
	}

	totpUri := gotp.NewDefaultTOTP(secret).ProvisioningUri(au.GetEmail(), f.Enjin.SiteName())

	ctx.SetSpecific("TotpUri", totpUri)
	ctx.SetSpecific("TotpSecret", secret)

	if qrc, ee := qrcode.NewWith(totpUri,
		qrcode.WithEncodingMode(qrcode.EncModeByte),
		qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionQuart),
	); ee != nil {
		log.ErrorRF(r, "error constructing qrcode instance: %v", ee)
	} else {
		bb := strings.NewByteBuffer()
		qw := standard.NewWithWriter(bb,
			standard.WithBuiltinImageEncoder(standard.PNG_FORMAT),
			standard.WithQRWidth(4),
			standard.WithBgTransparent(),
		)
		if ee = qrc.Save(qw); ee != nil {
			log.ErrorRF(r, "error saving qrcode to buffer: %v", ee)
		} else {
			data := bb.Bytes()
			encoded := base64.StdEncoding.EncodeToString(data)
			dataUri := fmt.Sprintf(`data:image/png;base64,%s`, encoded)
			ctx.SetSpecific("TotpQrcodeUri", dataUri)
		}
	}

	t := f.Site().SiteTheme()
	if err = f.Site().PrepareAndServePage("site-auth", "app-totp--setup", r.URL.Path, t, w, r, ctx); err != nil {
		log.ErrorRF(r, "error preparing and serving totp--setup page: %v", err)
	}
	return
}