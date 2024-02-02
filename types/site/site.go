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
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"

	clPath "github.com/go-corelibs/path"
	"github.com/go-corelibs/x-text/message"
	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/feature/signaling"
	uses_actions "github.com/go-enjin/be/pkg/feature/uses-actions"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/menu"
)

const (
	BaseTag feature.Tag = "site-feature"
)

var (
	_ feature.SiteFeature = (*CSiteFeature[feature.SiteMakeFeature[feature.MakeFeature]])(nil)
)

type CSiteFeature[M interface{}] struct {
	feature.CFeature
	signaling.CSignaling
	uses_actions.CUsesActions

	IncludeSitePathNameFlag bool

	site feature.Site

	featureKey   string
	featureIcon  string
	featureLabel feature.SiteFeatureLabelFn

	themeName string
	theme     feature.Theme
}

func (f *CSiteFeature[M]) SelfFeature() (self feature.SiteFeature) {
	self, _ = f.This().(feature.SiteFeature)
	return
}

func (f *CSiteFeature[M]) Construct(this interface{}) {
	f.CFeature.Construct(this)
	f.CUsesActions.ConstructUsesActions(this)
	return
}

func (f *CSiteFeature[M]) Init(this interface{}) {
	f.CFeature.Init(this)
	f.CFeature.PackageTag = BaseTag
	f.CSignaling.InitSignaling()
	f.IncludeSitePathNameFlag = true
	return
}

func (f *CSiteFeature[M]) SetSiteFeatureKey(kebab string) M {
	f.featureKey = strcase.ToKebab(kebab)
	t, _ := f.This().(M)
	return t
}

func (f *CSiteFeature[M]) SetSiteFeatureIcon(icon string) M {
	f.featureIcon = icon
	t, _ := f.This().(M)
	return t
}

func (f *CSiteFeature[M]) SetSiteFeatureTheme(name string) M {
	f.themeName = name
	t, _ := f.This().(M)
	return t
}

func (f *CSiteFeature[M]) SetSiteFeatureLabel(fn feature.SiteFeatureLabelFn) M {
	f.featureLabel = fn
	t, _ := f.This().(M)
	return t
}

func (f *CSiteFeature[M]) Build(b feature.Buildable) (err error) {
	if err = f.CFeature.Build(b); err != nil {
		return
	}

	if f.IncludeSitePathNameFlag {
		b.AddFlags(&cli.StringFlag{
			Name:     f.KebabTag + "-path-name",
			Usage:    "specify the URL path name for this site feature",
			EnvVars:  b.MakeEnvKeys(f.KebabTag + "-path-name"),
			Category: f.KebabTag,
			Value:    f.SiteFeatureKey(),
		})
	}
	return
}

func (f *CSiteFeature[M]) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}

	pathKey := f.KebabTag + "-path-name"
	if ctx.IsSet(pathKey) {
		if v := ctx.String(pathKey); v != "" {
			if v = forms.StrictSanitize(strings.TrimSpace(v)); v != "" {
				f.featureKey = clPath.TrimSlashes(strcase.ToKebab(v))
			}
		}
	}
	log.InfoF("%v site feature key: %q", f.Tag(), f.SiteFeatureKey())

	return
}

func (f *CSiteFeature[M]) Shutdown() {
	f.CFeature.Shutdown()
}

func (f *CSiteFeature[M]) UserActions() (actions feature.Actions) {
	actions = feature.Actions{
		f.Action("access", "feature"),
	}
	return
}

func (f *CSiteFeature[M]) SetupSiteFeature(s feature.Site) (err error) {
	f.site = s

	if f.themeName == "" {
		f.theme = f.Site().SiteTheme()
	} else {
		f.theme = f.Enjin.MustGetThemeNamed(f.themeName)
	}
	//log.DebugF("%v using site feature theme: %v", f.Tag(), f.theme.Name())
	return
}

func (f *CSiteFeature[M]) Site() (s feature.Site) {
	if f.site == nil {
		panic(fmt.Sprintf("%v.Site() method called before .SetupSiteFeature happens", f.Tag()))
	}
	s = f.site
	return
}

func (f *CSiteFeature[M]) RouteSiteFeature(r chi.Router) {
	return
}

func (f *CSiteFeature[M]) SiteFeatureInfo(r *http.Request) (info *feature.CSiteFeatureInfo) {
	printer := message.GetPrinter(r)
	info = feature.NewSiteFeatureInfo(
		f.Self().Tag().Kebab(),
		f.SelfFeature().SiteFeatureKey(),
		f.SelfFeature().SiteFeatureIcon(),
		f.SelfFeature().SiteFeatureLabel(printer),
	)
	return
}

func (f *CSiteFeature[M]) SiteFeatureKey() (name string) {
	name = f.featureKey
	return
}

func (f *CSiteFeature[M]) SiteFeatureIcon() (icon string) {
	if f.featureIcon != "" {
		icon = f.featureIcon
		return
	}
	icon = "fa-solid fa-question"
	return
}

func (f *CSiteFeature[M]) SiteFeaturePath() (path string) {
	var sitePath string
	if v := f.site.SitePath(); v != "/" {
		sitePath = v
	}
	path = sitePath + "/" + f.featureKey
	return
}

func (f *CSiteFeature[M]) SiteFeatureMenu(r *http.Request) (m menu.Menu) {
	return
}

func (f *CSiteFeature[M]) SiteFeatureTheme() (t feature.Theme) {
	t = f.theme
	return
}

func (f *CSiteFeature[M]) SiteFeatureLabel(printer *message.Printer) (label string) {
	if f.featureLabel != nil {
		label = f.featureLabel(printer)
		return
	}
	label = "Unimplemented"
	return
}

func (f *CSiteFeature[M]) IsBackupProvider() (backup bool) {
	return false
}

func (f *CSiteFeature[M]) SiteSettingsFields(r *http.Request) (fields beContext.Fields) {
	return
}

func (f *CSiteFeature[M]) SiteSettingsPanel(settingsPath string) (serve, handle http.HandlerFunc) {
	return
}
