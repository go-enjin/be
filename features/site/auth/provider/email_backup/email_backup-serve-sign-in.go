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
	"fmt"
	"net/http"
	"time"

	"github.com/Shopify/gomail"
	"github.com/iancoleman/strcase"

	"github.com/go-enjin/be/features/site/auth"
	"github.com/go-enjin/be/pkg/context"
	berrs "github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/request"
	"github.com/go-enjin/be/pkg/strings"
)

func (f *CFeature) SiteAuthSignInHandler(w http.ResponseWriter, r *http.Request, saf feature.SiteAuthFeature) (claims *feature.CSiteAuthClaims, redirect string, err error) {

	printer := lang.GetPrinterFromRequest(r)

	var denied string
	if r.Method == http.MethodGet {
		if nonceValue := r.URL.Query().Get(SignInLinkNonceName); nonceValue != "" {
			if !f.Enjin.VerifyNonce(SignInLinkNonceKey, nonceValue) {
				denied = berrs.IncompleteLinkError(printer)
			}
		}
	} else if r.Method == http.MethodPost {
		if nonceValue := r.URL.Query().Get(SignInFormNonceName); nonceValue != "" {
			if !f.Enjin.VerifyNonce(SignInFormNonceKey, nonceValue) {
				denied = berrs.IncompleteFormError(printer)
			}
		}
	} else {
		denied = printer.Sprintf("bad request")
	}

	if denied != "" {
		r = feature.AddErrorNotice(r, true, denied)
		//saf.ServeSignInPage(w, r)f
		f.ServeSignInConfirmationPage("", "", saf, w, r)
		return
	}

	// get query params
	var email, backupEmail, token string

	if email = request.SafeQueryFormEmail(r, "email"); email == "" {
		// if no email, serve sign-in page
		f.ServeSignInConfirmationPage("", "", saf, w, r)
		return
	} else if backupEmail = request.SafeQueryFormEmail(r, "backup-email"); backupEmail == "" {
		// if no backup email, serve sign-in page
		f.ServeSignInConfirmationPage(email, "", saf, w, r)
		return

	} else {
		su := f.Site().SiteUsers()
		rid := su.MakeRealID(email)
		eid := su.MakeEnjinID(rid)

		checkEmailMessage := printer.Sprintf("If both the email addresses are correct, an email with the confirmation token has been sent.")

		if su := f.Site().SiteUsers(); su.UserPresent(eid) {

			if _, present := f.getProvisionByEmail(eid, backupEmail, r); !present {
				// backup-email given is not actually provisioned
				su.UnlockUser(r, eid)
				log.WarnRF(r, "visitor attempting to sign-in via backup-email when no backup-email has been configured")
				r = feature.AddInfoNotice(r, true, checkEmailMessage)
				f.ServeSignInConfirmationPage(email, backupEmail, saf, w, r)
				return
			}

		} else {
			// there is no user matching the primary email
			log.WarnRF(r, "visitor attempting to sign-in via backup-email when primary email given does not exist")
			r = feature.AddInfoNotice(r, true, checkEmailMessage)
			f.ServeSignInConfirmationPage(email, backupEmail, saf, w, r)
			return
		}
	}

	// both email address fields received are correct (primary exists with backup provisioned to that same user)
	// check for token or send email with token
	emailTokenKey := "sign-in-backup-email-token-" + strcase.ToKebab(email)

	if token = request.SafeQueryFormHash10(r, "token"); token != "" {

		if f.Enjin.VerifyToken(emailTokenKey, token) {
			ctx := context.Context{}
			claims = saf.MakeAuthClaims(f.SiteFeatureKey(), email, ctx)
			return
		}

		log.ErrorRF(r, "invalid backup sign-in token received for: %q", email)
		r = feature.AddErrorNotice(r, true, printer.Sprintf("sign-in backup token expired or invalid"))
		saf.ServeSignInPage(w, r) // start from scratch
		return

	}

	// no token, send email confirmation

	duration := time.Minute * 5
	_, emailLinkToken := f.Enjin.CreateTokenWith(emailTokenKey, duration)
	emailLinkNonce := f.Enjin.CreateNonce(SignInLinkNonceKey)
	link := fmt.Sprintf(
		"%s%s?submit=backup-email&email=%s&backup-email=%s&token=%s&%s=%s",
		request.ParseDomainUrl(r),
		saf.SiteAuthSignInPath(),
		email,
		backupEmail,
		emailLinkToken,
		SignInLinkNonceName,
		emailLinkNonce,
	)

	bodyCtx := context.Context{
		"Name":       strings.NameFromEmail(email),
		"Email":      email,
		"Link":       link,
		"Token":      emailLinkToken,
		"Duration":   duration,
		"Expiration": time.Now().Add(duration),
	}

	unexpectedError := printer.Sprintf("an unexpected error occurred")

	var msg *gomail.Message
	if msg, err = f.emailProvider.NewEmail("sign-in--email-backup", bodyCtx); err != nil {
		log.ErrorRF(r, "error making sign-in email: %q - %v", email, err)
		r = feature.AddErrorNotice(r, true, unexpectedError)
	} else {
		msg.SetHeader("To", backupEmail)
		msg.SetHeader("Subject", printer.Sprintf("%[1]s Backup Sign-In Token", f.Enjin.SiteName()))
		if err = f.emailSender.SendEmail(r, f.emailAccount, msg); err != nil {
			log.ErrorRF(r, "error sending backup sign-in email: %q - %v", email, err)
			r = feature.AddErrorNotice(r, true, unexpectedError)
		} else {
			log.DebugF("sign-in email sent to: %q", email)
			r = feature.AddInfoNotice(r, false, printer.Sprintf("New sign-in backup token email sent"))
		}
	}

	f.ServeSignInConfirmationPage(email, backupEmail, saf, w, r)
	return
}

func (f *CFeature) SiteAuthLoginCallback(w http.ResponseWriter, r *http.Request, s feature.SiteAuthFeature) (err error) {
	return
}

func (f *CFeature) SiteAuthSignOutHandler(w http.ResponseWriter, r *http.Request, s feature.SiteAuthFeature) (handled bool, redirect string, err error) {
	// nothing to do, site-auth handles nonce checks and resetting the current user claims
	return
}

func (f *CFeature) ServeSignInConfirmationPage(email, backupEmail string, s feature.SiteAuthFeature, w http.ResponseWriter, r *http.Request) {

	ctx := context.Context{
		"FeatureInfo": f.SiteFeatureInfo(r),
		"FormAction":  r.URL.Path,
		"Nonces": feature.Nonces{
			{Name: auth.SignInNonceName, Key: auth.SignInNonceKey},
			{Name: SignInFormNonceName, Key: SignInFormNonceKey},
		},
		"EmailAddress":       email,
		"BackupEmailAddress": backupEmail,
	}

	// need list of auth provider form fields and actions?

	t := s.Site().SiteTheme()
	r = request.SetHomePath(r, s.SiteAuthSignInPath())
	if err := s.Site().PrepareAndServePage("site-auth", "sign-in--email-backup", s.SiteAuthSignInPath(), t, w, r, ctx); err != nil {
		log.ErrorRF(r, "error preparing and serving sign-in--email-backup page: %v", err)
		panic(err)
		return
	}

	return
}