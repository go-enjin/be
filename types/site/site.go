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

package site

import (
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/feature/signaling"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/menu"
	bePath "github.com/go-enjin/be/pkg/path"
)

const (
	BaseTag feature.Tag = "site-feature"
)

var (
	_ feature.SiteFeature = (*CSiteFeature[feature.SiteMakeFeature[feature.SiteFeature]])(nil)
	//_ feature.SiteMakeFeature[feature.SiteFeature] = (*CSiteFeature[feature.SiteMakeFeature[feature.SiteFeature]])(nil)
)

type CSiteFeature[MakeTypedFeature interface{}] struct {
	feature.CFeature
	signaling.CSignaling

	site feature.Site

	include      feature.Features
	sitePathName string
}

func (f *CSiteFeature[MakeTypedFeature]) Site() (s feature.Site) {
	s = f.site
	return
}

func (f *CSiteFeature[MakeTypedFeature]) Init(this interface{}) {
	f.CFeature.Init(this)
	f.CFeature.PackageTag = BaseTag
	f.CSignaling.InitSignaling()
	return
}

func (f *CSiteFeature[MakeTypedFeature]) Include(features ...feature.Feature) MakeTypedFeature {
	f.include = append(f.include, features...)
	t, _ := f.This().(MakeTypedFeature)
	return t
}

func (f *CSiteFeature[MakeTypedFeature]) SetSiteFeaturePathName(name string) MakeTypedFeature {
	f.sitePathName = strcase.ToKebab(name)
	t, _ := f.This().(MakeTypedFeature)
	return t
}

func (f *CSiteFeature[MakeTypedFeature]) Build(b feature.Buildable) (err error) {
	if err = f.CFeature.Build(b); err != nil {
		return
	}
	for _, ef := range f.include {
		b.AddFeature(ef)
	}
	category := f.Tag().String()
	prefix := f.Tag().Kebab()
	b.AddFlags(&cli.StringFlag{
		Name:     prefix + "-path-name",
		Usage:    "specify the URL path name for this site feature",
		Category: category,
		Value:    f.sitePathName,
	})
	return
}

func (f *CSiteFeature[MakeTypedFeature]) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}

	prefix := f.Tag().Kebab()
	pathKey := prefix + "-path-name"
	if ctx.IsSet(pathKey) {
		if v := ctx.String(pathKey); v != "" {
			if v = forms.StrictSanitize(strings.TrimSpace(v)); v != "" {
				f.sitePathName = bePath.TrimSlashes(strcase.ToKebab(v))
			}
		}
	}
	log.InfoF("%v site feature path name: %v", f.Tag(), f.sitePathName)

	return
}

func (f *CSiteFeature[MakeTypedFeature]) Shutdown() {
	f.CFeature.Shutdown()
}

func (f *CSiteFeature[MakeTypedFeature]) SetupSiteFeature(s feature.Site) {
	f.site = s
	return
}

func (f *CSiteFeature[MakeTypedFeature]) RouteSiteFeature(r chi.Router) {
	log.FatalF("%v.RouteSiteFeature method unimplemented", f.Tag())
}

func (f *CSiteFeature[MakeTypedFeature]) SiteFeaturePathName() (name string) {
	name = f.sitePathName
	return
}

func (f *CSiteFeature[MakeTypedFeature]) SiteFeaturePath() (path string) {
	var sitePath string
	if v := f.site.SitePath(); v != "/" {
		sitePath = v
	}
	path = sitePath + "/" + f.sitePathName
	return
}

func (f *CSiteFeature[MakeTypedFeature]) SiteFeatureMenu() (m menu.Menu) {
	return
}