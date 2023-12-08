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

package email_token

import (
	"fmt"
	"net/http"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/golang-org-x-text/message"

	"github.com/go-enjin/be/pkg/feature"
	uses_kvc "github.com/go-enjin/be/pkg/feature/uses-kvc"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/types/site"
)

const (
	SignInFormNonceKey  = "email-token--sign-in--form"
	SignInFormNonceName = "email-token--sign-in--form--nonce"
	SignInLinkNonceKey  = "email-token--sign-in--link"
	SignInLinkNonceName = "nonce"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "site-auth-provider-email-token"

type Feature interface {
	feature.SiteFeature
	feature.SiteAuthProvider
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

	emailSender   feature.EmailSender
	emailProvider feature.EmailProvider
	emailAccount  string

	redirectOnSignIn string
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.SetSiteFeatureKey("email")
	f.SetSiteFeatureIcon("fa-solid fa-envelope")
	f.SetSiteFeatureLabel(func(printer *message.Printer) (label string) {
		label = printer.Sprintf("Email")
		return
	})
	f.CSiteFeature.Construct(f)
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
	}
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CSiteFeature.Startup(ctx); err != nil {
		return
	} else if err = f.CUsesKVC.StartupUsesKVC(f.Enjin.Features()); err != nil {
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
	info.Usage = printer.Sprintf("Email sign-ins require just an email address with no passwords for you to keep track of and protect. Instead, upon sign-in, a secure one-time-use token is sent and that is used to confirm your user account.")
	info.Hint = printer.Sprintf("Request email token")
	info.Placeholder = printer.Sprintf("a0b1c2d3e4")
	return
}