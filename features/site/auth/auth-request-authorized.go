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
	"strings"

	"github.com/go-corelibs/x-text/message"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/userbase"
)

func (f *CFeature) AuthorizeUserSignIn(w http.ResponseWriter, r *http.Request, claims *feature.CSiteAuthClaims) (handled bool, m *http.Request) {

	printer := message.GetPrinter(r)
	email := strings.ToLower(claims.Email)
	unknownErrMessage := printer.Sprintf("an unknown error occurred")
	su := f.Site().SiteUsers()

	if f.IsUserAllowed(email) {
		if !su.UserPresent(claims.EID) {
			// user allowed but doesn't exist yet
			if err := su.SignUpUser(r, claims); err != nil {
				handled = true
				log.ErrorRF(r, "error creating new user: %v", err)
				r = feature.AddErrorNotice(r, true, unknownErrMessage)
				f.ServeSignInPage(w, r)
				return
			}
		}
	} else {
		handled = true
		if f.allowSignups {
			r = feature.AddErrorNotice(r, true, unknownErrMessage)
		} else {
			r = feature.AddErrorNotice(r, true, printer.Sprintf("This site does not allow public sign-ups."))
		}
		if modified := f.resetCurrentUser(w, r); modified != nil {
			r = modified
		}
		f.ServeSignInPage(w, r)
		return
	}

	var err error
	su.LockUser(r, claims.EID)

	var au feature.User
	if au, err = su.RetrieveUser(r, claims.EID); err != nil {
		su.UnlockUser(r, claims.EID)
		log.ErrorRF(r, "error retrieving user %q: %v", claims.EID, err)
		r = feature.AddErrorNotice(r, true, unknownErrMessage)
		f.ServeSignInPage(w, r)
		handled = true
		return
	}

	su.UnlockUser(r, claims.EID)

	if handled = au.GetAdminLocked(); handled {
		r = feature.AddErrorNotice(r, false, printer.Sprintf("Your account is locked by site management, sign-in request denied."))
		f.Enjin.ServeForbidden(w, r)
		return
	}

	r = userbase.SetCurrentUser(au, f.setPrivateClaims(r, claims))

	for _, sf := range f.Site().SiteFeatures() {
		if uu, ok := sf.This().(feature.SiteUserRequestModifier); ok {
			if mm := uu.ModifyUserRequest(au, r); mm != nil {
				r = mm
			}
		}
	}

	// user is authenticated with the primary account factor

	/*
		- get a list of all features with account setup stages
		- all primary backup factors are required to be setup
		- configured number of distinct otp factors must be setup
		- the process for performing the steps loops through the above (in the same order) until all features report ready
	*/

	if handled = f.enforceUserSetupStages(claims, w, r); handled {
		// enforcing user-account must-setup stages
		return
	} else if handled = f.enforceMultiFactors(claims, w, r); handled {
		// enforcing multi-factor requirements
		return
	} else if handled = f.enforceVerifications(claims, w, r); handled {
		// enforcing required verifications
		return
	}

	m = r
	return
}

func (f *CFeature) VerifyRequiredVerifications(claims *feature.CSiteAuthClaims, w http.ResponseWriter, r *http.Request) (handled bool, redirect string) {
	if parts := strings.Split(r.URL.Path, "/"); len(parts) > 0 {
		var revoked bool
		var verifyTarget string
		for _, part := range parts[1:] {
			verifyTarget += "/" + part
			if claim, ok := claims.GetVerifiedFactor(verifyTarget); ok {
				if revoked {
					claims.RevokeVerifiedFactor(verifyTarget)
				} else {
					mfp := f.mfa.Features.Get(f.mfa.Features.Find(claim.K))
					if revoked = mfp == nil; revoked {
						claims.RevokeVerifiedFactor(verifyTarget)
						redirect = verifyTarget
					} else if revoked = !mfp.VerifyClaimFactor(claim, f, r); revoked {
						claims.RevokeVerifiedFactor(verifyTarget)
						redirect = verifyTarget
					}
				}
			}
		}
	}
	return
}
