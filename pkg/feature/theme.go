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

package feature

import (
	htmlTemplate "html/template"
	"net/http"
	textTemplate "text/template"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/net/headers/policy/csp"
	"github.com/go-enjin/be/pkg/net/headers/policy/permissions"
)

type Author struct {
	Name     string
	Homepage string
}

type ThemeConfig struct {
	Name        string
	Parent      string
	License     string
	LicenseLink string
	Description string
	Homepage    string
	Authors     []Author
	Extends     string

	RootStyles  []htmlTemplate.CSS
	BlockStyles map[string][]htmlTemplate.CSS
	BlockThemes map[string]map[string]interface{}

	FontawesomeLinks   map[string]string
	FontawesomeClasses []string

	CacheControl string

	PermissionsPolicy     []permissions.Directive
	ContentSecurityPolicy csp.ContentSecurityPolicyConfig

	Context context.Context
}

type Theme interface {
	Name() (name string)

	GetParent() (parent Theme)
	GetBlockThemeNames() (names []string)

	ThemeFS() fs.FileSystem
	StaticFS() fs.FileSystem
	Locales() (locales fs.FileSystem, ok bool)

	GetConfig() (config ThemeConfig)

	Layouts() ThemeLayouts
	FindLayout(named string) (layout ThemeLayout, name string, ok bool)

	AddFormatProvider(providers ...PageFormatProvider)
	PageFormatProvider

	Middleware(next http.Handler) http.Handler

	NewHtmlTemplate(enjin Internals, name string, ctx context.Context) (tmpl *htmlTemplate.Template, err error)
	NewTextTemplate(enjin Internals, name string, ctx context.Context) (tmpl *textTemplate.Template, err error)
}

type ThemeLayout interface {
	Name() (name string)

	NewHtmlTemplate(enjin Internals, ctx context.Context) (tmpl *htmlTemplate.Template, err error)
	NewTextTemplate(enjin Internals, ctx context.Context) (tmpl *textTemplate.Template, err error)

	NewHtmlTemplateFrom(enjin Internals, parent ThemeLayout, ctx context.Context) (tmpl *htmlTemplate.Template, err error)
	NewTextTemplateFrom(enjin Internals, parent ThemeLayout, ctx context.Context) (tmpl *textTemplate.Template, err error)

	ApplyHtmlTemplate(enjin Internals, tt *htmlTemplate.Template, ctx context.Context) (err error)
	ApplyTextTemplate(enjin Internals, tt *textTemplate.Template, ctx context.Context) (err error)
}

type ThemeLayouts interface {
	GetLayout(name string) (layout ThemeLayout)
	SetLayout(name string, layout ThemeLayout)

	NewHtmlTemplate(enjin Internals, name string, ctx context.Context) (tmpl *htmlTemplate.Template, err error)
	NewTextTemplate(enjin Internals, name string, ctx context.Context) (tmpl *textTemplate.Template, err error)

	ApplyHtmlTemplates(enjin Internals, tmpl *htmlTemplate.Template, ctx context.Context) (err error)
	ApplyTextTemplates(enjin Internals, tmpl *textTemplate.Template, ctx context.Context) (err error)
}

type ThemeRenderer interface {
	Feature
	Render(t Theme, view string, ctx context.Context) (data []byte, err error)
	RenderPage(t Theme, ctx context.Context, p Page) (data []byte, redirect string, err error)

	NewHtmlTemplateWith(t Theme, name string, ctx context.Context) (tt *htmlTemplate.Template, err error)
	NewTextTemplateWith(t Theme, name string, ctx context.Context) (tt *textTemplate.Template, err error)

	RenderHtmlTemplateContent(t Theme, ctx context.Context, tmplContent string) (rendered string, err error)
	RenderTextTemplateContent(t Theme, ctx context.Context, tmplContent string) (rendered string, err error)

	NewHtmlTemplateFromContext(t Theme, view string, ctx context.Context) (tt *htmlTemplate.Template, err error)
	NewTextTemplateFromContext(t Theme, view string, ctx context.Context) (tt *textTemplate.Template, err error)
}

type RoutePagesHandler interface {
	Feature
	RoutePage(w http.ResponseWriter, r *http.Request)
}