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

package app_totp

import (
	"net/http"
	"time"

	"github.com/xlzd/gotp"

	"github.com/go-corelibs/x-text/message"
	"github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
)

func (f *CFeature) ProcessChallenge(name, challenge string, saf feature.SiteAuthFeature, claims *feature.CSiteAuthClaims, w http.ResponseWriter, r *http.Request) (handled bool, redirect string) {

	epoch := time.Now().Unix() // early to keep as close to the TOTP cycle as possible
	printer := message.GetPrinter(r)

	if claim, ok := claims.GetFactor(f.KebabTag, name); ok && f.VerifyClaimFactor(claim, saf, r) {
		// existing factor verified
		return
	}
	claims.RevokeFactor(f.KebabTag, name)

	if userSecret, err := f.getSecureProvision(name, r); err != nil {
		log.ErrorRF(r, "developer error: cannot process challenge for a factor that is not present in the secure context")
		handled = true
		f.Enjin.ServeNotFound(w, r)
		return
	} else if totp := gotp.NewDefaultTOTP(userSecret); totp.Verify(challenge, epoch) {
		claim := feature.NewSiteAuthClaimsFactor(f.KebabTag, name, -1, epoch, challenge)
		claims.SetFactor(claim)
		// request allowed
		return
	}

	handled = true
	r = feature.AddErrorNotice(r, true, errors.OtpChallengeFailed(printer))
	saf.ServeChallengeRequest(w, r)
	return
}
