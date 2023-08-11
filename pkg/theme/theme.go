// Copyright (c) 2022  The Go-Enjin Authors
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
	"fmt"
	"html/template"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/fvbommel/sortorder"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/net/headers/policy/csp"
	"github.com/go-enjin/be/pkg/net/headers/policy/permissions"
	bePath "github.com/go-enjin/be/pkg/path"
	"github.com/go-enjin/be/pkg/types/theme-types"
)

var (
	DefaultCacheControl = "public, max-age=604800, no-transform, immutable"
)

type Author struct {
	Name     string
	Homepage string
}

type Config struct {
	Name        string
	Parent      string
	License     string
	LicenseLink string
	Description string
	Homepage    string
	Authors     []Author
	Extends     string

	GoogleAnalytics GoogleAnalyticsConfig

	RootStyles  []template.CSS
	BlockStyles map[string][]template.CSS
	BlockThemes map[string]map[string]interface{}

	FontawesomeLinks   map[string]string
	FontawesomeClasses []string

	CacheControl string

	PermissionsPolicy     []permissions.Directive
	ContentSecurityPolicy csp.ContentSecurityPolicyConfig

	Context context.Context
}

type GoogleAnalyticsConfig struct {
	GTM string
}

type Archetype struct {
}

var _ types.Theme = (*Theme)(nil)

type Theme struct {
	Origin string
	Name   string
	Path   string
	Config Config
	Minify bool

	FuncMap    template.FuncMap
	Layouts    *Layouts
	Archetypes map[string]*Archetype

	FileSystem fs.FileSystem
	StaticFS   fs.FileSystem

	FormatProviders []types.FormatProvider
}

func New(origin, path string, themeFs, staticFs fs.FileSystem) (t *Theme, err error) {
	t = new(Theme)
	t.Origin = origin
	t.Name = bePath.Base(path)
	t.Path = path
	t.FileSystem = themeFs
	if staticFs != nil {
		t.StaticFS = staticFs
		fs.RegisterFileSystem("/", staticFs)
	}
	err = t.init()
	return
}

func (t *Theme) init() (err error) {
	t.Layouts = nil
	t.Archetypes = make(map[string]*Archetype)
	if t.FileSystem == nil {
		err = fmt.Errorf(`missing filesystem`)
		return
	}
	var cfg []byte
	if cfg, err = t.FileSystem.ReadFile("theme.toml"); err != nil {
		return
	}
	ctx := context.New()
	if _, err = toml.Decode(string(cfg), &ctx); err != nil {
		return
	}
	t.initConfig(ctx)
	t.initFuncMap()
	if err = t.initLayouts(); err != nil {
		return
	}
	t.initArchetypes()

	addThemeInstance(t)
	return
}

func (t *Theme) Reload() (err error) {
	var nt *Theme
	if nt, err = New(t.Origin, t.Path, t.FileSystem, t.StaticFS); err == nil {
		nt.FormatProviders = t.FormatProviders
		*t = *nt
		if parent := nt.GetParent(); parent != nil {
			err = parent.Reload()
		}
	}
	return
}

func (t *Theme) AddFormatProvider(providers ...types.FormatProvider) {
	t.FormatProviders = append(t.FormatProviders, providers...)
}

func (t *Theme) ListFormats() (names []string) {
	for _, p := range t.FormatProviders {
		names = append(names, p.ListFormats()...)
	}
	sort.Sort(sortorder.Natural(names))
	return
}

func (t *Theme) GetFormat(name string) (format types.Format) {
	for _, provider := range t.FormatProviders {
		if format = provider.GetFormat(name); format != nil {
			return
		}
	}
	return
}

func (t *Theme) MatchFormat(filename string) (format types.Format, match string) {
	for _, provider := range t.FormatProviders {
		if format, match = provider.MatchFormat(filename); format != nil {
			return
		}
	}
	return
}

func (t *Theme) FS() fs.FileSystem {
	return t.FileSystem
}

func (t *Theme) Locales() (locales fs.FileSystem, ok bool) {
	if _, err := t.FileSystem.ReadDir("locales"); err == nil {
		log.DebugF("found %v theme locales", t.Name)
		if locales, err = fs.Wrap("locales", t.FileSystem); err == nil {
			ok = true
		} else {
			log.ErrorF("error wrapping %v theme locales: %v", t.Name, err)
			locales = nil
		}
	}
	return
}

func (t *Theme) initConfig(ctx context.Context) {
	t.Config = Config{
		Name:             ctx.String("name", t.Name),
		Parent:           ctx.String("parent", ""),
		Extends:          ctx.String("extends", ""),
		License:          ctx.String("license", ""),
		LicenseLink:      ctx.String("licenselink", ""),
		Description:      ctx.String("description", ""),
		Homepage:         ctx.String("homepage", ""),
		FontawesomeLinks: make(map[string]string),
		Context:          context.New(),
	}

	if ctx.Has("static") {
		if static, ok := ctx.Get("static").(map[string]interface{}); ok {
			if cacheControl, ok := static["cache-control"].(string); ok {
				t.Config.CacheControl = cacheControl
			}
		}
	}

	t.Config.Authors = make([]Author, 0)
	if ctx.Has("author") {
		v := ctx.Get("author")
		switch value := v.(type) {
		case map[string]interface{}:
			actx := context.NewFromMap(value)
			author := Author{}
			author.Name = actx.String("name", "")
			author.Homepage = actx.String("homepage", "")
			t.Config.Authors = append(t.Config.Authors, author)
		}
	}

	if ctx.Has("google-analytics") {
		if ga, ok := ctx.Get("google-analytics").(map[string]interface{}); ok {
			if gtm, ok := ga["gtm"].(string); ok {
				t.Config.GoogleAnalytics.GTM = gtm
			}
		}
	}

	if ctx.Has("permissions-policy") {
		log.DebugF("permissions-policy theme config unimplemented")
	}

	if ctx.Has("content-security-policy") {
		if ctxCsp, ok := ctx.Get("content-security-policy").(map[string]interface{}); ok {
			// log.WarnF("theme config has content-security-policy - %#+v", ctxCsp)
			var ee error
			if t.Config.ContentSecurityPolicy, ee = csp.ParseContentSecurityPolicyConfig(ctxCsp); ee != nil {
				log.ErrorF("%v theme errors:\n%v", t.Name, ee)
			}
		}
	}

	if ctx.Has("fontawesome") {
		if fa, ok := ctx.Get("fontawesome").(map[string]interface{}); ok {
			for k, v := range fa {
				key := strings.ToLower(k)
				switch key {
				case "classes":
					if vv, ok := v.([]interface{}); ok {
						for _, vvv := range vv {
							if class, ok := vvv.(string); ok {
								t.Config.FontawesomeClasses = append(t.Config.FontawesomeClasses, strings.ToLower(class))
							} else {
								log.ErrorF("error parsing fontawesome config: expected string, found: %T", vvv)
							}
						}
					} else {
						log.ErrorF("error parsing fontawesome config: expected []interface{}, found: %T", vv)
					}
				default:
					if vv, ok := v.(string); ok {
						t.Config.FontawesomeLinks[key] = vv
					} else {
						log.ErrorF("error parsing fontawesome config: expected string, found: %T", v)
					}
				}
			}
		}
	}

	t.Config.Context.SetSpecific("SiteMenuMobileStyle", "side")
	t.Config.Context.SetSpecific("SiteMenuDesktopStyle", "menu")

	if v := ctx.Get("semantic"); v != nil {
		if semantic, ok := v.(map[string]interface{}); ok {
			// log.DebugF("semantic configuration: %T %+v", v, maps.DebugWalk(semantic))

			if siteInfo, ok := semantic["site"].(map[string]interface{}); ok {
				if siteMenu, ok := siteInfo["menu"].(map[string]interface{}); ok {
					if siteMenuMobile, ok := siteMenu["mobile"].(map[string]interface{}); ok {
						if siteMenuMobileStyle, ok := siteMenuMobile["style"].(string); ok {
							t.Config.Context.SetSpecific("SiteMenuMobileStyle", siteMenuMobileStyle)
							log.DebugF("site menu mobile style: %v", siteMenuMobileStyle)
						}
					}
					if siteMenuDesktop, ok := siteMenu["desktop"].(map[string]interface{}); ok {
						if siteMenuDesktopStyle, ok := siteMenuDesktop["style"].(string); ok {
							t.Config.Context.SetSpecific("SiteMenuDesktopStyle", siteMenuDesktopStyle)
							log.DebugF("site menu desktop style: %v", siteMenuDesktopStyle)
						}
					}
				}

				if sitePage, ok := siteInfo["page"].(map[string]interface{}); ok {

					if stylesheets, ok := sitePage["stylesheets"].([]interface{}); ok {
						var found []string
						for _, stylesheet := range stylesheets {
							if href, ok := stylesheet.(string); ok {
								found = append(found, href)
							}
						}
						t.Config.Context.SetSpecific("PageStyleSheets", found)
					}

					if stylesheets, ok := sitePage["font-stylesheets"].([]interface{}); ok {
						var found []string
						for _, stylesheet := range stylesheets {
							if href, ok := stylesheet.(string); ok {
								found = append(found, href)
							}
						}
						t.Config.Context.SetSpecific("PageFontStyleSheets", found)
					}

				}
			}

			var walkStyles func(keys []string, src map[string]interface{}) (styles []string)
			walkStyles = func(keys []string, src map[string]interface{}) (styles []string) {
				sk := maps.SortedKeys(src)
				for _, k := range sk {
					s, _ := src[k]
					switch typedStyle := s.(type) {
					case map[string]interface{}:
						r := walkStyles(append(keys, k), typedStyle)
						styles = append(styles, r...)
					default:
						joined := strings.Join(append(keys, k), "--")
						styles = append(
							styles,
							fmt.Sprintf(
								"--%v: %v;",
								joined,
								typedStyle,
							),
						)
					}
				}
				return
			}

			if style, ok := semantic["style"].(map[string]interface{}); ok {
				// log.DebugF("found semantic styles: %+v", maps.DebugWalk(style))

				keyOrder := []string{
					"color",
					"overlay",
					"primary",
					"secondary",
					"accent",
					"highlight",
					"alternate",
					"style",
					"desktop",
					"mobile",
					"page",
					"z",
					"fa",
					"fa-solid",
					"fa-regular",
					"fa-brands",
					"icon",
					"theme",
				}

				var rootStyles []string
				for _, key := range keyOrder {
					if keyVal, ok := style[key].(map[string]interface{}); ok {
						results := walkStyles([]string{key}, keyVal)
						if len(results) > 0 {
							rootStyles = append(rootStyles, results...)
						}
					}
				}
				t.Config.RootStyles = make([]template.CSS, len(rootStyles))
				for idx, rootStyle := range rootStyles {
					t.Config.RootStyles[idx] = template.CSS(rootStyle)
				}
			}

			if block, ok := semantic["block"].(map[string]interface{}); ok {
				t.Config.BlockStyles = make(map[string][]template.CSS)
				t.Config.BlockThemes = make(map[string]map[string]interface{})

				if blockThemes, ok := block["theme"].(map[string]interface{}); ok {
					for k, vv := range blockThemes {
						if blockTheme, ok := vv.(map[string]interface{}); ok {
							results := walkStyles([]string{"style"}, blockTheme)
							resultStyles := make([]template.CSS, len(results))
							for idx, result := range results {
								resultStyles[idx] = template.CSS(result)
							}
							t.Config.BlockThemes[k] = blockTheme
							t.Config.BlockStyles[k] = resultStyles
						}
					}
				}

			}

		} else {
			log.ErrorF("semantic structure is not a map[string]interface{}: %T", v)
		}
	} else {
		// log.DebugF("no semantic enjin configuration found")
	}

	for k, v := range ctx {
		switch k {
		case "author", "styles", "semantic":
		default:
			t.Config.Context[k] = v
			// log.DebugF("%v theme: adding context: %v => %+v", t.Config.Name, k, v)
		}
	}
	t.Config.Context.CamelizeKeys()
}

func (t *Theme) GetConfig() (config Config) {
	config = Config{
		Name:                  t.Config.Name,
		Parent:                t.Config.Parent,
		License:               t.Config.License,
		LicenseLink:           t.Config.LicenseLink,
		Description:           t.Config.Description,
		Homepage:              t.Config.Homepage,
		Authors:               t.Config.Authors,
		Extends:               t.Config.Extends,
		FontawesomeLinks:      make(map[string]string),
		GoogleAnalytics:       t.Config.GoogleAnalytics,
		PermissionsPolicy:     t.Config.PermissionsPolicy,
		ContentSecurityPolicy: t.Config.ContentSecurityPolicy,
	}

	config.BlockStyles = make(map[string][]template.CSS)
	config.BlockThemes = make(map[string]map[string]interface{})

	if parent := t.GetParent(); parent != nil {

		config.RootStyles = append(
			parent.Config.RootStyles,
			t.Config.RootStyles...,
		)

		for k, v := range parent.Config.BlockStyles {
			config.BlockStyles[k] = append([]template.CSS{}, v...)
		}

		for k, v := range parent.Config.BlockThemes {
			config.BlockThemes[k] = make(map[string]interface{})
			for j, vv := range v {
				config.BlockThemes[k][j] = vv
			}
		}

		for _, v := range parent.Config.FontawesomeClasses {
			config.FontawesomeClasses = append(config.FontawesomeClasses, v)
		}
		for k, v := range parent.Config.FontawesomeLinks {
			config.FontawesomeLinks[k] = v
		}

		config.Context = parent.Config.Context.Copy()
	} else {
		config.RootStyles = t.Config.RootStyles
		config.Context = context.New()
	}

	config.Context.Apply(t.Config.Context)

	for k, v := range t.Config.BlockStyles {
		config.BlockStyles[k] = append([]template.CSS{}, v...)
	}
	for k, v := range t.Config.BlockThemes {
		config.BlockThemes[k] = make(map[string]interface{})
		for j, vv := range v {
			config.BlockThemes[k][j] = vv
		}
	}

	return
}

func (t *Theme) GetBlockThemeNames() (names []string) {
	names = append(names, "primary", "secondary")
	for k, _ := range t.GetConfig().BlockThemes {
		names = append(names, k)
	}
	return
}

func (t *Theme) initFuncMap() {
	t.FuncMap = DefaultFuncMap()
}

func (t *Theme) initLayouts() (err error) {
	t.Layouts, err = NewLayouts(t)
	return
}

func (t *Theme) initArchetypes() {
	t.Archetypes = make(map[string]*Archetype)
}