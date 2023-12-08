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

package email_backup

import (
	"fmt"
	"net/http"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/golang-org-x-text/message"

	"github.com/go-enjin/be/pkg/feature"
	site_secure_context "github.com/go-enjin/be/pkg/feature/site-secure-context"
	uses_kvc "github.com/go-enjin/be/pkg/feature/uses-kvc"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/types/site"
)

const (
	SignInLinkNonceKey  = "email-backup--link"
	SignInLinkNonceName = "nonce"
	SignInFormNonceKey  = "email-backup--sign-in--form"
	SignInFormNonceName = "email-backup--sign-in--nonce"
	ManageNonceKey      = "email-backup--manage--form"
	ManageNonceName     = "email-backup--manage--nonce"
	SetupNonceKey       = "email-backup--setup--form"
	SetupNonceName      = "email-backup--setup--nonce"
	RevokeNonceKey      = "email-backup--revoke--form"
	RevokeNonceName     = "email-backup--revoke--nonce"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "site-auth-provider-email-backup"

type Feature interface {
	feature.SiteFeature
	feature.SiteAuthProvider
	feature.SiteUserSetupStage
}

type MakeFeature interface {
	feature.SiteMakeFeature[MakeFeature]
	uses_kvc.MakeFeature[MakeFeature]

	SetEmailAccount(account string) MakeFeature
	SetEmailProvider(tag feature.Tag) MakeFeature

	Make() Feature
}

type CFeature struct {
	site.CSiteFeature[MakeFeature]
	uses_kvc.CUsesKVC[MakeFeature]

	emailProviderTag feature.Tag
	emailSender      feature.EmailSender
	emailProvider    feature.EmailProvider
	emailAccount     string

	ssc *site_secure_context.CSecureContext
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.SetSiteFeatureKey("backup-email")
	f.SetSiteFeatureIcon("fa-solid fa-envelopes-bulk")
	f.SetSiteFeatureLabel(func(printer *message.Printer) (label string) {
		label = printer.Sprintf("Backup Email")
		return
	})
	f.CSiteFeature.Construct(f)
	f.ssc = site_secure_context.New(f)
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CSiteFeature.Init(this)
	f.CUsesKVC.InitUsesKVC(f)
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

func (f *CFeature) Make() (feat Feature) {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	if err = f.CSiteFeature.Build(b); err != nil {
		return
	} else if err = f.BuildUsesKVC(); err != nil {
		return
	} else if err = f.ssc.Build(b); err != nil {
		return
	}
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CSiteFeature.Startup(ctx); err != nil {
		return
	} else if err = f.CUsesKVC.StartupUsesKVC(f.Enjin.Features()); err != nil {
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

func (f *CFeature) SiteFeatureInfo(r *http.Request) (info *feature.CSiteFeatureInfo) {
	printer := lang.GetPrinterFromRequest(r)
	info = feature.NewSiteFeatureInfo(
		f.KebabTag,
		f.SiteFeatureKey(),
		f.SiteFeatureIcon(),
		f.SiteFeatureLabel(printer),
	)
	info.Backup = true
	info.Usage = printer.Sprintf("Email backup sign-ins require your primary and backup email addresses, with no passwords for you to keep track of and protect. Instead, upon sign-in, a secure one-time-use token is sent and that is used to confirm your user account.")
	info.Hint = printer.Sprintf("Request backup email token")
	info.Placeholder = printer.Sprintf("a0b1c2d3e4")
	return
}

func (f *CFeature) IsBackupProvider() (backup bool) {
	return true
}