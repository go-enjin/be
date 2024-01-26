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

	"github.com/iancoleman/strcase"

	"github.com/go-enjin/be/pkg/context"
	berrs "github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/request"
	"github.com/go-enjin/be/pkg/strings"
	"github.com/go-enjin/be/pkg/userbase"
)

func (f *CFeature) SiteUserSetupStageReady(eid string, r *http.Request) (ready bool) {
	su := f.Site().SiteUsers()
	su.RLockUser(r, eid)
	defer su.RUnlockUser(r, eid)
	ready = f.countSecureProvisions(eid, r) > 0
	return
}

func (f *CFeature) SiteUserSetupStageHandler(saf feature.SiteAuthFeature, w http.ResponseWriter, r *http.Request) {
	if redirect := f.ServeSetupPage(r.URL.Path, saf, w, r); redirect != "" {
		f.Enjin.ServeRedirect(redirect, w, r)
		return
	}
	f.Enjin.ServeRedirect(r.URL.Path, w, r)
}

func (f *CFeature) ServeSetupPage(settingsPath string, saf feature.SiteAuthFeature, w http.ResponseWriter, r *http.Request) (redirect string) {
	printer := lang.GetPrinterFromRequest(r)
	au := userbase.GetCurrentUser(r)
	eid := au.GetEID()
	email := au.GetEmail()
	su := f.Site().SiteUsers()

	if submit := request.SafeQueryFormValue(r, "submit"); submit == "cancel" {
		// cancel is just resetting form values with a reload
		f.Enjin.ServeRedirect(r.URL.Path, w, r)
		return
	}

	var provision string
	if provision = request.SafeQueryFormValue(r, "provision"); provision == "" {
		provision = printer.Sprintf("Backup Email")
	}
	for f.hasSecureProvision(eid, provision, r) {
		provision = strings.IncrementFileName(provision)
	}

	var denied string
	if r.Method == http.MethodGet {
		if nonceValue := request.SafeQueryFormValue(r, SetupNonceName); nonceValue != "" {
			if !f.Enjin.VerifyNonce(SetupNonceKey, nonceValue) {
				denied = berrs.IncompleteLinkError(printer)
			}
		} else {
			// hmm
		}
	} else if r.Method == http.MethodPost {
		if nonceValue := request.SafeQueryFormValue(r, SetupNonceName); nonceValue != "" {
			if !f.Enjin.VerifyNonce(SetupNonceKey, nonceValue) {
				denied = berrs.IncompleteFormError(printer)
			}
		} else {
			// hmm
		}
	} else {
		f.ServeSetupConfirmationPage(provision, email, "", saf, w, r)
		return
	}

	if denied != "" {
		r = feature.AddErrorNotice(r, true, denied)
		f.ServeSetupConfirmationPage(provision, email, "", saf, w, r)
		return
	}

	// get query params
	var backupEmail, token string

	if backupEmail = request.SafeQueryFormEmail(r, "backup-email"); backupEmail == "" {
		// if no backup email, serve sign-in page
		f.ServeSetupConfirmationPage(provision, email, "", saf, w, r)
		return

	} else {

		checkEmailMessage := printer.Sprintf("check your email for the confirmation token")

		if su.UserPresent(eid) {

			if _, present := f.getProvisionByEmail(eid, backupEmail, r); present {
				r = feature.AddErrorNotice(r, true, printer.Sprintf("%[1]s is already provisioned as \"%[2]s\"", backupEmail, provision))
				f.ServeSetupConfirmationPage(provision, email, "", saf, w, r)
				return
			} else if email == backupEmail {
				r = feature.AddErrorNotice(r, true, printer.Sprintf("A backup email address that is not your primary account is required"))
				f.ServeSetupConfirmationPage(provision, email, "", saf, w, r)
				return
			}

		} else {
			r = feature.AddWarnNotice(r, true, checkEmailMessage)
			f.ServeSetupConfirmationPage(provision, email, backupEmail, saf, w, r)
			return
		}
	}

	// both email address fields received are valid (primary exists, backup does not)
	// check for token or send email with token
	emailTokenKey := "sign-in-backup-email-token-" + strcase.ToKebab(backupEmail)

	if token = request.SafeQueryFormHash10(r, "token"); token != "" {
		if f.Enjin.VerifyToken(emailTokenKey, token) {
			berrs.Must(f.setSecureProvision(eid, provision, backupEmail, r))
			// reload current page
			f.Enjin.ServeRedirect(r.URL.Path, w, r)
			return
		}

		log.ErrorRF(r, "invalid backup sign-in token received for: %q", backupEmail)
		r = feature.AddErrorNotice(r, true, printer.Sprintf("sign-in backup token expired or invalid"))
		f.ServeSetupConfirmationPage(provision, email, "", saf, w, r)
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

	unexpectedError := berrs.UnexpectedError(printer)

	if msg, err := f.emailProvider.NewEmail("sign-in--email-backup", bodyCtx); err != nil {
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

	f.ServeSetupConfirmationPage(provision, email, backupEmail, saf, w, r)
	return
}

func (f *CFeature) ServeSetupConfirmationPage(provision, email, backupEmail string, s feature.SiteAuthFeature, w http.ResponseWriter, r *http.Request) {

	ctx := context.Context{
		"FeatureInfo": f.SiteFeatureInfo(r),
		"FormAction":  r.URL.Path,
		"Nonces": feature.Nonces{
			{Name: SetupNonceName, Key: SetupNonceKey},
		},
		"EmailAddress":       email,
		"BackupEmailAddress": backupEmail,
		"ProvisionLabel":     provision,
	}

	t := s.Site().SiteTheme()
	if err := s.Site().PrepareAndServePage("site-auth", "email-backup--setup", s.SiteAuthSignInPath(), t, w, r, ctx); err != nil {
		log.ErrorRF(r, "error preparing and serving sign-in--email-backup--setup page: %v", err)
		panic(err)
	}

	return
}
