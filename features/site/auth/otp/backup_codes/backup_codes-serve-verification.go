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

package backup_codes

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-corelibs/slices"
	"github.com/go-corelibs/x-text/message"
	berrs "github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
)

func (f *CFeature) VerifyClaimFactor(claim *feature.CSiteAuthClaimsFactor, saf feature.SiteAuthFeature, r *http.Request) (verified bool) {
	if claim.E > 1 && time.Now().After(time.Unix(claim.E, 0)) {
		// claim is expired
		claim.T = 0
		claim.C = ""
	} else if _, consumed, err := f.getSecureProvision(claim.N, r); err != nil {
		// err or user secret empty
	} else if len(consumed) > 0 {
		// ready for verification - challenge is within consumed
		if verified = slices.Within(claim.C, consumed); verified {
			claim.E = time.Now().Add(saf.GetVerifiedDuration()).Unix()
		}
	}
	return
}

func (f *CFeature) ProcessVerification(verifyTarget, name, challenge string, saf feature.SiteAuthFeature, claims *feature.CSiteAuthClaims, w http.ResponseWriter, r *http.Request) (handled bool, redirect string) {
	if claim, ok := claims.GetVerifiedFactor(verifyTarget); ok && f.VerifyClaimFactor(claim, saf, r) {
		// existing factor verified
		return
	}
	claims.RevokeVerifiedFactor(verifyTarget)

	if codes, consumed, err := f.getSecureProvision(name, r); err != nil {
		log.ErrorRF(r, "developer error: cannot process verification for a factor that is not present in the secure context")
		handled = true
		f.Enjin.ServeNotFound(w, r)
		return
	} else if len(codes) > 0 {
		if challenge = strings.TrimSpace(strings.ToLower(challenge)); len(challenge) == 10 {

			if slices.Within(challenge, codes) {

				codes = slices.Prune(codes, challenge)
				consumed = append(consumed, challenge)
				berrs.Must(f.setSecureProvision(name, codes, consumed, r))
				expires := time.Now().Add(saf.GetVerifiedDuration()).Unix()
				claim := feature.NewSiteAuthClaimsFactor(f.KebabTag, name, expires, -1, challenge)
				claims.SetVerifiedFactor(verifyTarget, claim)
				// request allowed
				return

			}

		}
	}

	printer := message.GetPrinter(r)
	r = feature.AddErrorNotice(r, true, berrs.OtpChallengeFailed(printer))
	handled, redirect = saf.ServeVerificationRequest(verifyTarget, w, r)
	return
}
