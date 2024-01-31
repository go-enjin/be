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

	"github.com/go-corelibs/x-text/message"
	"github.com/go-enjin/be/pkg/context"
	berrs "github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/request"
)

func (f *CFeature) HandleSignOutPage(w http.ResponseWriter, r *http.Request) {
	if claims := f.getPrivateClaims(r); claims == nil {
		f.Enjin.ServeNotFound(w, r)
		return
	}

	var err error
	var handled bool
	var redirect string
	if handled, redirect, err = f.SignOutPageHandler(w, r); err != nil && !errors.Is(err, berrs.ErrProviderNotFound) {
		log.ErrorRF(r, "sign-out page handler error: %v", err)
		f.ServeSignOutPage(w, r)
		return
	} else if handled {
		return
	} else if err != nil {
		log.ErrorRF(r, "sign-out page handler error: %v", err)
	}

	r = f.resetCurrentUser(w, r)

	if redirect != "" {
		f.Enjin.ServeRedirect(redirect, w, r)
	} else {
		f.Enjin.ServeRedirect("/", w, r)
	}
	return
}

func (f *CFeature) SignOutPageHandler(w http.ResponseWriter, r *http.Request) (handled bool, redirect string, err error) {
	if nonce := request.SafeQueryFormValue(r, SignOutNonceName); nonce != "" {
		if f.Enjin.VerifyNonce(SignOutNonceKey, nonce) {
			if claims := f.getPrivateClaims(r); claims != nil {
				if len(claims.Audience) > 0 {
					audience := claims.Audience[0]
					for _, sap := range f.sap.Features {
						if sap.SiteFeatureKey() == audience {
							handled, redirect, err = sap.SiteAuthSignOutHandler(w, r, f)
							return
						}
					}
				}
			}
			err = berrs.ErrProviderNotFound
			return
		}
		handled = true
		r = feature.AddErrorNotice(r, true, berrs.FormExpiredError(message.GetPrinter(r)))
		f.ServeSignInPage(w, r)
		return
	}
	// reload this page
	redirect = f.SiteAuthSignOutPath()
	return
}

func (f *CFeature) ServeSignOutPage(w http.ResponseWriter, r *http.Request) {
	if claims := f.getPrivateClaims(r); claims == nil {
		f.Enjin.ServeNotFound(w, r)
		return
	}

	signOutPath := f.SiteAuthSignOutPath()
	ctx := context.Context{
		"FeatureInfo": f.SiteFeatureInfo(r),
		"FormAction":  signOutPath,
		"Nonces": feature.Nonces{
			{Name: SignOutNonceName, Key: SignOutNonceKey},
		},
	}

	t := f.Site().SiteTheme()
	if err := f.Site().PrepareAndServePage("site-auth", "sign-out", signOutPath, t, w, r, ctx); err != nil {
		log.ErrorRF(r, "error preparing and serving sign-out page: %v", err)
		panic(err)
	}
}
