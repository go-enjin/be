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

package feature

import (
	"net/http"
	"time"

	"github.com/go-enjin/be/pkg/context"
)

type SiteAuthFeature interface {
	Feature

	Site() Site
	SiteFeatureKey() (name string)
	SiteFeaturePath() (path string)
	SiteFeatureTheme() (t Theme)

	SiteAuthSignInPath() (url string)
	SiteAuthSignOutPath() (url string)
	SiteAuthChallengePath() (path string)

	NumFactorsPresent() (count int)
	NumFactorsRequired() (count int)

	IsUserAllowed(email string) (allowed bool)
	GetSignUpsAllowed() (allowed bool)
	GetSessionDuration() (duration time.Duration)
	GetVerifiedDuration() (duration time.Duration)

	MakeAuthClaims(aud, email string, ctx context.Context) (claims *CSiteAuthClaims)
	GenerateJWT(claims *CSiteAuthClaims) (token string, err error)
	VerifyJWT(r *http.Request) (claims *CSiteAuthClaims, err error)
	ResetUserFactors(r *http.Request, eid string) (err error)
	SetUserFactor(r *http.Request, claim *CSiteAuthClaimsFactor)

	AuthorizeUserSignIn(w http.ResponseWriter, r *http.Request, claims *CSiteAuthClaims) (handled bool, modified *http.Request)

	SiteAuthRequestHandler

	ServeSignInPage(w http.ResponseWriter, r *http.Request)
	ServeSignOutPage(w http.ResponseWriter, r *http.Request)
	ServeChallengeRequest(w http.ResponseWriter, r *http.Request)

	ServeVerificationRequest(verifyTarget string, w http.ResponseWriter, r *http.Request) (handled bool, redirect string)
	RequireVerification(path string, w http.ResponseWriter, r *http.Request) (handled bool, redirect string)
}