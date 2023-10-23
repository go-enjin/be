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
	_ feature.SiteFeature = (*CSiteFeature[feature.SiteFeature, feature.SiteMakeFeature[feature.MakeFeature]])(nil)
	//_ feature.SiteMakeFeature[feature.SiteFeature] = (*CSiteFeature[feature.SiteMakeFeature[feature.SiteFeature]])(nil)
)

type CSiteFeature[T interface{}, M interface{}] struct {
	feature.CFeature
	signaling.CSignaling
	feature.CSiteIncluding[T, M]

	site feature.Site

	sitePathName string

	themeName string
	theme     feature.Theme
}

func (f *CSiteFeature[T, M]) Init(this interface{}) {
	f.CFeature.Init(this)
	f.CFeature.PackageTag = BaseTag
	f.CSignaling.InitSignaling()
	f.CSiteIncluding.InitSiteIncluding(this)
	return
}

func (f *CSiteFeature[T, M]) SetSiteFeatureName(name string) M {
	f.sitePathName = strcase.ToKebab(name)
	t, _ := f.This().(M)
	return t
}

func (f *CSiteFeature[T, M]) SetSiteFeatureTheme(name string) M {
	f.themeName = name
	t, _ := f.This().(M)
	return t
}

func (f *CSiteFeature[T, M]) Build(b feature.Buildable) (err error) {
	if err = f.CFeature.Build(b); err != nil {
		return
	}

	f.CSiteIncluding.BuildSiteIncluding(b)

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

func (f *CSiteFeature[T, M]) Startup(ctx *cli.Context) (err error) {
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

	if f.themeName == "" {
		f.theme = f.Site().SiteTheme()
	} else {
		f.theme = f.Enjin.MustGetThemeNamed(f.themeName)
	}
	log.DebugF("%v using site feature theme: %v", f.Tag(), f.theme.Name())
	return
}

func (f *CSiteFeature[T, M]) Shutdown() {
	f.CFeature.Shutdown()
}

func (f *CSiteFeature[T, M]) Site() (s feature.Site) {
	s = f.site
	return
}

func (f *CSiteFeature[T, M]) SetupSiteFeature(s feature.Site) {
	f.site = s
	f.CSiteIncluding.StartupSiteIncluding(f.Enjin)
	return
}

func (f *CSiteFeature[T, M]) RouteSiteFeature(r chi.Router) {
	log.FatalF("%v.RouteSiteFeature method unimplemented", f.Tag())
}

func (f *CSiteFeature[T, M]) SiteFeatureName() (name string) {
	name = f.sitePathName
	return
}

func (f *CSiteFeature[T, M]) SiteFeaturePath() (path string) {
	var sitePath string
	if v := f.site.SitePath(); v != "/" {
		sitePath = v
	}
	path = sitePath + "/" + f.sitePathName
	return
}

func (f *CSiteFeature[T, M]) SiteFeatureMenu() (m menu.Menu) {
	return
}

func (f *CSiteFeature[T, M]) SiteFeatureTheme() (t feature.Theme) {
	t = f.theme
	return
}