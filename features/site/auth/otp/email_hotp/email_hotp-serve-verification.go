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
	"time"

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
	} else if _, userSecret, _, err := f.getSecureProvision(claim.N, r); err != nil || userSecret == "" {
		// err or user secret empty
	} else if hotp := f.makeHotp(userSecret); claim.C != "" && claim.T > 0 {
		// ready for verification
		if verified = hotp.Verify(claim.C, int(claim.T)); verified {
			claim.E = time.Now().Add(saf.GetVerifiedDuration()).Unix()
		}
	}
	return
}

func (f *CFeature) ProcessVerification(verifyTarget, name, challenge string, saf feature.SiteAuthFeature, claims *feature.CSiteAuthClaims, w http.ResponseWriter, r *http.Request) (handled bool, redirect string) {

	printer := message.GetPrinter(r)

	if claim, ok := claims.GetVerifiedFactor(verifyTarget); ok && f.VerifyClaimFactor(claim, saf, r) {
		// existing factor verified
		return
	}
	claims.RevokeVerifiedFactor(verifyTarget)

	var err error
	var count int64
	var email, userSecret string
	if email, userSecret, count, err = f.getSecureProvision(name, r); err != nil {
		log.ErrorRF(r, "developer error: cannot process verification for a factor that is not present in the secure context")
		handled = true
		f.Enjin.ServeInternalServerError(w, r)
		return
	}

	switch r.FormValue("submit") {

	case f.SiteMultiFactorKey(), f.KebabTag:
		hotp := f.makeHotp(userSecret)
		if m := f.sendNewToken(email, hotp.At(int(count)), r); m != nil {
			r = m
		}
		handled, redirect = saf.ServeVerificationRequest(verifyTarget, w, r)
		return

	case "challenge":
		if hotp := f.makeHotp(userSecret); hotp.Verify(challenge, int(count)) {
			berrs.Must(f.setSecureProvision(name, email, userSecret, count+1, r))
			claim := feature.NewSiteAuthClaimsFactor(f.KebabTag, name, -1, count, challenge)
			claims.SetVerifiedFactor(verifyTarget, claim)
			// request allowed
			return
		}
	}

	r = feature.AddErrorNotice(r, true, berrs.OtpChallengeFailed(printer))
	handled, redirect = saf.ServeVerificationRequest(verifyTarget, w, r)
	return
}
