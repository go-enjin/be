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

package user_manager

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/urfave/cli/v2"

	"github.com/go-corelibs/x-text/message"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/menu"
	"github.com/go-enjin/be/types/site"
)

const (
	UserManagerNonceName = "user-manager-nonce"
	UserManagerNonceKey  = "user-manager-form"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "site-user-manager"

type Feature interface {
	feature.SiteFeature
}

type MakeFeature interface {
	feature.SiteMakeFeature[MakeFeature]

	Make() Feature
}

type CFeature struct {
	site.CSiteFeature[MakeFeature]
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.SetSiteFeatureKey("user-manager")
	f.SetSiteFeatureIcon("fa-solid fa-users")
	f.SetSiteFeatureLabel(func(printer *message.Printer) (label string) {
		label = printer.Sprintf("User Manager")
		return
	})
	f.CSiteFeature.Construct(f)
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CSiteFeature.Init(this)
	return
}

func (f *CFeature) Make() (feat Feature) {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	if err = f.CSiteFeature.Build(b); err != nil {
		return
	}
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CSiteFeature.Startup(ctx); err != nil {
		return
	}
	return
}

func (f *CFeature) SetupSiteFeature(s feature.Site) (err error) {
	if err = f.CSiteFeature.SetupSiteFeature(s); err != nil {
		return
	}
	if f.Site().SiteUsers() == nil {
		err = fmt.Errorf("%q feature requires a site with .SiteUsers configured", f.Tag())
		return
	}
	if f.Site().SiteAuth() == nil {
		err = fmt.Errorf("%q feature requires a site with authentication configured", f.Tag())
		return
	}
	return
}

func (f *CFeature) UserActions() (actions feature.Actions) {
	actions = feature.Actions{
		f.Action("access", "feature"),
		f.Action("create", "user"),
		f.Action("update", "user"),
		f.Action("delete", "user"),
	}
	return
}

func (f *CFeature) SiteFeatureMenu(r *http.Request) (m menu.Menu) {
	info := f.SiteFeatureInfo(r)
	m = menu.Menu{{
		Text: info.Label,
		Href: f.SiteFeaturePath(),
		Icon: info.Icon,
	}}
	return
}

func (f *CFeature) RouteSiteFeature(r chi.Router) {
	r.Post("/*", f.HandleUserManager)
	r.Get("/*", f.RenderUserManager)
}
