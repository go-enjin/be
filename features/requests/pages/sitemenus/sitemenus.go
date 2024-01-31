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

package sitemenus

import (
	"net/http"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"

	"github.com/go-corelibs/x-text/message"
	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/menu"
	bePath "github.com/go-enjin/be/pkg/path"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "requests-pages-sitemenus"

type Feature interface {
	feature.Feature
	feature.PrepareServePagesFeature
}

type MakeFeature interface {
	Make() Feature
}

type CFeature struct {
	feature.CFeature
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.CFeature.Construct(f)
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	return
}

func (f *CFeature) Make() (feat Feature) {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	if err = f.CFeature.Build(b); err != nil {
		return
	}
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	return
}

func (f *CFeature) Shutdown() {
	f.CFeature.Shutdown()
}

func ProcessActiveItems(m menu.Menu, r *http.Request) (modified menu.Menu, found bool) {
	var path string
	if path = bePath.TrimSlashes(r.URL.Path); path == "" {
		modified = m
		return
	}

	isActive := func(href string) (active bool) {
		trimmed := bePath.TrimSlashes(href)
		active = path == trimmed || strings.HasPrefix(path, trimmed+"/")
		return
	}

	for _, mm := range m {
		var active bool
		if mm.SubMenu, active = ProcessActiveItems(mm.SubMenu, r); active {
			mm.Active = true
		} else {
			mm.Active = isActive(mm.Href)
		}
		if !found && mm.Active {
			found = true
		}
		modified = append(modified, mm)
	}

	return
}

func (f *CFeature) PrepareServePage(ctx beContext.Context, t feature.Theme, p feature.Page, w http.ResponseWriter, r *http.Request) (out beContext.Context, modified *http.Request, stop bool) {
	reqLangTag := message.GetTag(r)

	var siteMenu map[string]interface{}
	if v := ctx.Get("SiteMenu"); v != nil {
		if vm, ok := v.(beContext.Context); ok {
			siteMenu = vm
		} else {
			log.ErrorRF(r, "invalid .SiteMenu value type: (%T) %#+v", v, v)
		}
	}

	allMenus := make(map[string]interface{})
	for _, mp := range f.Enjin.GetMenuProviders() {
		for name, m := range mp.GetMenus(reqLangTag) {
			camel := strcase.ToCamel(name)
			if v, ok := siteMenu[camel]; ok {
				allMenus[camel] = v
				log.DebugRF(r, "retaining [%v] menu: %v (.SiteMenu.%v)", reqLangTag.String(), name, camel)
			} else {
				allMenus[camel] = m
				log.DebugRF(r, "providing [%v] menu: %v (.SiteMenu.%v)", reqLangTag.String(), name, camel)
			}
		}
	}

	if len(allMenus) > 0 {
		if v, present := allMenus["MainMenu"]; present {
			if vt, ok := v.(menu.Menu); ok {
				allMenus["MainMenu"], _ = ProcessActiveItems(vt, r)
			}
		}
		ctx.SetSpecific("SiteMenu", allMenus)
	}

	out = ctx
	return
}
