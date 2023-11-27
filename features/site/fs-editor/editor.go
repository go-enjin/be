//go:build editor || all

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

package editor

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/golang-org-x-text/message"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	site_including "github.com/go-enjin/be/pkg/feature/site-including"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/menu"
	"github.com/go-enjin/be/pkg/userbase"
	"github.com/go-enjin/be/types/site"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "fs-editor"

type Feature interface {
	feature.SiteFeature
	feature.UseMiddleware
	feature.UserActionsProvider
	feature.PageContextFieldsProvider
}

type MakeFeature interface {
	feature.SiteMakeFeature[MakeFeature]
	site_including.MakeFeature[MakeFeature]

	SetWithoutSubMenus(without bool) MakeFeature

	Make() Feature
}

type CFeature struct {
	site.CSiteFeature[MakeFeature]
	site_including.CSiteIncluding[feature.EditorFeature, MakeFeature]

	withoutSubMenus bool
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.SetSiteFeatureKey("fs-editor")
	f.SetSiteFeatureIcon("fa-solid fa-pen-ruler")
	f.SetSiteFeatureLabel(func(printer *message.Printer) (label string) {
		label = printer.Sprintf("FS Editor")
		return
	})
	f.CUsesActions.ConstructUsesActions(f)
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CSiteFeature.Init(this)
	f.CSiteIncluding.InitSiteIncluding(this)
	return
}

func (f *CFeature) SetWithoutSubMenus(without bool) MakeFeature {
	f.withoutSubMenus = without
	return f
}

func (f *CFeature) Make() (feat Feature) {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	if err = f.CSiteFeature.Build(b); err != nil {
		return
	}
	f.CSiteIncluding.BuildSiteIncluding(b)
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CSiteFeature.Startup(ctx); err != nil {
		return
	}
	f.CSiteIncluding.StartupSiteIncluding(f.Enjin)

	if handler := f.Enjin.GetServePagesHandler(); handler == nil {
		err = fmt.Errorf("enjin serve-pages handler not found")
		return
	}

	return
}

func (f *CFeature) UserActions() (list feature.Actions) {
	list = append(f.CSiteFeature.UserActions(),
		f.Action("edit", "file-editor"),
		f.Action("view", "file-editor"),
		f.Action("view", "file-browser"),
		f.Action("create", "file-browser"),
		f.Action("delete", "file-browser"),
	)
	return
}

func (f *CFeature) Use(s feature.System) feature.MiddlewareFn {
	t := f.SiteFeatureTheme()
	if s.MustGetTheme().Name() != t.Name() {
		return t.Middleware
	}
	return nil
}

func (f *CFeature) SetupSiteFeature(s feature.Site) (err error) {
	if err = f.CSiteFeature.SetupSiteFeature(s); err != nil {
		return
	}
	// setup site feature first
	for _, ef := range f.Features {
		if err = ef.SetupSiteFeature(s); err != nil {
			return
		}
	}
	// then setup editors
	for _, ef := range f.Features {
		ef.SetupEditor(f)
	}
	return
}

func (f *CFeature) RouteSiteFeature(r chi.Router) {

	r.Use(userbase.RequireUserCan(f.Enjin, f.Action("access", "feature")))

	if len(f.Features) > 0 {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			f.Enjin.ServeRedirect(f.Features[0].GetEditorPath(), w, r)
		})
	}
	for _, ef := range f.Features {
		r.Route("/"+ef.GetEditorKey(), func(r chi.Router) {
			ef.SetupEditorRoute(r)
		})
	}

	return
}

func (f *CFeature) SiteFeatureMenu(r *http.Request) (m menu.Menu) {
	info := f.SiteFeatureInfo(r)
	item := &menu.Item{
		Text: info.Label,
		Href: f.SiteFeaturePath(),
		Icon: info.Icon,
	}
	for _, ef := range f.Features {
		if userbase.CurrentUserCan(r, ef.Action("access", "feature")) {
			item.SubMenu = append(item.SubMenu, ef.EditorMenu(r)...)
		}
	}
	if f.withoutSubMenus && len(item.SubMenu) > 0 {
		m = item.SubMenu
		return
	}
	m = menu.Menu{item}
	return
}

func (f *CFeature) MakePageContextFields(r *http.Request) (fields beContext.Fields) {
	printer := lang.GetPrinterFromRequest(r)
	fields = beContext.Fields{
		"title": {
			Key:      "title",
			Tab:      "page",
			Label:    printer.Sprintf("The page's title, used in the browser tab name and other places"),
			Category: "important",
			Weight:   100,
			Input:    "text",
			Format:   "string",
			Required: true,
		},
		"description": {
			Key:      "description",
			Tab:      "page",
			Label:    printer.Sprintf("A brief description of the page, used for SEO headers and page excerpts"),
			Category: "important",
			Weight:   99,
			Input:    "text",
			Format:   "string",
			Required: true,
		},
		"type": {
			Key:          "type",
			Tab:          "page",
			Label:        printer.Sprintf("The type of page this is"),
			Category:     "file",
			Weight:       90,
			Input:        "select",
			Format:       "kebab-option",
			DefaultValue: "page",
			ValueOptions: f.ListPageTypes(),
		},
		"created": {
			Key:      "created",
			Tab:      "page",
			Label:    printer.Sprintf("The date the page file was created"),
			Category: "file",
			Weight:   77,
			Input:    "datetime-local",
			Format:   "time-struct",
		},
		"updated": {
			Key:      "updated",
			Tab:      "page",
			Label:    printer.Sprintf("The date the page file was last updated"),
			Category: "file",
			Weight:   77,
			Input:    "datetime-local",
			Format:   "time-struct",
		},
		"layout": {
			Key:          "layout",
			Tab:          "page",
			Label:        printer.Sprintf("Specify the theme layout to use"),
			Category:     "theme",
			Weight:       100,
			Input:        "select",
			Format:       "kebab-option",
			DefaultValue: "defaults",
			ValueOptions: f.ListPageLayouts(),
		},
		"thumbnail-url": {
			Key:      "thumbnail-url",
			Tab:      "page",
			Label:    printer.Sprintf("Specify an image URL to use for this page's thumbnail"),
			Category: "theme",
			Weight:   77,
			Input:    "text",
			Format:   "url",
		},
		"thumbnail-alt": {
			Key:      "thumbnail-alt",
			Tab:      "page",
			Label:    printer.Sprintf("Specify the alt attribute text to use for this page's thumbnail"),
			Category: "theme",
			Weight:   77,
			Input:    "text",
			Format:   "string",
		},
		"no-page-indexing": {
			Key:      "no-page-indexing",
			Tab:      "page",
			Label:    printer.Sprintf("Omit this page from any indexing"),
			Category: "indexing",
			Weight:   100,
			Input:    "checkbox",
			Format:   "bool",
		},
		"no-search-indexing": {
			Key:      "no-search-indexing",
			Tab:      "page",
			Label:    printer.Sprintf("Omit this page from search indexing"),
			Category: "indexing",
			Weight:   100,
			Input:    "checkbox",
			Format:   "bool",
		},
	}

	tag := lang.GetTag(r)
	if f.Enjin.SiteDefaultLanguage().String() != tag.String() {
		fields["translates"] = &beContext.Field{
			Key:      "translates",
			Tab:      "page",
			Label:    printer.Sprintf("Specify the original page on this site that this page translates"),
			Category: "file",
			Weight:   30,
			Input:    "text",
			Format:   "relative-url",
		}
	}
	return
}

func (f *CFeature) ListPageTypes() (values []string) {
	unique := map[string]struct{}{}
	for _, ef := range f.Enjin.GetPageTypeProcessors() {
		for _, name := range ef.PageTypeNames() {
			unique[name] = struct{}{}
		}
	}
	values = append([]string{"page"}, maps.SortedKeys(unique)...)
	return
}

func (f *CFeature) ListPageLayouts() (names []string) {
	names = f.Enjin.MustGetTheme().Layouts().ListLayouts()
	return
}
