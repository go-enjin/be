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

	"github.com/go-corelibs/x-text/message"
	"github.com/go-enjin/be/pkg/feature"
	bePath "github.com/go-enjin/be/pkg/path"
)

func (f *CFeature) enforceUserSetupStages(claims *feature.CSiteAuthClaims, w http.ResponseWriter, r *http.Request) (handled bool) {

	printer := message.GetPrinter(r)

	/*
			- get a list of all features with account setup stages
			- all primary backup factors are required to be setup
			- configured number of distinct otp factors must be setup
			- the process for performing the steps loops through the above (in the same order) until all features report ready

		- total is:
			- backup accounts
			- backup multi-factors
			- required multi-factors
			- other site features
	*/

	var allUserSetupStages []feature.SiteUserSetupStage
	for _, sf := range f.Site().SiteFeatures() {
		if usf, ok := sf.This().(feature.SiteUserSetupStage); ok {
			allUserSetupStages = append(allUserSetupStages, usf)
		}
	}

	numProviderBackups := f.sab.Features.Len()
	numMultiFactorBackups := f.mfb.Features.Len()
	mustBackupFactors := f.mustBackupFactors

	if mustBackupFactors.Len() == 0 && numMultiFactorBackups > 0 {
		// if there are no backups explicitly required and the user has at least one factor configured, then all mfa
		// backups are required
		var required bool
		for _, mfp := range f.mfa.Features {
			if required = mfp.CurrentUserFactorsReadyCount(r) > 0; required {
				break
			}
		}
		if required {
			mustBackupFactors = f.mfb.Features.Tags()
		}
	}

	var totalStages, currentStage int
	totalStages += f.mustBackupAccounts.Len() // required account backups
	totalStages += mustBackupFactors.Len()    // required multi-factor backups
	totalStages += f.numFactorsRequired       // configured number of standard factors are required
	totalStages += len(allUserSetupStages)    // all site features are required

	currentStage += 0 // starts at one

	// provider backups...
	var numProviderBackupsReady int
	if f.mustBackupAccounts.Len() > 0 {
		for _, sab := range f.sab.Features {
			if sab.SiteUserSetupStageReady(claims.EID, r) {
				numProviderBackupsReady += 1
			}
		}
		currentStage += numProviderBackupsReady
	}

	// mfa backups...
	var numMultiFactorBackupsReady int
	if mustBackupFactors.Len() > 0 {
		for _, mfb := range f.mfb.Features {
			if mfb.SiteUserSetupStageReady(claims.EID, r) {
				numMultiFactorBackupsReady += 1
			}
		}
		currentStage += numMultiFactorBackupsReady
	}

	// mfa features...
	var numFactorsRequiredReady int
	var canSetupFactors []feature.SiteMultiFactorProvider
	if f.numFactorsRequired > 0 {
		for _, mfp := range f.mfa.Features {
			if !mfp.IsMultiFactorBackup() {
				if mfp.SiteUserSetupStageReady(claims.EID, r) {
					numFactorsRequiredReady += 1
				} else {
					canSetupFactors = append(canSetupFactors, mfp)
				}
			}
		}
		currentStage += numFactorsRequiredReady
	}

	// site features...
	var numUserSetupStagesReady int
	for _, sf := range allUserSetupStages {
		if sf.SiteUserSetupStageReady(claims.EID, r) {
			numUserSetupStagesReady += 1
		}
	}
	currentStage += numUserSetupStagesReady

	if currentStage > totalStages {
		// nothing further to do
		return
	}

	makeSummaryNotice := func() (text string, argv []interface{}, dismiss bool) {

		text = printer.Sprintf(`Account Setup (Stage %[1]d of %[2]d)`, currentStage+1, totalStages)

		if f.mustBackupAccounts.Len() > numProviderBackupsReady {
			argv = append(argv, printer.Sprintf("%[1]d backup accounts", numProviderBackups-numProviderBackupsReady))
		}
		if mustBackupFactors.Len() > numMultiFactorBackupsReady {
			argv = append(argv, printer.Sprintf("%[1]d backup multi-factors", numMultiFactorBackups-numMultiFactorBackupsReady))
		}
		if f.numFactorsRequired > numFactorsRequiredReady {
			argv = append(argv, printer.Sprintf("%[1]d normal multi-factors", f.numFactorsRequired-numFactorsRequiredReady))
		}
		if count := len(allUserSetupStages); count > numUserSetupStagesReady {
			argv = append(argv, printer.Sprintf("%[1]d other site features", count-numUserSetupStagesReady))
		}

		if dismiss = len(argv) == 0; dismiss {
			argv = append(argv, printer.Sprintf("Your account setup is complete!"))
		}

		return
	}

	summaryText, argv, dismissNotice := makeSummaryNotice()
	r = feature.AddImportantNotice(r, dismissNotice, summaryText, argv...)

	siteAuthUrl := f.SiteAuthSignInPath()

	if suffix, ok := bePath.MatchCut(r.URL.Path, siteAuthUrl); ok && suffix != "" {
		// this isn't the top sign-in path, which means we're handling mfa setup
		if f.numFactorsRequired > numFactorsRequiredReady {
			if tagPath := bePath.TopPathName(suffix); tagPath != "" {
				// serving a specific sub-path
				tag := f.mfa.Features.Find(tagPath)
				if handled = tag.IsNil(); handled {
					// not an actual feature path
					f.Enjin.ServeRedirect(siteAuthUrl, w, r)
					return
				}
				mfp := f.mfa.Features.Get(tag)
				if handled = mfp.SiteUserSetupStageReady(claims.EID, r); handled {
					// feature is now ready, redirect back to top to continue stages
					f.Enjin.ServeRedirect(siteAuthUrl, w, r)
					return
				}
				handled = true
				mfp.SiteUserSetupStageHandler(f, w, r)
				return
			}
		}

	} else if handled = !ok && f.numFactorsRequired > numFactorsRequiredReady; handled {
		// user is not on the sign-in path and not enough factors have been configured yet
		// must enforce setup URL for correct handling because the path their on may not accept POST requests
		claims.Context.SetSpecific("post-setup-redirect", r.URL.Path)
		f.Enjin.ServeRedirect(siteAuthUrl, w, r)
		return
	}

	// setup backup account features
	for _, tag := range f.mustBackupAccounts {
		sab := f.sab.Features.Get(tag)
		if handled = !sab.SiteUserSetupStageReady(claims.EID, r); handled {
			// add notice with progress
			sab.SiteUserSetupStageHandler(f, w, r)
			return
		}
	}

	// setup backup factors
	for _, tag := range mustBackupFactors {
		mfb := f.mfb.Features.Get(tag)
		if handled = !mfb.SiteUserSetupStageReady(claims.EID, r); handled {
			// add notice with progress
			mfb.SiteUserSetupStageHandler(f, w, r)
			return
		}
	}

	// setup standard factors
	if f.numFactorsRequired > numFactorsRequiredReady {
		if handled = f.NumFactorsPresent() != f.numFactorsRequired; handled {
			// display the mfa setup selector page
			f.ServeSettingsPanelSetupSelectorPage(r.URL.Path, w, r)
			return
		}
		// the number of required factors is the same as the total number of multifactor features, no need to
		// display the mfa selector
		for _, mfp := range canSetupFactors {
			if handled = !mfp.IsMultiFactorBackup(); handled {
				// process one at a time, in the order added during the enjin build phase
				mfp.SiteUserSetupStageHandler(f, w, r)
				return
			}
		}
	}

	// setup other site features
	for _, usf := range allUserSetupStages {
		if !usf.SiteUserSetupStageReady(claims.EID, r) {
			//usf.SiteUserSetupStageHandler(f, w, addSummaryNotice(r))
			usf.SiteUserSetupStageHandler(f, w, r)
			handled = true
			return
		}
	}

	r = feature.FilterUserNotices(r, func(notice *feature.UserNotice) (keep bool) {
		keep = notice.Summary != summaryText
		return
	})

	// check for post-setup redirection
	redirect := claims.Context.String("post-setup-redirect", "")
	if handled = redirect != ""; handled {
		claims.Context.DeleteKeys("post-setup-redirect")
		f.Enjin.ServeRedirect(redirect, w, r)
		return
	}

	return
}
