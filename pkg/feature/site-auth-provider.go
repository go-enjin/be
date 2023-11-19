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

	"github.com/go-enjin/golang-org-x-text/message"
)

type SiteAuthProvider interface {
	SiteFeature

	IsBackupProvider() (backup bool)

	SiteAuthLoginCallback(w http.ResponseWriter, r *http.Request, saf SiteAuthFeature) (err error)
	SiteAuthSignInHandler(w http.ResponseWriter, r *http.Request, saf SiteAuthFeature) (claims *CSiteAuthClaims, redirect string, err error)
	SiteAuthSignOutHandler(w http.ResponseWriter, r *http.Request, saf SiteAuthFeature) (handled bool, redirect string, err error)
}

type SiteUserSetupStage interface {
	SiteFeature

	SiteUserSetupStageReady(eid string, r *http.Request) (ready bool)
	SiteUserSetupStageHandler(saf SiteAuthFeature, w http.ResponseWriter, r *http.Request)
}

type SiteMultiFactorProvider interface {
	SiteFeature
	SiteUserSetupStage

	SetupSiteAuthProvider(saf SiteAuthFeature)
	IsMultiFactorBackup() (backup bool)

	SiteMultiFactorKey() (key string)
	SiteMultiFactorInfo(r *http.Request) (info *CSiteAuthMultiFactorInfo)
	SiteMultiFactorLabel(printer *message.Printer) (label string)
	CurrentUserFactorsReady(r *http.Request) (names []string)
	CurrentUserFactorsReadyCount(r *http.Request) (count int)

	ResetUserFactors(r *http.Request, eid string) (err error)

	ProcessSetupPage(saf SiteAuthFeature, w http.ResponseWriter, r *http.Request) (handled bool)

	ProcessChallenge(factor, challenge string, saf SiteAuthFeature, claims *CSiteAuthClaims, w http.ResponseWriter, r *http.Request) (handled bool, redirect string)

	VerifyClaimFactor(claim *CSiteAuthClaimsFactor, saf SiteAuthFeature, r *http.Request) (verified bool)
	ProcessVerification(verifyTarget, factor, challenge string, saf SiteAuthFeature, claims *CSiteAuthClaims, w http.ResponseWriter, r *http.Request) (handled bool, redirect string)
}

type SiteAuthSettingsPanel interface {
	SiteFeature

	SiteAuthSettingsPanel(settingsPath string, saf SiteAuthFeature) (serve, handle http.HandlerFunc)
}