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

const (
	DefaultAudience    = "default"
	MinSessionDuration = time.Minute

	SignInNonceKey        = "sign-in--form"
	SignInNonceName       = "sign-in--nonce"
	SignOutNonceKey       = "sign-out--form"
	SignOutNonceName      = "sign-out--nonce"
	ChallengeNonceKey     = "otp--challenge--form"
	ChallengeNonceName    = "otp--challenge--nonce"
	VerificationNonceKey  = "otp--verification--form"
	VerificationNonceName = "otp--verification--nonce"

	SettingsNonceKey  = "settings--form"
	SettingsNonceName = "settings--nonce"
)

const (
	gRedirectKey        = "redirect"
	gVerifyTargetKey    = "verify-target"
	gVerifyingTargetKey = "verifying-target"
)
