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
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/iancoleman/strcase"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/userbase"
)

func (f *CFeature) MakeAuthClaims(aud, email string, ctx context.Context) (claims *feature.CSiteAuthClaims) {
	expiration := time.Now().Add(f.sessionDuration)
	su := f.Site().SiteUsers()
	rid := su.MakeRealID(email)
	eid := su.MakeEnjinID(rid)
	audience := strcase.ToKebab(aud)
	claims = &feature.CSiteAuthClaims{
		RID:     rid,
		EID:     eid,
		Email:   email,
		Context: ctx,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    f.Tag().Kebab(),
			Subject:   eid,
			Audience:  []string{audience},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expiration),
		},
	}
	claims.ResetFactorValueTypes()
	return
}

func (f *CFeature) GenerateJWT(claims *feature.CSiteAuthClaims) (token string, err error) {
	var audience string
	if audience = claims.GetAudience(); audience == "" {
		audience = DefaultAudience
	} else if _, present := f.audienceKeys[audience]; !present {
		audience = DefaultAudience
	}
	token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(f.audienceKeys[audience])
	return
}

func (f *CFeature) VerifyJWT(r *http.Request) (claims *feature.CSiteAuthClaims, err error) {
	var tokenCookie *http.Cookie
	if tokenCookie, err = r.Cookie(f.jwtCookieName); err != nil {
		return
	}

	var validCookie bool
	var audience, reqToken string
	if audience, reqToken, validCookie = strings.Cut(tokenCookie.Value, "="); !validCookie {
		log.ErrorRF(r, "invalid cookie value received: %q", tokenCookie.Value)
		err = errors.ErrBadCookie
		return
	} else if reqToken = forms.StrictSanitize(reqToken); reqToken == "" {
		log.ErrorRF(r, "missing token")
		err = errors.ErrTokenNotFound
		return
	} else if audience = forms.StrictCleanKebabValue(audience); audience == "" {
		log.ErrorRF(r, "missing audience")
		err = errors.ErrAudienceNotFound
		return
	} else if _, present := f.audienceKeys[audience]; !present {
		log.ErrorRF(r, "unsupported audience requested: %q; falling back to %q", audience, DefaultAudience)
		audience = DefaultAudience
	}

	claims = &feature.CSiteAuthClaims{}
	var token *jwt.Token
	if token, err = jwt.ParseWithClaims(reqToken, claims, func(token *jwt.Token) (jwtKey interface{}, err error) {
		jwtKey = f.audienceKeys[audience]
		return
	}); err != nil {
		return
	} else if token == nil || !token.Valid {
		err = errors.ErrBadRequest
		return
	}
	claims.ResetFactorValueTypes()
	return
}

func (f *CFeature) SetUserFactor(r *http.Request, claim *feature.CSiteAuthClaimsFactor) {
	claims := f.getPrivateClaims(r)
	claims.SetFactor(claim)
	return
}

func (f *CFeature) deleteCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     f.Site().SitePath(),
		MaxAge:   -1,
		Secure:   f.secureCookies,
		HttpOnly: true,
	})
}

func (f *CFeature) resetCurrentUser(w http.ResponseWriter, r *http.Request) (m *http.Request) {
	f.deleteCookie(w, f.jwtCookieName)
	m = f.setPrivateClaims(r, nil)
	m = userbase.SetCurrentAuthUser(nil, m)
	m = userbase.SetCurrentPermissions(m, f.Enjin.GetPublicAccess()...)
	return
}
