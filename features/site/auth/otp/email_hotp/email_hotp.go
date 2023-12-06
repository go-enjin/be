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

package email_hotp

import (
	"fmt"
	"net/http"

	"github.com/urfave/cli/v2"
	"github.com/xlzd/gotp"

	"github.com/go-enjin/golang-org-x-text/message"

	"github.com/go-enjin/be/pkg/feature"
	site_secure_context "github.com/go-enjin/be/pkg/feature/site-secure-context"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/menu"
	"github.com/go-enjin/be/types/site"
)

const (
	gSecureNewSecretKey = "new-email-hotp-secret"

	SetupNonceName  = "email-hotp--setup--nonce"
	SetupNonceKey   = "email-hotp--setup--form"
	RevokeNonceName = "email-hotp--revoke--nonce"
	RevokeNonceKey  = "email-hotp--revoke--form"
	ManageNonceName = "email-hotp--manage--nonce"
	ManageNonceKey  = "email-hotp--manage--form"
)

var (
	DefaultHotpDigits = 8
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "site-auth-otp-email-hotp"

type Feature interface {
	feature.SiteFeature
	feature.SiteMultiFactorProvider
}

type MakeFeature interface {
	feature.SiteMakeFeature[MakeFeature]

	SetEmailAccount(account string) MakeFeature
	SetEmailProvider(tag feature.Tag) MakeFeature
	SetHotpDigits(count int) MakeFeature

	Make() Feature
}

type CFeature struct {
	site.CSiteFeature[MakeFeature]

	ssc *site_secure_context.CSecureContext
	saf feature.SiteAuthFeature

	emailProviderTag feature.Tag
	emailSender      feature.EmailSender
	emailProvider    feature.EmailProvider
	emailAccount     string

	hotpDigits int
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.SetSiteFeatureKey("email-hotp")
	f.SetSiteFeatureIcon("fa-solid fa-envelope-circle-check")
	f.SetSiteFeatureLabel(func(printer *message.Printer) (label string) {
		label = printer.Sprintf("Email Passcode")
		return
	})
	f.CSiteFeature.Construct(f)
	f.ssc = site_secure_context.New(f)
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CSiteFeature.Init(this)
	f.IncludeSitePathNameFlag = false
	f.hotpDigits = DefaultHotpDigits
	return
}

func (f *CFeature) SetEmailAccount(account string) MakeFeature {
	f.emailAccount = account
	return f
}

func (f *CFeature) SetEmailProvider(tag feature.Tag) MakeFeature {
	f.emailProviderTag = tag
	return f
}

func (f *CFeature) SetHotpDigits(count int) MakeFeature {
	if count < 6 {
		log.FatalDF(1, "the number of HOTP digits must be greater than 5")
	}
	f.hotpDigits = count
	return f
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

	if f.emailAccount == "" {
		err = fmt.Errorf(".SetEmailAccount is required")
		return
	} else if f.emailSender = f.Enjin.FindEmailAccount(f.emailAccount); f.emailSender == nil {
		err = fmt.Errorf("%v email sender not found", f.emailAccount)
	} else if f.emailProviderTag.IsNil() {
		err = fmt.Errorf(".SetEmailProvider is required")
		return
	} else if epf, ok := f.Enjin.Features().Get(f.emailProviderTag); !ok {
		err = fmt.Errorf("%v email provider feature not found", f.emailProviderTag)
		return
	} else if ep, ok := epf.This().(feature.EmailProvider); !ok {
		err = fmt.Errorf("%v feature is not a feature.EmailProvider", f.emailProviderTag)
		return
	} else {
		f.emailProvider = ep
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
		f.Tag().Kebab(),
		f.SiteMultiFactorKey(),
		f.SiteFeatureIcon(),
		f.SiteMultiFactorLabel(printer),
	)
	info.Hint = printer.Sprintf("Request passcode email")
	info.Usage = printer.Sprintf("Email passcodes are similar to normal authenticator app passcodes except that these passcodes are sent to your email address.")
	info.Placeholder = printer.Sprintf("email passcode")
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
	info.Submit = true
	info.Hint = fInfo.Hint
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

func (f *CFeature) makeHotp(secret string) (hotp *gotp.HOTP) {
	hotp = gotp.NewHOTP(secret, f.hotpDigits, nil)
	return
}