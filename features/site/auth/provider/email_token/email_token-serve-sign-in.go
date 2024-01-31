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

package email_token

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/iancoleman/strcase"

	clStrings "github.com/go-corelibs/strings"
	"github.com/go-corelibs/x-text/message"
	"github.com/go-enjin/be/features/site/auth"
	"github.com/go-enjin/be/pkg/context"
	berrs "github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/request"
)

func (f *CFeature) SiteAuthSignInHandler(w http.ResponseWriter, r *http.Request, saf feature.SiteAuthFeature) (claims *feature.CSiteAuthClaims, redirect string, err error) {
	// TODO: factor in nonce checks with email links and form posts

	printer := message.GetPrinter(r)

	emailSentMessage := printer.Sprintf("New sign-in token email sent")

	if r.Form.Has("email") {

		var denied string
		if r.Method == http.MethodGet {
			if nonceValue := r.URL.Query().Get(SignInLinkNonceName); nonceValue != "" {
				if !f.Enjin.VerifyNonce(SignInLinkNonceKey, nonceValue) {
					denied = berrs.IncompleteLinkError(printer)
				}
			}
		} else if r.Method == http.MethodPost {
			if nonceValue := r.FormValue(SignInFormNonceName); nonceValue != "" {
				if !f.Enjin.VerifyNonce(SignInFormNonceKey, nonceValue) {
					denied = berrs.IncompleteFormError(printer)
				}
			}
		} else {
			denied = berrs.BadRequestError(printer)
		}

		if denied != "" {
			r = feature.AddErrorNotice(r, true, denied)
			f.ServeSignInConfirmationPage("", saf, w, r)
			return
		}

	}

	// get query params
	var email, token, emailTokenKey string

	if email = request.SafeQueryFormEmail(r, "email"); email == "" {
		// if no email, serve sign-in page
		f.ServeSignInConfirmationPage("", saf, w, r)
		return
	} else if email = strings.ToLower(email); email == "" {
		// nop
	} else if emailTokenKey = "sign-in-email-token-" + strcase.ToKebab(email); emailTokenKey == "" {
		// chained nop
	} else if token = request.SafeQueryFormHash10(r, "token"); token == "" {
		// no token, send email confirmation

		su := f.Site().SiteUsers()
		rid := su.MakeRealID(email)
		eid := su.MakeEnjinID(rid)
		if active, locked, _, ee := su.GetUserStatus(r, eid); ee == nil && (!active || locked) {
			// denied access
			if !active {
				log.WarnRF(r, "sign-in attempted with deactivated account: %q", email)
			} else if locked {
				log.WarnRF(r, "sign-in attempted with admin-locked account: %q", email)
			}
			r = feature.AddInfoNotice(r, true, emailSentMessage)
			f.ServeSignInConfirmationPage(email, saf, w, r)
			return
		}

		// sign-ups
		if !saf.IsUserAllowed(email) {
			log.WarnRF(r, "sign-in attempted with denied email: %q", email)
			r = feature.AddInfoNotice(r, true, emailSentMessage)
			f.ServeSignInConfirmationPage(email, saf, w, r)
			return
		}

		duration := time.Minute * 5
		_, emailLinkToken := f.Enjin.CreateTokenWith(emailTokenKey, duration)
		emailLinkNonce := f.Enjin.CreateNonce(SignInLinkNonceKey)
		link := fmt.Sprintf(
			"%s%s?submit=email&email=%s&token=%s&%s=%s",
			request.ParseDomainUrl(r),
			saf.SiteAuthSignInPath(),
			email,
			emailLinkToken,
			SignInLinkNonceName,
			emailLinkNonce,
		)

		subject := printer.Sprintf("%[1]s Sign-In Token", f.Enjin.SiteName())
		bodyCtx := context.Context{
			"Name":       clStrings.NameFromEmail(email),
			"Email":      email,
			"Link":       link,
			"Token":      emailLinkToken,
			"Duration":   strings.TrimSuffix(duration.Round(time.Minute).String(), "0s"),
			"Expiration": time.Now().Add(duration),
			"SiteName":   f.Enjin.SiteName(),
		}

		if err = f.sendUserEmail(r, email, subject, "sign-in--email-token", bodyCtx); err != nil {
			log.ErrorRF(r, "error sending sign-in email: %q - %v", email, err)
			r = feature.AddErrorNotice(r, true, berrs.UnexpectedError(printer))
		} else {
			log.DebugF("sign-in email sent to: %q", email)
			r = feature.AddInfoNotice(r, true, emailSentMessage)
		}

		f.ServeSignInConfirmationPage(email, saf, w, r)
		return
	}

	if !f.Enjin.VerifyToken(emailTokenKey, token) {
		log.ErrorRF(r, "invalid sign-in token received for: %q", email)
		r = feature.AddErrorNotice(r, true, printer.Sprintf("sign-in token expired or invalid"))
		f.ServeSignInConfirmationPage(email, saf, w, r)
		return
	}

	ctx := context.Context{}
	claims = saf.MakeAuthClaims(f.SiteFeatureKey(), email, ctx)

	su := f.Site().SiteUsers()
	if active, ee := su.GetUserActive(r, claims.EID); ee == nil && !active {
		err = errors.New(berrs.UnexpectedError(printer))

	}

	return
}

func (f *CFeature) SiteAuthLoginCallback(w http.ResponseWriter, r *http.Request, saf feature.SiteAuthFeature) (err error) {
	return
}

func (f *CFeature) SiteAuthSignOutHandler(w http.ResponseWriter, r *http.Request, saf feature.SiteAuthFeature) (handled bool, redirect string, err error) {
	// nothing to do, site auth handles claims reset
	return
}

func (f *CFeature) ServeSignInConfirmationPage(email string, saf feature.SiteAuthFeature, w http.ResponseWriter, r *http.Request) {

	ctx := context.Context{
		"FeatureInfo": f.SiteFeatureInfo(r),
		"FormAction":  r.URL.Path,
		"Nonces": feature.Nonces{
			{Name: auth.SignInNonceName, Key: auth.SignInNonceKey},
			{Name: SignInFormNonceName, Key: SignInFormNonceKey},
		},
		"EmailAddress": email,
	}

	// need list of auth provider form fields and actions?

	t := saf.Site().SiteTheme()
	r = request.SetHomePath(r, saf.SiteAuthSignInPath())
	if err := saf.Site().PrepareAndServePage("site-auth", "sign-in--email-token", saf.SiteAuthSignInPath(), t, w, r, ctx); err != nil {
		log.ErrorRF(r, "error preparing and serving sign-in--email-token page: %v", err)
		panic(err)
	}

	return
}
