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

package signals

import (
	"net/http"

	"github.com/go-enjin/be/pkg/feature/signaling"
)

const (
	PreNewEnjin          signaling.Signal = "pre-new-enjin"
	PostNewEnjin         signaling.Signal = "post-new-enjin"
	PreNewEnjinIncluded  signaling.Signal = "pre-new-enjin-included"
	PostNewEnjinIncluded signaling.Signal = "post-new-enjin-included"

	RootEnjinSetup    signaling.Signal = "root-enjin-setup"
	RootEnjinStarting signaling.Signal = "root-enjin-starting"
	RootEnjinShutdown signaling.Signal = "root-enjin-shutdown"

	PreEnjinSetupRouter  signaling.Signal = "pre-enjin-setup-router"
	PostEnjinSetupRouter signaling.Signal = "post-enjin-setup-router"

	PreInitConsoleFeatures   signaling.Signal = "pre-init-console-features"
	PostInitConsoleFeatures  signaling.Signal = "post-init-console-features"
	PreEnjinIntegrityChecks  signaling.Signal = "pre-enjin-integrity-checks"
	PostEnjinIntegrityChecks signaling.Signal = "post-enjin-integrity-checks"

	PreEnjinReloadLocales  signaling.Signal = "pre-enjin-reload-locales"
	PostEnjinReloadLocales signaling.Signal = "post-enjin-reload-locales"

	PreHotReloadFeatures  signaling.Signal = "pre-hot-reload-features"
	PostHotReloadFeatures signaling.Signal = "post-hot-reload-features"

	PreBuildFeaturesPhase     signaling.Signal = "pre-enjin-build-features-phase"
	PostBuildFeaturesPhase    signaling.Signal = "post-enjin-build-features-phase"
	PreSetupFeaturesPhase     signaling.Signal = "pre-enjin-setup-features-phase"
	PostSetupFeaturesPhase    signaling.Signal = "post-enjin-setup-features-phase"
	PreStartupFeaturesPhase   signaling.Signal = "pre-enjin-startup-features-phase"
	PostStartupFeaturesPhase  signaling.Signal = "post-enjin-startup-features-phase"
	PreShutdownFeaturesPhase  signaling.Signal = "pre-enjin-shutdown-features-phase"
	PostShutdownFeaturesPhase signaling.Signal = "post-enjin-shutdown-features-phase"

	ServedHttpRedirect signaling.Signal = "served-http-redirect"
	ServedHtmlRedirect signaling.Signal = "served-meta-redirect"

	Served204      signaling.Signal = "served-204"
	Served400      signaling.Signal = "served-400"
	Served401      signaling.Signal = "served-401"
	ServedBasic401 signaling.Signal = "served-basic-401"
	Served403      signaling.Signal = "served-403"
	Served404      signaling.Signal = "served-404"
	Served405      signaling.Signal = "served-405"
	Served500      signaling.Signal = "served-500"

	PreServeStatusPage  signaling.Signal = "pre-serve-status-page"
	PostServeStatusPage signaling.Signal = "post-serve-status-page"
	ServedPath          signaling.Signal = "served-path"
	ServedPathPage      signaling.Signal = "served-path-page"
	PreServePage        signaling.Signal = "pre-serve-page"
	PostServePage       signaling.Signal = "post-serve-page"
	PreServeData        signaling.Signal = "pre-serve-data"
	PostServeData       signaling.Signal = "post-serve-data"

	EnjinPanicRecovery          signaling.Signal = "enjin-panic-recovery"
	EnjinSecondaryPanicRecovery signaling.Signal = "enjin-secondary-panic-recovery"
)

func UnpackEnjinPanicRecovery(argv []interface{}) (w http.ResponseWriter, r *http.Request, err error, ok bool) {
	if ok = len(argv) == 3; ok {
		if w, ok = argv[0].(http.ResponseWriter); ok {
			if r, ok = argv[1].(*http.Request); ok {
				if err, ok = argv[2].(error); ok {
					return
				}
			}
		}
	}
	return
}

func UnpackEnjinSecondaryPanicRecovery(argv []interface{}) (w http.ResponseWriter, r *http.Request, err, ee error, ok bool) {
	if ok = len(argv) == 4; ok {
		if w, ok = argv[0].(http.ResponseWriter); ok {
			if r, ok = argv[1].(*http.Request); ok {
				if err, ok = argv[2].(error); ok {
					if ee, ok = argv[3].(error); ok {
						return
					}
				}
			}
		}
	}
	return
}