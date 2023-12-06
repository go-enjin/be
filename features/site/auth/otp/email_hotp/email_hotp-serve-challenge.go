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

	berrs "github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
)

func (f *CFeature) ProcessChallenge(name, challenge string, saf feature.SiteAuthFeature, claims *feature.CSiteAuthClaims, w http.ResponseWriter, r *http.Request) (handled bool, redirect string) {

	kebab := f.Tag().Kebab()
	printer := lang.GetPrinterFromRequest(r)

	if claim, ok := claims.GetFactor(kebab, name); ok && f.VerifyClaimFactor(claim, saf, r) {
		// existing factor verified
		return
	}
	claims.RevokeFactor(kebab, name)

	var err error
	var count int64
	var email, userSecret string
	if email, userSecret, count, err = f.getSecureProvision(name, r); err != nil {
		log.ErrorRF(r, "developer error: cannot process challenge for a factor that is not present in the secure context")
		handled = true
		f.Enjin.ServeInternalServerError(w, r)
		return
	}

	switch r.FormValue("submit") {

	case f.SiteMultiFactorKey(), kebab:
		hotp := f.makeHotp(userSecret)
		if m := f.sendNewToken(email, hotp.At(int(count)), r); m != nil {
			r = m
		}
		handled = true
		saf.ServeChallengeRequest(w, r)
		return

	case "challenge":
		if hotp := f.makeHotp(userSecret); hotp.Verify(challenge, int(count)) {
			berrs.Must(f.setSecureProvision(name, email, userSecret, count+1, r))
			claim := feature.NewSiteAuthClaimsFactor(kebab, name, -1, count, challenge)
			claims.SetFactor(claim)
			// request allowed
			return
		}
	}

	handled = true
	r = feature.AddErrorNotice(r, true, berrs.OtpChallengeFailed(printer))
	saf.ServeChallengeRequest(w, r)
	return
}