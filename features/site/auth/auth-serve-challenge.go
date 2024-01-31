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

	"github.com/go-corelibs/slices"
	"github.com/go-corelibs/x-text/message"
	"github.com/go-enjin/be/pkg/context"
	berrs "github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/request"
)

type multiFactor struct {
	Tag     string
	Key     string
	Ready   bool
	Label   string
	Factors []string
}

func (f *CFeature) enforceMultiFactors(claims *feature.CSiteAuthClaims, w http.ResponseWriter, r *http.Request) (handled bool) {
	printer := message.GetPrinter(r)
	// setup and enforce multi-factor requirements
	if f.numFactorsRequired > 0 {
		count := claims.GetAllFactors().Len()
		if handled = count < f.numFactorsRequired; handled {
			var text string
			if f.numFactorsRequired > 1 {
				text = printer.Sprintf(`Account Verification (%[1]d of %[2]d)`, count+1, f.numFactorsRequired)
			} else {
				text = printer.Sprintf(`Account Verification`)
			}
			r = feature.AddImportantNotice(r, false, text)
			if r.Method == http.MethodPost {
				f.ProcessChallengeRequest(w, r)
				return
			}
			claims.Context.SetSpecific(gRedirectKey, f.Site().SitePath())
			f.ServeChallengeRequest(w, r)
		}
	}
	return
}

func (f *CFeature) makeMfaKeysLookup(r *http.Request) (distinct, total int, order []string, lookup map[string]*feature.CSiteAuthMultiFactorInfo) {
	lookup = make(map[string]*feature.CSiteAuthMultiFactorInfo)

	claims := f.getPrivateClaims(r)
	factors := claims.GetAllFactors()

	for _, tag := range f.mfa.Features.Tags() {
		mff := f.mfa.Features.Get(tag)
		if mff.CurrentUserFactorsReadyCount(r) > 0 {
			kebab := mff.Tag().Kebab()
			order = append(order, kebab)
			lookup[kebab] = mff.SiteMultiFactorInfo(r)
			for _, factor := range factors {
				if slices.Within(factor.N, lookup[kebab].Factors) {
					lookup[kebab].Claimed = append(lookup[kebab].Claimed, factor.N)
				}
			}
			if lookup[kebab].Ready {
				distinct += 1
				total += len(lookup[kebab].Factors)
			}
		}
	}

	return
}

func (f *CFeature) prepareChallengeContext(r *http.Request, ctx context.Context) (ready int) {
	var order []string
	var lookup map[string]*feature.CSiteAuthMultiFactorInfo
	ready, _, order, lookup = f.makeMfaKeysLookup(r)
	ctx.SetSpecific("MultiFactorKeys", order)
	ctx.SetSpecific("MultiFactorLookup", lookup)
	ctx.SetSpecific("MultiFactorsReady", ready)
	return
}

func (f *CFeature) ProcessChallengeRequest(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		f.ServeChallengeRequest(w, r)
		return
	}

	var claims *feature.CSiteAuthClaims
	if claims = f.getPrivateClaims(r); claims == nil {
		f.Enjin.ServeNotFound(w, r)
		return
	}

	var handled bool
	var redirect string
	printer := message.GetPrinter(r)

	if nonce := request.SafeQueryFormValue(r, ChallengeNonceName); nonce != "" {
		// challenge form present
		if !f.Enjin.VerifyNonce(ChallengeNonceKey, nonce) {
			r = feature.AddErrorNotice(r, true, berrs.FormExpiredError(printer))
			f.ServeChallengeRequest(w, r)
			return
		}
		// challenge nonce verified
	} else {
		// no nonce means start of process
		f.ServeChallengeRequest(w, r)
		return
	}

	// proceed with challenge process

	submit := request.SafeQueryFormValue(r, "submit")
	provider := request.SafeQueryFormValue(r, "provider")
	provision := request.SafeQueryFormValue(r, "provision")
	challenge := request.SafeQueryFormValue(r, "challenge")

	var validChallengeRequest bool
	if validChallengeRequest = submit == "challenge"; !validChallengeRequest {
		validChallengeRequest = !f.mfa.Features.Find(submit).IsNil()
	}
	if !validChallengeRequest {
		r = feature.AddErrorNotice(r, true, berrs.IncompleteFormError(printer))
		f.ServeChallengeRequest(w, r)
		return
	}

	// submit value verified

	if kebab, pFactor, ok := strings.Cut(provision, ";"); ok {

		if provider != kebab {
			// select first factor with provider
			kebab = provider
			pFactor = ""
		}

		var tag feature.Tag
		if tag = f.mfa.Features.Find(kebab); !tag.IsNil() {
			mfp := f.mfa.Features.Get(tag)

			if pFactor == "" {
				names := mfp.CurrentUserFactorsReady(r)
				if len(names) == 0 {
					log.ErrorRF(r, "invalid challenge factor received")
					r = feature.AddErrorNotice(r, true, berrs.UnexpectedError(printer))
					f.ServeChallengeRequest(w, r)
					return
				}
				pFactor = names[0]
			}

			if handled, redirect = mfp.ProcessChallenge(pFactor, challenge, f, claims, w, r); handled {
				// challenge handled
				return
			} else {
				if redirect == "" {
					if redirect = claims.Context.String(gRedirectKey, ""); redirect != "" {
						claims.Context.Delete(gRedirectKey)
						f.Enjin.ServeRedirect(redirect, w, r)
						return
					}
				}
				f.Enjin.ServeRedirectHomePath(w, r)
				return
			}

		} else {
			log.ErrorRF(r, "invalid challenge provision received")
			r = feature.AddErrorNotice(r, true, berrs.UnexpectedError(printer))
		}

	} else {
		log.ErrorRF(r, "broken challenge provision received")
		r = feature.AddErrorNotice(r, true, berrs.UnexpectedError(printer))
	}

	f.ServeChallengeRequest(w, r)
	return
}

func (f *CFeature) ServeChallengeRequest(w http.ResponseWriter, r *http.Request) {
	var claims *feature.CSiteAuthClaims
	if claims = f.getPrivateClaims(r); claims == nil {
		f.Enjin.ServeNotFound(w, r)
		return
	}

	thisPath := r.URL.Path
	challengePath := f.SiteAuthChallengePath()

	if !claims.Context.HasExact(gRedirectKey) {
		if signInPath := f.SiteAuthSignInPath(); r.URL.Path != signInPath {
			if r.URL.Path != challengePath {
				// not on the sign-in or mfa page, note the eventual redirect
				// not sure how this can happen anymore
				claims.Context.SetSpecific(gRedirectKey, thisPath)
			}
		}
	}

	ctx := context.Context{
		"FeatureInfo": f.SiteFeatureInfo(r),
		"FormAction":  challengePath,
		"Nonces": feature.Nonces{
			{Name: ChallengeNonceName, Key: ChallengeNonceKey},
		},
	}

	ready := f.prepareChallengeContext(r, ctx)
	factors := claims.GetAllFactors()

	provision := request.SafeQueryFormValue(r, "provision")
	if mfpTag, factor, ok := strings.Cut(provision, ";"); ok {
		ctx.SetSpecific("ProvisionTag", mfpTag)
		ctx.SetSpecific("ProvisionFactor", factor)
	}

	ctx.SetSpecific("NumFactorsReady", ready)
	ctx.SetSpecific("NumFactorsRequired", f.numFactorsRequired)
	ctx.SetSpecific("NumFactorsVerified", factors.Len())

	t := f.Site().SiteTheme()
	if err := f.Site().PrepareAndServePage("site-auth", "otp--challenge", thisPath, t, w, r, ctx); err != nil {
		log.ErrorRF(r, "error serving prepared mfa--challenge page: %v", err)
		panic(err)
	}
	return
}
