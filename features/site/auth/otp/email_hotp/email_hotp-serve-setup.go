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

	"github.com/Shopify/gomail"
	"github.com/xlzd/gotp"

	beContext "github.com/go-enjin/be/pkg/context"
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
	if handled := f.ProcessSetupPage(saf, w, r); !handled {
		f.Enjin.ServeRedirect(r.URL.Path, w, r)
	}
	return
}

func (f *CFeature) ProcessSetupPage(saf feature.SiteAuthFeature, w http.ResponseWriter, r *http.Request) (handled bool) {

	var err error
	var email, provision, secret string
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

		if secret = f.getNewSecretKey(r); secret != "" {

			if nonce := request.SafeQueryFormValue(r, SetupNonceName); nonce != "" {
				if !f.Enjin.VerifyNonce(SetupNonceKey, nonce) {
					r = feature.AddErrorNotice(r, true, errors.FormExpiredError(printer))

				} else if email = request.SafeQueryFormEmail(r, "email"); email != "" {

					hotp := f.makeHotp(secret)
					submit := r.FormValue("submit")

					switch submit {

					case "setup":

						if m := f.sendNewToken(email, hotp.At(0), r); m != nil {
							r = m
						}

					case "setup-confirm":

						if challenge := request.SafeQueryFormSixDigits(r, "challenge"); challenge != "" {
							if !hotp.Verify(challenge, 0) {
								r = feature.AddErrorNotice(r, true, errors.OtpChallengeFailed(printer))
							} else {
								errors.Must(f.removeNewSecretKey(r))
								errors.Must(f.setSecureProvision(provision, email, secret, 1, r))
								claim := feature.NewSiteAuthClaimsFactor(f.KebabTag, provision, -1, 0, challenge)
								saf.SetUserFactor(r, claim)
								return
							}
						}

					}
				}
			}

		}

	}

	handled = true

	ctx := beContext.Context{
		"FeatureInfo": f.SiteFeatureInfo(r),
		"FormAction":  r.URL.Path,
		"Nonces": feature.Nonces{
			{Name: SetupNonceName, Key: SetupNonceKey},
		},
		"ProvisionLabel": provision,
		"EmailAddress":   email,
	}

	if secret == "" {
		// re-use existing new-secret, simple page reload clears and generates another new secret
		secret = gotp.RandomSecret(16)
		errors.Must(f.setNewSecretKey(secret, r))
	}

	t := f.Site().SiteTheme()
	if err = f.Site().PrepareAndServePage("site-auth", "email-hotp--setup", r.URL.Path, t, w, r, ctx); err != nil {
		log.ErrorRF(r, "error preparing and serving totp--setup page: %v", err)
	}
	return
}

func (f *CFeature) sendNewToken(email, challenge string, r *http.Request) (m *http.Request) {
	var err error
	printer := lang.GetPrinterFromRequest(r)

	bodyCtx := beContext.Context{
		"Name":  strings.NameFromEmail(email),
		"Email": email,
		"Token": challenge,
	}

	var msg *gomail.Message
	if msg, err = f.emailProvider.NewEmail("email-hotp--passcode", bodyCtx); err != nil {
		log.ErrorRF(r, "error making email-hotp--passcode email: %q - %v", email, err)
		//m = feature.AddErrorNotice(r, true, errors.UnexpectedError(printer))
		f.Site().PushErrorNotice(userbase.GetCurrentEID(r), true, errors.UnexpectedError(printer))
	} else {
		msg.SetHeader("To", email)
		msg.SetHeader("Subject", printer.Sprintf("%[1]s Email Passcode", f.Enjin.SiteName()))
		if err = f.emailSender.SendEmail(r, f.emailAccount, msg); err != nil {
			log.ErrorRF(r, "error sending email-hotp--passcode email: %q - %v", email, err)
			//m = feature.AddErrorNotice(r, true, errors.UnexpectedError(printer))
			f.Site().PushErrorNotice(userbase.GetCurrentEID(r), true, errors.UnexpectedError(printer))
		} else {
			log.DebugF("email passcode sent to: %q", email)
			//m = feature.AddInfoNotice(r, false, printer.Sprintf("Email passcode sent"))
			f.Site().PushInfoNotice(userbase.GetCurrentEID(r), true, printer.Sprintf("Email passcode sent"))
		}
	}
	return

}
