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
	"time"
)

var (
	DefaultSignInPath    = "/sign-in"
	DefaultSignOutPath   = "/sign-out"
	DefaultSettingsPath  = "/settings"
	DefaultChallengePath = "/mfa"

	DefaultJwtCookieName  = "enjin-site-jwt-cookie"
	DefaultXsrfCookieName = "enjin-site-xsrf-Cookie"
	DefaultXsrfHeaderName = "enjin-site-xsrf-header"

	DefaultRequiredFactors  = 0
	DefaultSessionDuration  = time.Hour
	DefaultVerifiedDuration = time.Minute * 10

	DefaultDeleteOwnUserConfirmation = "Please delete my own user, thank you!"
)