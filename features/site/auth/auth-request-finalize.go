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

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
)

func (f *CFeature) FinalizeSiteRequest(w http.ResponseWriter, r *http.Request) (m *http.Request) {
	var au feature.AuthUser
	var claims *feature.CSiteAuthClaims

	if claims = f.getPrivateClaims(r); claims == nil {
		m = f.resetCurrentUser(w, r)
		return
	}

	var err error
	f.Site().SiteUsers().LockUser(r, claims.EID)
	defer f.Site().SiteUsers().UnlockUser(r, claims.EID)

	if au, err = f.Site().SiteUsers().RetrieveUser(r, claims.EID); err != nil {
		log.ErrorRF(r, "error retrieving user %q: %v", claims.EID, err)
		m = f.resetCurrentUser(w, r)
		return
	} else if au.IsVisitor() {
		m = f.resetCurrentUser(w, r)
		return
	}

	auCtx := au.UnsafeContext()
	_ = auCtx.SetKV(".last-seen", time.Now())

	if err = f.Site().SiteUsers().SetUserContext(r, claims.EID, auCtx); err != nil {
		log.ErrorRF(r, "error setting user context: %v", err)
		m = f.resetCurrentUser(w, r)
		return
	}

	var token string
	if token, err = f.GenerateJWT(claims); err != nil {
		log.ErrorRF(r, "error getting site auth token string: %v", err)
		m = f.resetCurrentUser(w, r)
		return
	}

	var audience string
	if audience = claims.GetAudience(); audience == "" {
		audience = DefaultAudience
	}

	cookie := &http.Cookie{
		Name:     f.jwtCookieName,
		Value:    audience + "=" + token,
		Path:     f.Site().SitePath(),
		Secure:   f.secureCookies,
		HttpOnly: true,
		SameSite: f.sameSiteCookies,
	}
	cookie.Expires = time.Now().Add(f.sessionDuration)

	http.SetCookie(w, cookie)

	return
}
