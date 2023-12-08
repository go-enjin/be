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

package app_totp

import (
	"net/http"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/golang-org-x-text/message"

	"github.com/go-enjin/be/pkg/feature"
	site_secure_context "github.com/go-enjin/be/pkg/feature/site-secure-context"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/menu"
	"github.com/go-enjin/be/types/site"
)

const (
	gSecureNewSecretKey = "new-app-totp-secret"

	SetupNonceName  = "app-totp--setup--nonce"
	SetupNonceKey   = "app-totp--setup--form"
	RevokeNonceName = "app-totp--revoke--nonce"
	RevokeNonceKey  = "app-totp--revoke--form"
	ManageNonceName = "app-totp--manage--nonce"
	ManageNonceKey  = "app-totp--manage--form"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "site-auth-otp-app-totp"

type Feature interface {
	feature.SiteFeature
	feature.SiteMultiFactorProvider
}

type MakeFeature interface {
	feature.SiteMakeFeature[MakeFeature]

	Make() Feature
}

type CFeature struct {
	site.CSiteFeature[MakeFeature]

	ssc *site_secure_context.CSecureContext
	saf feature.SiteAuthFeature
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.SetSiteFeatureKey("app-totp")
	f.SetSiteFeatureIcon("fa-solid fa-stopwatch")
	f.SetSiteFeatureLabel(func(printer *message.Printer) (label string) {
		label = printer.Sprintf("Authenticator App")
		return
	})
	f.CSiteFeature.Construct(f)
	f.ssc = site_secure_context.New(f)
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CSiteFeature.Init(this)
	f.IncludeSitePathNameFlag = false
	return
}

func (f *CFeature) Make() (feat Feature) {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	if err = f.CSiteFeature.Build(b); err != nil {
		return
	} else if err = f.ssc.Build(b); err != nil {
		return
	}
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CSiteFeature.Startup(ctx); err != nil {
		return
	} else if err = f.ssc.Startup(ctx); err != nil {
		return
	}
	return
}

func (f *CFeature) Shutdown() {
	f.CFeature.Shutdown()
}

func (f *CFeature) SetupSiteAuthProvider(saf feature.SiteAuthFeature) {
	f.saf = saf
	return
}

func (f *CFeature) UserActions() (list feature.Actions) {
	list = f.CSiteFeature.UserActions()
	return
}

func (f *CFeature) SiteFeatureMenu(r *http.Request) (m menu.Menu) {
	m = menu.Menu{
		{
			Text: f.SiteFeatureKey(),
			Href: f.SiteFeaturePath(),
			Icon: f.SiteFeatureIcon(),
		},
	}
	return
}

func (f *CFeature) IsMultiFactorBackup() (backup bool) {
	return false
}

func (f *CFeature) SiteMultiFactorKey() (key string) {
	key = f.SiteFeatureKey()
	return
}

func (f *CFeature) SiteMultiFactorLabel(printer *message.Printer) (label string) {
	label = f.SiteFeatureLabel(printer)
	return
}

func (f *CFeature) SiteFeatureInfo(r *http.Request) (info *feature.CSiteFeatureInfo) {
	printer := lang.GetPrinterFromRequest(r)
	info = feature.NewSiteFeatureInfo(
		f.KebabTag,
		f.SiteMultiFactorKey(),
		f.SiteFeatureIcon(),
		f.SiteMultiFactorLabel(printer),
	)
	info.Usage = printer.Sprintf("Authenticator app passcodes are one-time-use tokens that change every thirty (30) seconds.")
	info.Placeholder = printer.Sprintf("app passcode")
	return
}

func (f *CFeature) SiteMultiFactorInfo(r *http.Request) (info *feature.CSiteAuthMultiFactorInfo) {
	fInfo := f.SiteFeatureInfo(r)
	info = feature.NewSiteAuthMultiFactorInfo(
		fInfo.Tag,
		fInfo.Key,
		fInfo.Icon,
		fInfo.Label,
		f.CurrentUserFactorsReady(r)...,
	)
	info.Usage = fInfo.Usage
	info.Placeholder = fInfo.Placeholder
	return
}

func (f *CFeature) CurrentUserFactorsReady(r *http.Request) (names []string) {
	names = f.listSecureProvisions(r)
	return
}

func (f *CFeature) CurrentUserFactorsReadyCount(r *http.Request) (count int) {
	count = f.countSecureProvisions(r)
	return
}

func (f *CFeature) ResetUserFactors(r *http.Request, eid string) (err error) {
	err = f.ssc.Delete(eid, r, f.Site().SiteUsers())
	return
}