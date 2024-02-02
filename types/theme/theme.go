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
	"net/http"
	"sort"
	"sync"

	"github.com/maruel/natural"

	clPath "github.com/go-corelibs/path"
	"github.com/go-corelibs/slices"
	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/log"
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
	parent string
	config *feature.ThemeConfig

	minify bool

	layouts *layouts.Layouts

	fs       fs.FileSystem
	staticFS fs.FileSystem

	autoload bool

	formatProviders []feature.PageFormatProvider

	tomlCache beContext.Context

	sync.RWMutex
}

func New(origin, path string, themeFs, staticFs fs.FileSystem, autoload bool) (ft feature.Theme, err error) {
	ft, err = newTheme(origin, path, themeFs, staticFs, autoload)
	return
}

func newTheme(origin, path string, themeFs, staticFs fs.FileSystem, autoload bool) (t *CTheme, err error) {
	t = new(CTheme)
	t.origin = origin
	t.name = clPath.Base(path)
	t.path = path
	t.fs = themeFs
	if staticFs != nil {
		t.staticFS = staticFs
	}
	t.autoload = autoload
	if err = t.init(); err != nil {
		return
	}
	t.registerTheme()
	return
}

func (t *CTheme) Name() string {
	//t.RLock()
	//defer t.RUnlock()
	return t.name
}

func (t *CTheme) ThemeFS() fs.FileSystem {
	//t.RLock()
	//defer t.RUnlock()
	return t.fs
}

func (t *CTheme) StaticFS() fs.FileSystem {
	//t.RLock()
	//defer t.RUnlock()
	return t.staticFS
}

func (t *CTheme) Layouts() feature.ThemeLayouts {
	//t.RLock()
	//defer t.RUnlock()
	if t.autoload {
		if l, err := layouts.NewLayouts(t); err == nil {
			return l
		} else {
			log.ErrorF("error autoloading new layouts: %v", err)
		}
	}
	return t.layouts
}

func (t *CTheme) GetConfig() (config *feature.ThemeConfig) {

	if t.autoload {
		if ctx, err := t.readToml(); err != nil {
			log.ErrorF("error autoloading theme.toml: %v", err)
			return
		} else {
			config = t.makeConfig(ctx)
		}
	} else {
		config = t.makeConfig(t.tomlCache.Copy())
	}
	if t.parent != config.Parent {
		t.Lock()
		t.parent = config.Parent
		t.Unlock()
	}

	//t.RLock()
	//defer t.RUnlock()

	if parent := t.GetParent(); parent != nil {

		parentConfig := parent.GetConfig()

		config.RootStyles = slices.Merge(config.RootStyles, parentConfig.RootStyles)

		for k, v := range parentConfig.BlockStyles {
			config.BlockStyles[k] = slices.Merge(config.BlockStyles[k], v)
		}

		for k, v := range parentConfig.BlockThemes {
			if _, present := config.BlockThemes[k]; !present {
				// don't clobber the child theme
				config.BlockThemes[k] = make(map[string]interface{})
				for j, vv := range v {
					config.BlockThemes[k][j] = vv
				}
			}
		}

		config.FontawesomeClasses = slices.Merge(config.FontawesomeClasses, parentConfig.FontawesomeClasses)
		for k, v := range parentConfig.FontawesomeLinks {
			if _, present := config.FontawesomeLinks[k]; !present {
				config.FontawesomeLinks[k] = v
			}
		}

		for _, key := range []string{"SiteMenuMobileStyle", "SiteMenuDesktopStyle"} {
			if v := config.Context.String(key, ""); v == "" {
				if v = parentConfig.Context.String(key, ""); v != "" {
					config.Context.SetSpecific(key, v)
				}
			}
		}
		for _, key := range []string{"PageEarlyStyleSheets", "PageStyleSheets", "PageFontStyleSheets"} {
			list := config.Context.Strings(key)
			if pv := parentConfig.Context.Strings(key); len(pv) > 0 {
				for _, pvv := range pv {
					if !slices.Within(pvv, list) {
						list = append(list, pvv)
					}
				}
			}
			config.Context.SetSpecific(key, list)
		}

		config.Supports.Menus = parentConfig.Supports.Menus.Append(config.Supports.Menus)
		config.Supports.Locales = slices.Merge(config.Supports.Locales, parentConfig.Supports.Locales)
		config.Supports.Layouts = slices.Merge(config.Supports.Layouts, parentConfig.Supports.Layouts)

		for pk, pv := range parentConfig.Supports.Archetypes {
			if kv, ok := config.Supports.Archetypes[pk]; ok {
				for pfk, pfv := range pv {
					if _, present := kv[pfk]; !present {
						config.Supports.Archetypes[pk][pfk] = pfv
					}
				}
			} else {
				config.Supports.Archetypes[pk] = pv
			}
		}

		config.ContentSecurityPolicy = config.ContentSecurityPolicy.Merge(parentConfig.ContentSecurityPolicy)
	}

	if l := t.Layouts(); l != nil {
		config.Supports.Layouts = l.ListLayouts()
	}

	config.Context.CamelizeKeys()
	return
}

func (t *CTheme) GetBlockThemeNames() (names []string) {
	names = append(names, "primary", "secondary")
	for k := range t.GetConfig().BlockThemes {
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
	sort.Sort(natural.StringSlice(names))
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
		log.TraceDF(1, "found %v theme locales", t.Name())
		if locales, err = fs.Wrap("locales", t.fs); err == nil {
			ok = true
		} else {
			log.ErrorF("error wrapping %v theme locales: %v", t.Name(), err)
			locales = nil
		}
	}
	return
}

func (t *CTheme) FindLayout(named string) (layout feature.ThemeLayout, name string, ok bool) {
	if named == "" {
		named = globals.DefaultThemeLayoutName
	}
	name = named

	layout = t.Layouts().GetLayout(name)
	if ok = layout != nil; ok {
		if t.parent == "" {
			log.TraceF("found layout in %v (theme) context: %v", t.Name(), name)
		} else {
			log.TraceF("found layout in %v (%v) context: %v", t.Name(), t.parent, name)
		}
	}
	return
}

func (t *CTheme) ReadStaticFile(path string) (data []byte, mime string, err error) {
	if t.staticFS != nil {
		if data, err = t.staticFS.ReadFile(path); err == nil {
			mime, _ = t.staticFS.MimeType(path)
			return
		}
	}
	if t.parent != "" {
		if parent := t.GetParent(); parent != nil {
			data, mime, err = parent.ReadStaticFile(path)
		}
	}
	return
}

func (t *CTheme) Middleware(next http.Handler) http.Handler {
	log.DebugF("including %v theme static middleware", t.Name())
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if t.staticFS == nil {
			if parent := t.GetParent(); parent == nil {
				next.ServeHTTP(w, r)
				return
			}
		}
		path := clPath.TrimSlashes(r.URL.Path)
		var err error
		var data []byte
		var mime string
		if data, mime, err = t.ReadStaticFile(path); err != nil {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Type", mime)
		if t.config.CacheControl == "" {
			w.Header().Set("Cache-Control", DefaultCacheControl)
		} else {
			w.Header().Set("Cache-Control", t.config.CacheControl)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
		log.DebugRF(r, "%v theme served: %v (%v)", t.Name(), path, mime)
	})
}
