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
	beContext "github.com/go-enjin/be/pkg/context"
	berrs "github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	bePath "github.com/go-enjin/be/pkg/path"
	"github.com/go-enjin/be/pkg/request"
)

/*
- ServeChallengeRequest is only ever called upon successful sign-in, at the end of the authorize sign-in step
- ServeChallengeRequest includes a radio list of already-setup multifactor options
- if no mfa has been configured by the user, force them to select one and set it up right now
- this process may or may not redirect the user to a specific path
- the form uses SiteAuthChallengePath as the action
- ProcessChallengeRequest receives the form submission and finds the actual feature to process with
- ProcessChallengeRequest calls mfp.ProcessChallenge
- success challenge+timestamp+mfp-tag are stored in claims .a[0]{c,t,f} (.a is []Context{})
- if more than one factor is required, repeat the process until the number of required factors is met
- every request uses mfp.VerifyChallengeRequest(claims, w, r)bool, logging users out if it returns false

- RequireVerification can be called from any enjin router handler
- this should never redirect the user away from the page they're on
- this should never sign the user out, just deny access to the pages which RequireVerification
- form uses "this URL" for the action instead of SiteAuthChallengePath
- RequireVerification sanity-wraps the ServeVerificationRequest
- ServeVerificationRequest includes a radio list of already-setup multifactor options
- ProcessVerificationRequest receives the form submission and finds the actual feature to process with
- ProcessVerificationRequest calls mfp.ProcessVerificationRequest
- success challenge+timestamp+mfp-tag are stored in claims .v[shasum]{c,t,f} (.v is map[string]Context{})
*/

func (f *CFeature) enforceVerifications(claims *feature.CSiteAuthClaims, w http.ResponseWriter, r *http.Request) (handled bool) {

	// handle required verification requests
	verifyTarget := claims.Context.String(gVerifyingTargetKey, "")
	if verifyTarget != "" {
		var proceed bool
		if _, proceed = bePath.MatchCut(r.URL.Path, f.SiteAuthSignInPath()); !proceed {
			if _, proceed = bePath.MatchCut(r.URL.Path, gVerifyingTargetKey); !proceed {
				claims.Context.DeleteKeys(gVerifyTargetKey, gVerifyingTargetKey)
			}
		}
		if proceed {
			var redirect string
			if handled, redirect = f.ProcessVerificationRequest(verifyTarget, w, r); handled {
				return
			} else if handled = redirect != ""; handled {
				f.Enjin.ServeRedirect(redirect, w, r)
				return
			}
		}
	}

	// validate existing required verifications
	var redirect string
	if handled, redirect = f.VerifyRequiredVerifications(claims, w, r); handled {
		return
	} else if handled = redirect != ""; handled {
		f.Enjin.ServeRedirect(redirect, w, r)
		return
	}

	return
}

func (f *CFeature) RequireVerification(verifyTarget string, w http.ResponseWriter, r *http.Request) (handled bool, redirect string) {
	var claims *feature.CSiteAuthClaims
	if claims = f.getPrivateClaims(r); claims == nil {
		f.Enjin.ServeNotFound(w, r)
		return
	}

	if claim, ok := claims.GetVerifiedFactor(verifyTarget); ok {
		if tag := f.mfa.Features.Find(claim.K); tag.IsNil() {
			claims.RevokeVerifiedFactor(verifyTarget)
		} else if mfp := f.mfa.Features.Get(tag); mfp.VerifyClaimFactor(claim, f, r) {
			// existing claim verified request
			return
		}
	}

	claims.Context.SetSpecific(gVerifyTargetKey, verifyTarget)

	if r.Method == http.MethodPost {
		handled, redirect = f.ProcessVerificationRequest(verifyTarget, w, r)
		return
	}
	handled, redirect = f.ServeVerificationRequest(verifyTarget, w, r)
	return
}

func (f *CFeature) ProcessVerificationRequest(verifyTarget string, w http.ResponseWriter, r *http.Request) (handled bool, redirect string) {

	var claims *feature.CSiteAuthClaims
	if claims = f.getPrivateClaims(r); claims == nil {
		f.Enjin.ServeNotFound(w, r)
		return
	} else if claim, ok := claims.GetVerifiedFactor(verifyTarget); ok {
		if tag := f.mfa.Features.Find(claim.K); tag.IsNil() {
			claims.RevokeVerifiedFactor(verifyTarget)
		} else if mfp := f.mfa.Features.Get(tag); mfp.VerifyClaimFactor(claim, f, r) {
			// existing claim verified request
			redirect = claims.Context.String(gVerifyingTargetKey, claims.Context.String(gVerifyTargetKey, verifyTarget))
			return
		}
	}

	if r.Method != http.MethodPost {
		f.ServeVerificationRequest(verifyTarget, w, r)
		return
	}

	printer := message.GetPrinter(r)

	if nonce := request.SafeQueryFormValue(r, VerificationNonceName); nonce != "" {

		if f.Enjin.VerifyNonce(VerificationNonceKey, nonce) {

			if provision := request.SafeQueryFormValue(r, "provision"); provision != "" {
				if kebab, pFactor, ok := strings.Cut(provision, ";"); ok {

					submit := r.FormValue("submit")

					if tag := f.mfa.Features.Find(submit); !tag.IsNil() {
						mfp := f.mfa.Features.Get(tag)
						if handled, redirect = mfp.ProcessVerification(verifyTarget, pFactor, "", f, claims, w, r); handled {
							// challenge handled
							return
						} else if redirect == "" {
							// challenge success
							redirect = claims.Context.String(gVerifyTargetKey, verifyTarget)
							claims.Context.DeleteKeys(gVerifyTargetKey, gVerifyingTargetKey)
						}
						f.Enjin.ServeRedirect(redirect, w, r)
						return
					}

					if submit == "challenge" {
						if challenge := request.SafeQueryFormValue(r, "challenge"); challenge != "" {
							if tag := f.mfa.Features.Find(kebab); tag.IsNil() {
								log.ErrorRF(r, "invalid challenge provision received")
								r = feature.AddErrorNotice(r, true, berrs.UnexpectedError(printer))
							} else {
								mfp := f.mfa.Features.Get(tag)
								if handled, redirect = mfp.ProcessVerification(verifyTarget, pFactor, challenge, f, claims, w, r); handled {
									// challenge handled
									return
								} else if redirect == "" {
									// challenge success
									redirect = claims.Context.String(gVerifyTargetKey, verifyTarget)
									claims.Context.DeleteKeys(gVerifyTargetKey, gVerifyingTargetKey)
								}
								f.Enjin.ServeRedirect(redirect, w, r)
								return

							}
						}
					}

				}
			}

		} else { // end f.Enjin.VerifyNonce
			r = feature.AddErrorNotice(r, true, berrs.FormExpiredError(printer))
		}

	}

	f.ServeVerificationRequest(verifyTarget, w, r)
	return
}

func (f *CFeature) ServeVerificationRequest(verifyTarget string, w http.ResponseWriter, r *http.Request) (handled bool, redirect string) {

	var claims *feature.CSiteAuthClaims
	if claims = f.getPrivateClaims(r); claims == nil {
		f.Enjin.ServeNotFound(w, r)
		return
	}

	claims.Context.SetSpecific(gVerifyingTargetKey, claims.Context.String(gVerifyTargetKey, verifyTarget))

	ctx := beContext.Context{
		"FeatureInfo": f.SiteFeatureInfo(r),
		"FormAction":  f.SiteAuthSignInPath(),
		"Nonces": feature.Nonces{
			{Name: VerificationNonceName, Key: VerificationNonceKey},
		},
	}

	var ready int
	if ready = f.prepareChallengeContext(r, ctx); ready == 0 {
		// verify needs just one factor to be ready and the site.RequireVerification method guards on the case of having
		// no factors ready so this code should not allow the request
		if f.numFactorsRequired == 0 {
			// none ready and none required, allow request to proceed
			return
		}
		handled = true
		f.Enjin.ServeNotFound(w, r)
		return
	}

	factors := claims.GetAllFactors()

	provision := request.SafeQueryFormValue(r, "provision")
	if mfpTag, factor, ok := strings.Cut(provision, ";"); ok {
		ctx.SetSpecific("ProvisionTag", mfpTag)
		ctx.SetSpecific("ProvisionFactor", factor)
	}

	ctx.SetSpecific("NumFactorsVerified", factors.Len())
	ctx.SetSpecific("NumFactorsRequired", f.numFactorsRequired)

	t := f.Site().SiteTheme()
	if err := f.Site().PrepareAndServePage("site-auth", "otp--verification", verifyTarget, t, w, r, ctx); err != nil {
		log.ErrorRF(r, "error serving prepared mfa--challenge page: %v", err)
		panic(err)
	}

	handled = true
	return
}
