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
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	clPath "github.com/go-corelibs/path"
)

func (f *CFeature) AuthenticateSiteRequest(w http.ResponseWriter, r *http.Request) (handled bool, modified *http.Request) {
	reqUri := forms.TrimQueryParams(r.RequestURI)
	sitePath := f.Site().SitePath()
	siteSignInPath := sitePath + f.signInPath
	var signInRequested bool
	if _, match := clPath.MatchCut(r.URL.Path, sitePath); !match {
		// site auth features only operate within their actual site's path
		return
	} else if signInRequested = reqUri == siteSignInPath; signInRequested {
		// this page is the login page
	}

	var err error
	var claims *feature.CSiteAuthClaims
	if claims, err = f.VerifyJWT(r); err != nil {
		if handled = signInRequested; handled {
			f.HandleSignInPage(w, r)
			return
		}
		//log.ErrorRF(r, "error verifying JWT and is not the sign-in page: %v", err)
		f.deleteCookie(w, f.jwtCookieName)
		f.Enjin.ServeNotFound(w, r)
		handled = true
		return
	}

	// rolling expiration optional? is this a finalize thing?
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(f.sessionDuration))

	// process post-sign-in things and finalize the request claims
	handled, modified = f.AuthorizeUserSignIn(w, r, claims)
	return
}
