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

	"github.com/go-chi/chi/v5"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/feature/signaling"
	site_including "github.com/go-enjin/be/pkg/feature/site-including"
	uses_actions "github.com/go-enjin/be/pkg/feature/uses-actions"
	uses_kvc "github.com/go-enjin/be/pkg/feature/uses-kvc"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/log"
	clPath "github.com/go-corelibs/path"
	"github.com/go-enjin/be/pkg/userbase"
)

var (
	DefaultSitePath = "/"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "site-enjin"

type Feature interface {
	feature.Feature
	signaling.Signaling
	feature.Site
	feature.ApplyMiddleware
	feature.UserActionsProvider
	feature.FinalizeServeRequestFeature
}

type MakeFeature interface {
	uses_kvc.MakeFeature[MakeFeature]

	SetSitePath(path string) MakeFeature
	SetSiteTheme(name string) MakeFeature
	SetSiteUsers(tag feature.Tag) MakeFeature
	SetSiteAuth(tag feature.Tag) MakeFeature

	IncludeSiteFeatures(features ...feature.Feature) MakeFeature
	IncludingSiteFeatures(tags ...feature.Tag) MakeFeature

	UseSiteRootFeature(srf feature.Feature) MakeFeature
	UsingSiteRootFeature(tag feature.Tag) MakeFeature

	Make() Feature
}

type CFeature struct {
	feature.CFeature
	signaling.CSignaling
	uses_actions.CUsesActions
	uses_kvc.CUsesKVC[MakeFeature]

	themeName string
	theme     feature.Theme

	sitePath string

	siteUsersProviderTag feature.Tag
	siteUsersProvider    feature.SiteUsersProvider

	authFeatureTag feature.Tag
	authFeature    feature.SiteAuthFeature
	siteFeatures   *site_including.CSiteIncluding[feature.SiteFeature, MakeFeature]

	siteRootFeatureTag feature.Tag
	siteRootFeature    feature.SiteRootFeature

	userNoticeLocker  feature.SyncLocker
	userNoticeBucket  feature.KeyValueStore
	userContextLocker feature.SyncLocker
	userContextBucket feature.KeyValueStore
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
	f.CUsesActions.ConstructUsesActions(f)
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.CUsesKVC.InitUsesKVC(this)
	f.CSignaling.InitSignaling()
	f.siteFeatures = site_including.New[feature.SiteFeature, MakeFeature](this)
	f.sitePath = DefaultSitePath
	return
}

func (f *CFeature) SetSitePath(path string) MakeFeature {
	f.sitePath = "/" + clPath.TrimSlashes(path)
	return f
}

func (f *CFeature) SetSiteTheme(name string) MakeFeature {
	f.themeName = name
	return f
}

func (f *CFeature) SetSiteUsers(tag feature.Tag) MakeFeature {
	f.siteUsersProviderTag = tag
	return f
}

func (f *CFeature) SetSiteAuth(tag feature.Tag) MakeFeature {
	f.authFeatureTag = tag
	return f
}

func (f *CFeature) IncludeSiteFeatures(features ...feature.Feature) MakeFeature {
	f.siteFeatures.Include(features...)
	return f
}

func (f *CFeature) IncludingSiteFeatures(tags ...feature.Tag) MakeFeature {
	f.siteFeatures.Including(tags...)
	return f
}

func (f *CFeature) UseSiteRootFeature(feat feature.Feature) MakeFeature {
	if srf, ok := feat.This().(feature.SiteRootFeature); ok {
		f.siteRootFeature = srf
	} else {
		log.FatalDF(1, "%q is not a feature.SiteRootFeature", feat.Tag().String())
	}
	return f
}

func (f *CFeature) UsingSiteRootFeature(tag feature.Tag) MakeFeature {
	f.siteRootFeatureTag = tag
	return f
}

func (f *CFeature) Make() (feat Feature) {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	if err = f.CFeature.Build(b); err != nil {
		return
	} else if err = f.CUsesKVC.BuildUsesKVC(); err != nil {
		return
	}
	f.siteFeatures.BuildSiteIncluding(b)
	if f.siteRootFeature != nil {
		b.AddFeature(f.siteRootFeature)
	}
	category := f.FeatureTag.Kebab()
	b.AddFlags(&cli.StringFlag{
		Name:     category + "-path",
		Usage:    "specify the top-level site path",
		EnvVars:  b.MakeEnvKeys(category + "-path"),
		Category: category,
		Value:    f.sitePath,
	})
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	} else if err = f.CUsesKVC.StartupUsesKVC(f.Enjin.Features()); err != nil {
		return
	}
	f.siteFeatures.StartupSiteIncluding(f.Enjin)

	f.userNoticeBucket = f.KVC().MustBucket("user-notices")
	noticeLockerBucket := f.KVC().MustBucket("user-notice-locker")
	f.userNoticeLocker = f.Enjin.NewSyncLocker(f.Tag(), "user-notice-locker", noticeLockerBucket)

	f.userContextBucket = f.KVC().MustBucket("user-context")
	contextLockerBucket := f.KVC().MustBucket("user-context-locker")
	f.userContextLocker = f.Enjin.NewSyncLocker(f.Tag(), "user-context-locker", contextLockerBucket)

	if f.siteRootFeature == nil {
		if !f.siteRootFeatureTag.IsNil() {
			if srf, ok := f.siteFeatures.Features.Get(f.siteRootFeatureTag).(feature.SiteRootFeature); ok {
				f.siteRootFeature = srf
			} else {
				err = fmt.Errorf("site root feature not found within this site: %v", f.siteRootFeatureTag)
				return
			}
		}
	}

	if f.siteRootFeature != nil {
		if f.sitePath == "/" {
			err = fmt.Errorf("domain root sites does not support having an internal site root feature, use %v.SetSitePath with something other than \"/\"", f.Tag())
			return
		}
	}

	if f.authFeatureTag.IsNil() {
		for _, sf := range f.siteFeatures.Features {
			if saf, ok := sf.This().(feature.SiteAuthFeature); ok {
				f.authFeature = saf
				break
			}
		}
	} else if saf, ok := f.siteFeatures.Features.Get(f.authFeatureTag).(feature.SiteAuthFeature); ok {
		f.authFeature = saf
	} else {
		err = fmt.Errorf("auth feature not found: %q", f.authFeatureTag)
		return
	}

	if f.authFeature == nil {

		log.DebugF("feature.SiteAuthFeature not found, site users are unrestricted")

	} else if tag := f.siteUsersProviderTag; !tag.IsNil() {

		enjinFeatures := f.Enjin.Features().List()
		if f.siteUsersProvider, err = feature.GetTyped[feature.SiteUsersProvider](tag, enjinFeatures); err != nil {
			return
		}

		if f.siteUsersProvider == nil {
			err = fmt.Errorf("%q SiteAuthFeature needs an enjin SiteUsersProvider feature", f.authFeature.Tag())
			return
		}
	}

	prefix := f.FeatureTag.Kebab()
	pathKey := prefix + "-path"
	if ctx.IsSet(pathKey) {
		if v := ctx.String(pathKey); v != "" {
			if v = forms.StrictSanitize(v); v != "" {
				if v = clPath.TrimSlashes(v); v != "" {
					f.sitePath = "/" + v
				}
			}
		}
	}
	log.InfoF("%v site path: %v", f.Tag(), f.sitePath)

	if f.themeName == "" {
		f.theme = f.Enjin.MustGetTheme()
	} else {
		f.theme = f.Enjin.MustGetThemeNamed(f.themeName)
	}
	log.DebugF("using site theme: %v", f.theme.Name())

	return
}

func (f *CFeature) PostStartup(ctx *cli.Context) (err error) {
	if f.siteRootFeature != nil {
		if err = f.siteRootFeature.SetupSiteFeature(f); err != nil {
			return
		}
	}
	for _, ef := range f.siteFeatures.Features {
		if err = ef.SetupSiteFeature(f); err != nil {
			return
		}
	}
	return
}

func (f *CFeature) Shutdown() {
	f.CFeature.Shutdown()
}

func (f *CFeature) UserActions() (list feature.Actions) {
	list = feature.Actions{
		f.Action("access", "site"),
	}
	return
}

func (f *CFeature) Use(s feature.System) feature.MiddlewareFn {
	if f.sitePath == "/" {
		s.Router().Use(f.requireUserMiddleware())
		s.Router().Use(f.homePathMiddleware)
		s.Router().Use(f.authEnjinProviderMiddleware)
	}
	if s.MustGetTheme().Name() != f.theme.Name() {
		return f.theme.Middleware
	}
	return nil
}

func (f *CFeature) Apply(s feature.System) (err error) {

	route := func(r chi.Router) {

		if f.sitePath != "/" {
			r.Use(f.requireUserMiddleware())
			r.Use(f.homePathMiddleware)
			r.Use(f.authEnjinProviderMiddleware)

			if f.siteRootFeature != nil {
				r.Handle("/", f.siteRootFeature.SiteRootHandler())
			} else if len(f.siteFeatures.Features) > 0 {
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
					if userbase.CurrentUserCan(r, f.siteFeatures.Features[0].Action("access", "feature")) {
						f.Enjin.ServeRedirect(f.siteFeatures.Features[0].SiteFeaturePath(), w, r)
					} else {
						f.Enjin.ServeNotFound(w, r)
					}
				})
			}
		}

		for _, ef := range f.siteFeatures.Features {
			if usrf, ok := ef.This().(feature.UpdateSiteRoutesFeature); ok {
				usrf.UpdateSiteRoutes(r)
				continue
			}
			r.Route("/"+ef.SiteFeatureKey(), func(r chi.Router) {
				r.Use(userbase.RequireUserCan(f.Enjin, ef.Action("access", "feature")))
				r.Use(f.homePathMiddleware)
				ef.RouteSiteFeature(r)
			})
		}

	}

	if f.sitePath == "/" {
		route(s.Router())
		return
	}

	s.Router().Route(f.sitePath, route)
	return
}
