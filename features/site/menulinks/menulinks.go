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

package menulinks

import (
	"net/http"

	"github.com/urfave/cli/v2"

	"github.com/go-corelibs/x-text/message"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/menu"
	"github.com/go-enjin/be/pkg/userbase"
	"github.com/go-enjin/be/types/site"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "site-menulinks"

type MenuItemFn func(printer *message.Printer) (item *menu.Item, permissions feature.Actions)

type Feature interface {
	feature.SiteFeature
}

type MakeFeature interface {
	feature.SiteMakeFeature[MakeFeature]

	AddItemFuncs(fns ...MenuItemFn) MakeFeature

	Make() Feature
}

type CFeature struct {
	site.CSiteFeature[MakeFeature]

	menuItemFns []MenuItemFn
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.SetSiteFeatureKey("menulinks")
	f.SetSiteFeatureIcon("fa-solid fa-paperclip")
	f.SetSiteFeatureLabel(func(printer *message.Printer) (label string) {
		label = printer.Sprintf("extra menus and top-level menu items")
		return
	})
	f.CSiteFeature.Construct(f)
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CSiteFeature.Init(this)
	return
}

func (f *CFeature) AddItemFuncs(fns ...MenuItemFn) MakeFeature {
	f.menuItemFns = append(f.menuItemFns, fns...)
	return f
}

func (f *CFeature) Make() (feat Feature) {
	return f
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CSiteFeature.Startup(ctx); err != nil {
		return
	}
	return
}

func (f *CFeature) SiteFeatureMenu(r *http.Request) (m menu.Menu) {
	printer := message.GetPrinter(r)
	for _, fn := range f.menuItemFns {
		if item, permissions := fn(printer); item != nil {
			if len(permissions) == 0 || userbase.CurrentUserCan(r, permissions...) {
				m = append(m, item)
			}
		}
	}
	return
}
