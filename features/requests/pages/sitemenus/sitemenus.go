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

	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
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

func (f *CFeature) PrepareServePage(ctx beContext.Context, t feature.Theme, p feature.Page, w http.ResponseWriter, r *http.Request) (out beContext.Context, modified *http.Request, stop bool) {
	reqLangTag := lang.GetTag(r)

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
		ctx.SetSpecific("SiteMenu", allMenus)
	}

	out = ctx
	return
}