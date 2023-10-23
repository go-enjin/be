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

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/menu"
	"github.com/go-enjin/be/pkg/userbase"
	"github.com/go-enjin/be/types/site"
	"github.com/go-enjin/golang-org-x-text/language"
)

var (
	DefaultEditorPath = "/fs-editor"
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
	Make() Feature
	feature.SiteIncludingMakeFeature[MakeFeature]

	SetEditorTheme(name string) MakeFeature
}

type CFeature struct {
	site.CSiteFeature[feature.EditorFeature, MakeFeature]

	themeName string
	theme     feature.Theme

	editorPath string
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.SetSiteFeaturePathName(tag.String())
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CSiteFeature.Init(this)
	f.editorPath = DefaultEditorPath
	return
}

func (f *CFeature) SetEditorTheme(name string) MakeFeature {
	f.themeName = name
	return f
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

	if handler := f.Enjin.GetServePagesHandler(); handler == nil {
		err = fmt.Errorf("enjin serve-pages handler not found")
		return
	}

	if f.themeName != "" {
		f.theme = f.Enjin.MustGetThemeNamed(f.themeName)
	}
	if f.theme == nil {
		f.theme = f.Enjin.MustGetTheme()
	}
	log.DebugF("%v editor theme: %v", f.Tag(), f.theme.Name())
	return
}

func (f *CFeature) UserActions() (list feature.Actions) {
	list = feature.Actions{
		feature.NewAction(f.Tag().String(), "access", "editor"),
		feature.NewAction(f.Tag().String(), "edit", "file-editor"),
		feature.NewAction(f.Tag().String(), "view", "file-editor"),
		feature.NewAction(f.Tag().String(), "view", "file-browser"),
		feature.NewAction(f.Tag().String(), "create", "file-browser"),
		feature.NewAction(f.Tag().String(), "delete", "file-browser"),
	}
	return
}

func (f *CFeature) Use(s feature.System) feature.MiddlewareFn {
	if s.MustGetTheme().Name() != f.theme.Name() {
		return f.theme.Middleware
	}
	return nil
}

func (f *CFeature) SetupSiteFeature(s feature.Site) {
	f.CSiteFeature.SetupSiteFeature(s)
	for _, ef := range f.Features {
		ef.SetupEditor(f)
	}
}

func (f *CFeature) RouteSiteFeature(r chi.Router) {

	r.Use(userbase.RequireUserCan(f.Enjin, feature.NewAction(f.Tag().String(), "access", "editor")))

	if len(f.Features) > 0 {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			f.Enjin.ServeRedirect(f.Features[0].GetEditorPath(), w, r)
		})
	}
	for _, ef := range f.Features {
		r.Route("/"+ef.GetEditorName(), func(r chi.Router) {
			ef.SetupEditorRoute(r)
		})
	}

	return
}

func (f *CFeature) SiteFeatureMenu() (m menu.Menu) {
	for _, ef := range f.Features {
		m = append(m, &menu.Item{
			Text:    ef.GetEditorName(),
			Href:    ef.GetEditorPath(),
			Lang:    language.English.String(),
			SubMenu: ef.SelfEditor().GetEditorMenu(),
		})
	}
	return
}

func (f *CFeature) EditorPath() (path string) {
	path = f.SiteFeaturePath()
	return
}

func (f *CFeature) EditorTheme() (t feature.Theme) {
	return f.theme
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