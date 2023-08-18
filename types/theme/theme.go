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

package theme

import (
	"html/template"
	"net/http"
	"sort"

	"github.com/fvbommel/sortorder"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/log"
	bePath "github.com/go-enjin/be/pkg/path"
	"github.com/go-enjin/be/types/theme/layouts"
)

var (
	DefaultCacheControl = "public, max-age=604800, no-transform, immutable"
)

var (
	_ feature.Theme = (*CTheme)(nil)
)

type CTheme struct {
	origin string
	name   string
	path   string
	config feature.ThemeConfig
	minify bool

	layouts *layouts.Layouts

	fs       fs.FileSystem
	staticFS fs.FileSystem

	formatProviders []feature.PageFormatProvider
}

func New(origin, path string, themeFs, staticFs fs.FileSystem) (ft feature.Theme, err error) {
	ft, err = newTheme(origin, path, themeFs, staticFs)
	return
}

func newTheme(origin, path string, themeFs, staticFs fs.FileSystem) (t *CTheme, err error) {
	t = new(CTheme)
	t.origin = origin
	t.name = bePath.Base(path)
	t.path = path
	t.fs = themeFs
	t.staticFS = staticFs
	if err = t.init(); err != nil {
		return
	}
	t.registerTheme()
	return
}

func (t *CTheme) Reload() (err error) {
	var nt *CTheme
	if nt, err = newTheme(t.origin, t.path, t.fs, t.staticFS); err == nil {
		nt.formatProviders = t.formatProviders
		*t = *nt
		if parent := t.GetParent(); parent != nil {
			err = parent.Reload()
		}
	}
	return
}

func (t *CTheme) Name() string {
	return t.name
}

func (t *CTheme) ThemeFS() fs.FileSystem {
	return t.fs
}

func (t *CTheme) StaticFS() fs.FileSystem {
	return t.staticFS
}

func (t *CTheme) Layouts() feature.ThemeLayouts {
	return t.layouts
}

func (t *CTheme) GetConfig() (config feature.ThemeConfig) {
	config = feature.ThemeConfig{
		Name:                  t.config.Name,
		Parent:                t.config.Parent,
		License:               t.config.License,
		LicenseLink:           t.config.LicenseLink,
		Description:           t.config.Description,
		Homepage:              t.config.Homepage,
		Authors:               t.config.Authors,
		Extends:               t.config.Extends,
		FontawesomeLinks:      make(map[string]string),
		PermissionsPolicy:     t.config.PermissionsPolicy,
		ContentSecurityPolicy: t.config.ContentSecurityPolicy,
	}

	config.BlockStyles = make(map[string][]template.CSS)
	config.BlockThemes = make(map[string]map[string]interface{})

	if parent := t.GetParent(); parent != nil {

		parentConfig := parent.GetConfig()

		config.RootStyles = append(
			parentConfig.RootStyles,
			t.config.RootStyles...,
		)

		for k, v := range parentConfig.BlockStyles {
			config.BlockStyles[k] = append([]template.CSS{}, v...)
		}

		for k, v := range parentConfig.BlockThemes {
			config.BlockThemes[k] = make(map[string]interface{})
			for j, vv := range v {
				config.BlockThemes[k][j] = vv
			}
		}

		for _, v := range parentConfig.FontawesomeClasses {
			config.FontawesomeClasses = append(config.FontawesomeClasses, v)
		}
		for k, v := range parentConfig.FontawesomeLinks {
			config.FontawesomeLinks[k] = v
		}

		config.Context = parentConfig.Context.Copy()
	} else {
		config.RootStyles = t.config.RootStyles
		config.Context = context.New()
	}

	for k, v := range t.config.BlockStyles {
		config.BlockStyles[k] = append([]template.CSS{}, v...)
	}
	for k, v := range t.config.BlockThemes {
		config.BlockThemes[k] = make(map[string]interface{})
		for j, vv := range v {
			config.BlockThemes[k][j] = vv
		}
	}

	for _, v := range t.config.FontawesomeClasses {
		config.FontawesomeClasses = append(config.FontawesomeClasses, v)
	}
	for k, v := range t.config.FontawesomeLinks {
		config.FontawesomeLinks[k] = v
	}

	config.Context.Apply(t.config.Context)
	return
}

func (t *CTheme) GetBlockThemeNames() (names []string) {
	names = append(names, "primary", "secondary")
	for k, _ := range t.GetConfig().BlockThemes {
		names = append(names, k)
	}
	return
}

func (t *CTheme) AddFormatProvider(providers ...feature.PageFormatProvider) {
	t.formatProviders = append(t.formatProviders, providers...)
}

func (t *CTheme) ListFormats() (names []string) {
	for _, p := range t.formatProviders {
		names = append(names, p.ListFormats()...)
	}
	sort.Sort(sortorder.Natural(names))
	return
}

func (t *CTheme) GetFormat(name string) (format feature.PageFormat) {
	for _, provider := range t.formatProviders {
		if format = provider.GetFormat(name); format != nil {
			return
		}
	}
	return
}

func (t *CTheme) MatchFormat(filename string) (format feature.PageFormat, match string) {
	for _, provider := range t.formatProviders {
		if format, match = provider.MatchFormat(filename); format != nil {
			return
		}
	}
	return
}

func (t *CTheme) Locales() (locales fs.FileSystem, ok bool) {
	if _, err := t.fs.ReadDir("locales"); err == nil {
		log.DebugF("found %v theme locales", t.name)
		if locales, err = fs.Wrap("locales", t.fs); err == nil {
			ok = true
		} else {
			log.ErrorF("error wrapping %v theme locales: %v", t.name, err)
			locales = nil
		}
	}
	return
}

func (t *CTheme) FindLayout(named string) (layout feature.ThemeLayout, name string, ok bool) {
	if named == "" {
		named = "_default"
	}
	name = named

	layout = t.layouts.GetLayout(name)
	if ok = layout != nil; ok {
		if t.config.Parent == "" {
			log.TraceF("found layout in %v (theme) context: %v", t.name, name)
		} else {
			log.TraceF("found layout in %v (%v) context: %v", t.name, t.config.Parent, name)
		}
	}
	return
}

func (t *CTheme) Middleware(next http.Handler) http.Handler {
	log.DebugF("including %v theme static middleware", t.name)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := bePath.TrimSlashes(r.URL.Path)
		var err error
		var data []byte
		var mime string
		if data, err = t.staticFS.ReadFile(path); err != nil {
			next.ServeHTTP(w, r)
			return
		}
		mime, _ = t.staticFS.MimeType(path)
		w.Header().Set("Content-Type", mime)
		if t.config.CacheControl == "" {
			w.Header().Set("Cache-Control", DefaultCacheControl)
		} else {
			w.Header().Set("Cache-Control", t.config.CacheControl)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
		log.DebugRF(r, "%v theme served: %v (%v)", t.name, path, mime)
	})
}