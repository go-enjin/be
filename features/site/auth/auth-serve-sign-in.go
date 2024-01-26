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
	"errors"
	"net/http"

	"github.com/go-enjin/be/pkg/context"
	berrs "github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/request"
)

func (f *CFeature) HandleSignInPage(w http.ResponseWriter, r *http.Request) {
	var err error
	var handled bool
	var redirect string
	var claims *feature.CSiteAuthClaims

	if claims, handled, redirect, err = f.SignInPageHandler(w, r); err != nil {
		if !errors.Is(err, berrs.ErrNothingToDo) {
			log.ErrorRF(r, "sign-in page handler error: %v", err)
		}
		f.ServeSignInPage(w, r)
		return
	} else if claims != nil {
		if handled, r = f.AuthorizeUserSignIn(w, r, claims); handled {
			return
		}
	} else if handled {
		// served by sign-in handler, do not continue writing to w
		return
	}

	if redirect != "" {
		f.Enjin.ServeRedirect(redirect, w, r)
		return
	}
	f.Enjin.ServeRedirectHomePath(w, r)
}

func (f *CFeature) findSapFeatureByKey(key string) (sap feature.SiteAuthProvider) {
	for _, p := range f.sap.Features {
		if p.SiteFeatureKey() == key {
			sap = p
			return
		}
	}
	return
}

func (f *CFeature) SignInPageHandler(w http.ResponseWriter, r *http.Request) (claims *feature.CSiteAuthClaims, handled bool, redirect string, err error) {
	printer := lang.GetPrinterFromRequest(r)

	if nonce := request.SafeQueryFormValue(r, SignInNonceName); nonce != "" && f.Enjin.VerifyNonce(SignInNonceKey, nonce) {
		submit := request.SafeQueryFormValue(r, "submit")
		if handled = submit == "cancel"; handled {
			if m := f.resetCurrentUser(w, r); m != nil {
				r = m
			}
			// redirect to clear any browser form submission history
			f.Enjin.ServeRedirect(f.SiteAuthSignInPath(), w, r)
			return
		}
		audience := request.SafeQueryFormValue(r, "audience")
		sap := f.findSapFeatureByKey(audience)
		if handled = sap != nil; handled {
			claims, redirect, err = sap.SiteAuthSignInHandler(w, r, f)
			return
		}
		r = feature.AddErrorNotice(r, true, berrs.IncompleteFormError(printer))
	} else if nonce != "" {
		r = feature.AddErrorNotice(r, true, berrs.FormExpiredError(printer))
	}

	handled = true
	f.ServeSignInPage(w, r)
	return
}

func (f *CFeature) ServeSignInPage(w http.ResponseWriter, r *http.Request) {
	var err error

	if claims := f.getPrivateClaims(r); claims != nil {
		if f.numFactorsRequired <= claims.GetAllFactors().Len() {
			f.Enjin.ServeRedirect(f.Site().SitePath(), w, r)
			return
		}
	}

	t := f.Site().SiteTheme()
	signInPath := f.SiteAuthSignInPath()

	providers := feature.NewSiteInfosLookup[*feature.CSiteFeatureInfo]()
	backupProviders := feature.NewSiteInfosLookup[*feature.CSiteFeatureInfo]()

	for _, sap := range f.sap.Features {
		info := sap.SiteFeatureInfo(r)
		if info.Backup {
			backupProviders.Set(info.Key, info)
		} else {
			providers.Set(info.Key, info)
		}
	}

	ctx := context.Context{
		"FormAction": signInPath,
		"Nonces": feature.Nonces{
			{Name: SignInNonceName, Key: SignInNonceKey, Value: f.Enjin.CreateNonce(SignInNonceKey)},
		},
		"Providers":       providers,
		"BackupProviders": backupProviders,
		"PublicSignups":   f.allowSignups,
		"PrivateSignups":  len(f.allowedEmails) > 0,
	}

	// need list of auth provider form fields and actions?
	r = request.SetHomePath(r, signInPath)

	if err = f.Site().PrepareAndServePage("site-auth", "sign-in", signInPath, t, w, r, ctx); err != nil {
		log.ErrorRF(r, "error preparing and serving sign-in page: %v", err)
		panic(err)
	}

	return
}
