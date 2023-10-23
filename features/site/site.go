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
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/feature/signaling"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/menu"
	bePath "github.com/go-enjin/be/pkg/path"
	"github.com/go-enjin/be/pkg/userbase"
	"github.com/go-enjin/be/types/page"
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
	feature.HotReloadableFeature
}

type MakeFeature interface {
	feature.SiteIncludingMakeFeature[MakeFeature]

	SetSitePath(path string) MakeFeature
	SetSiteTheme(name string) MakeFeature

	Make() Feature
}

type CFeature struct {
	feature.CFeature
	signaling.CSignaling
	feature.CSiteIncluding[feature.SiteFeature, MakeFeature]

	themeName string
	theme     feature.Theme

	sitePath string

	// TODO: make these kvs based
	userMutex   *sync.RWMutex
	userCache   map[string]beContext.Context
	userNotices map[string]feature.UserNotices
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
	f.CSignaling.InitSignaling()
	f.CSiteIncluding.InitSiteIncluding(this)
	f.sitePath = DefaultSitePath
	f.userMutex = &sync.RWMutex{}
	f.userCache = make(map[string]beContext.Context)
	f.userNotices = make(map[string]feature.UserNotices)
	return
}

func (f *CFeature) SetSitePath(path string) MakeFeature {
	f.sitePath = "/" + bePath.TrimSlashes(path)
	return f
}

func (f *CFeature) SetSiteTheme(name string) MakeFeature {
	f.themeName = name
	return f
}

func (f *CFeature) Make() (feat Feature) {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	if err = f.CFeature.Build(b); err != nil {
		return
	}
	f.BuildSiteIncluding(b)
	category := f.FeatureTag.String()
	prefix := f.FeatureTag.Kebab()
	b.AddFlags(&cli.StringFlag{
		Name:     prefix + "-path",
		Usage:    "specify the top-level site URL path",
		Category: category,
		Value:    f.sitePath,
	})
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}

	prefix := f.FeatureTag.Kebab()
	pathKey := prefix + "-path"
	if ctx.IsSet(pathKey) {
		if v := ctx.String(pathKey); v != "" {
			if v = forms.StrictSanitize(v); v != "" {
				if v = bePath.TrimSlashes(v); v != "" {
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

	f.CSiteIncluding.StartupSiteIncluding(f.Enjin)

	for _, ef := range f.Features {
		ef.SetupSiteFeature(f)
	}
	return
}

func (f *CFeature) Shutdown() {
	f.CFeature.Shutdown()
}

func (f *CFeature) HotReload() (err error) {
	return
}

func (f *CFeature) UserActions() (list feature.Actions) {
	list = feature.Actions{
		feature.NewAction(f.Tag().String(), "access", "site"),
	}
	return
}

func (f *CFeature) Use(s feature.System) feature.MiddlewareFn {
	if f.sitePath == "/" {
		s.Router().Use(userbase.RequireUserCan(f.Enjin, feature.NewAction(f.Tag().String(), "access", "site")))
	}
	if s.MustGetTheme().Name() != f.theme.Name() {
		return f.theme.Middleware
	}
	return nil
}

func (f *CFeature) Apply(s feature.System) (err error) {

	route := func(r chi.Router) {
		if f.sitePath != "/" {
			r.Use(userbase.RequireUserCan(f.Enjin, feature.NewAction(f.Tag().String(), "access", "site")))
			if len(f.Features) > 0 {
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
					f.Enjin.ServeRedirect(f.Features[0].SiteFeaturePath(), w, r)
				})
			}
		}
		for _, ef := range f.Features {
			r.Route("/"+ef.SiteFeatureName(), func(r chi.Router) {
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

func (f *CFeature) SitePath() (path string) {
	path = f.sitePath
	return
}

func (f *CFeature) SiteTheme() (t feature.Theme) {
	return f.theme
}

func (f *CFeature) SiteMenu() (siteMenu beContext.Context) {
	mainMenu := menu.Menu{}
	for _, ef := range f.Features {
		mainMenu = append(mainMenu, &menu.Item{
			Text:    ef.SiteFeatureName(),
			Href:    ef.SiteFeaturePath(),
			SubMenu: ef.SiteFeatureMenu(),
		})
	}
	return beContext.Context{
		"MainMenu": mainMenu,
	}
}

func (f *CFeature) PushInfoNotice(eid, message string, dismiss bool, actions ...feature.UserNoticeLink) {
	f.PushNotices(eid, feature.MakeInfoNotice(message, dismiss, actions...))
}

func (f *CFeature) PushWarnNotice(eid, message string, dismiss bool, actions ...feature.UserNoticeLink) {
	f.PushNotices(eid, feature.MakeWarnNotice(message, dismiss, actions...))
}

func (f *CFeature) PushErrorNotice(eid, message string, dismiss bool, actions ...feature.UserNoticeLink) {
	f.PushNotices(eid, feature.MakeErrorNotice(message, dismiss, actions...))
}

func (f *CFeature) PushNotices(eid string, notices ...*feature.UserNotice) {
	f.userMutex.Lock()
	defer f.userMutex.Unlock()
	f.userNotices[eid] = append(f.userNotices[eid], notices...)
	return
}

func (f *CFeature) PullNotices(eid string) (notices feature.UserNotices) {
	f.userMutex.Lock()
	defer f.userMutex.Unlock()
	notices = append(notices, f.userNotices[eid]...)
	delete(f.userNotices, eid)
	return
}

func (f *CFeature) GetContext(eid string) (ctx beContext.Context) {
	f.userMutex.Lock()
	defer f.userMutex.Unlock()
	var ok bool
	if ctx, ok = f.userCache[eid]; !ok {
		ctx = beContext.New()
		f.userCache[eid] = ctx
	}
	return
}

func (f *CFeature) SetContext(eid string, ctx beContext.Context) {
	f.userMutex.Lock()
	defer f.userMutex.Unlock()
	if v, ok := f.userCache[eid]; ok && v != nil {
		f.userCache[eid].ApplySpecific(ctx)
	} else {
		f.userCache[eid] = ctx
	}
	return
}

func (f *CFeature) PreparePage(layout, pageType, pagePath string, t feature.Theme) (pg feature.Page, ctx beContext.Context, err error) {
	content := feature.MakeRawPage(beContext.Context{
		"type":   pageType,
		"layout": layout,
	}, "")

	ctx = f.Enjin.Context()
	now := time.Now().Unix()

	if pg, err = page.New(f.Tag().String(), pagePath, content, now, now, t, ctx); err != nil {
		err = errors.Wrap(err, "error making new page instance")
		return
	}

	m := menu.Menu{}
	for _, sf := range f.Features {
		m = append(m, &menu.Item{
			Text:    sf.SiteFeatureName(),
			Href:    sf.SiteFeaturePath(),
			SubMenu: sf.SiteFeatureMenu(),
		})
	}

	ctx.SetSpecific("SiteMenu", beContext.Context{"MainMenu": m})

	return
}

func (f *CFeature) ServePreparedPage(pg feature.Page, ctx beContext.Context, t feature.Theme, w http.ResponseWriter, r *http.Request) {
	handler := f.Enjin.GetServePagesHandler()
	if err := handler.ServePage(pg, t, ctx, w, r); err != nil {
		log.ErrorRF(r, "error serving %v prepared page: %v", f.Tag(), err)
		f.Enjin.ServeInternalServerError(w, r)
	}
}