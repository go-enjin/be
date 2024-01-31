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

	"github.com/go-corelibs/slices"
	"github.com/go-corelibs/x-text/message"
	berrs "github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
)

func (f *CFeature) ProcessChallenge(name, challenge string, saf feature.SiteAuthFeature, claims *feature.CSiteAuthClaims, w http.ResponseWriter, r *http.Request) (handled bool, redirect string) {

	printer := message.GetPrinter(r)

	if claim, ok := claims.GetFactor(f.KebabTag, name); ok && f.VerifyClaimFactor(claim, saf, r) {
		// existing factor verified
		return
	}
	claims.RevokeFactor(f.KebabTag, name)

	if codes, consumed, err := f.getSecureProvision(name, r); err != nil {

		log.ErrorRF(r, "developer error: cannot process challenge for a factor that is not present in the secure context")
		handled = true
		f.Enjin.ServeNotFound(w, r)
		return

	} else {

		switch r.FormValue("submit") {

		case "challenge":

			if challenge = strings.TrimSpace(strings.ToLower(challenge)); len(challenge) == 10 {
				if !slices.Within(challenge, consumed) {
					if slices.Within(challenge, codes) {
						codes = slices.Prune(codes, challenge)
						consumed = append(consumed, challenge)
						berrs.Must(f.setSecureProvision(name, codes, consumed, r))
						claims.SetFactor(&feature.CSiteAuthClaimsFactor{
							K: f.KebabTag,
							N: name,
							E: -1,
							T: -1,
							C: challenge,
						})
						if count := len(codes); count > 0 {
							f.Site().PushWarnNotice(claims.EID, true, printer.Sprintf("There are %[1]d backup codes remaining with %[2]s", count, name))
						} else {
							f.Site().PushWarnNotice(claims.EID, true, printer.Sprintf("There are no backup codes remaining with %[1]s", name))
						}
						return
					}
				}
			}

		}

	}

	handled = true
	r = feature.AddErrorNotice(r, true, berrs.OtpChallengeFailed(printer))
	saf.ServeChallengeRequest(w, r)
	return
}
