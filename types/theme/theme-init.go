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
	"fmt"
	"html/template"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/net/headers/policy/csp"
	"github.com/go-enjin/be/types/theme/layouts"
)

func (t *Theme) init() (err error) {
	if t.fs == nil {
		err = fmt.Errorf(`missing filesystem`)
		return
	}

	var cfg []byte
	ctx := context.New()
	if cfg, err = t.fs.ReadFile("theme.toml"); err != nil {
		return
	} else if _, err = toml.Decode(string(cfg), &ctx); err != nil {
		return
	}

	t.initConfig(ctx)
	t.layouts, err = layouts.NewLayouts(t)

	return
}

func (t *Theme) initConfig(ctx context.Context) {
	t.config = feature.ThemeConfig{
		Name:             ctx.String("name", t.name),
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
				t.config.CacheControl = cacheControl
			}
		}
	}

	t.config.Authors = make([]feature.Author, 0)
	if ctx.Has("author") {
		v := ctx.Get("author")
		switch value := v.(type) {
		case map[string]interface{}:
			actx := context.NewFromMap(value)
			author := feature.Author{}
			author.Name = actx.String("name", "")
			author.Homepage = actx.String("homepage", "")
			t.config.Authors = append(t.config.Authors, author)
		}
	}

	if ctx.Has("google-analytics") {
		if ga, ok := ctx.Get("google-analytics").(map[string]interface{}); ok {
			if gtm, ok := ga["gtm"].(string); ok {
				_ = t.config.Context.SetKV(".GoogleAnalytics.GTM", gtm)
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
			if t.config.ContentSecurityPolicy, ee = csp.ParseContentSecurityPolicyConfig(ctxCsp); ee != nil {
				log.ErrorF("%v theme errors:\n%v", t.name, ee)
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
								t.config.FontawesomeClasses = append(t.config.FontawesomeClasses, strings.ToLower(class))
							} else {
								log.ErrorF("error parsing fontawesome config: expected string, found: %T", vvv)
							}
						}
					} else {
						log.ErrorF("error parsing fontawesome config: expected []interface{}, found: %T", vv)
					}
				default:
					if vv, ok := v.(string); ok {
						t.config.FontawesomeLinks[key] = vv
					} else {
						log.ErrorF("error parsing fontawesome config: expected string, found: %T", v)
					}
				}
			}
		}
	}

	t.config.Context.SetSpecific("SiteMenuMobileStyle", "side")
	t.config.Context.SetSpecific("SiteMenuDesktopStyle", "menu")

	if v := ctx.Get("semantic"); v != nil {
		if semantic, ok := v.(map[string]interface{}); ok {
			// log.DebugF("semantic configuration: %T %+v", v, maps.DebugWalk(semantic))

			if siteInfo, ok := semantic["site"].(map[string]interface{}); ok {
				if siteMenu, ok := siteInfo["menu"].(map[string]interface{}); ok {
					if siteMenuMobile, ok := siteMenu["mobile"].(map[string]interface{}); ok {
						if siteMenuMobileStyle, ok := siteMenuMobile["style"].(string); ok {
							t.config.Context.SetSpecific("SiteMenuMobileStyle", siteMenuMobileStyle)
							log.DebugF("site menu mobile style: %v", siteMenuMobileStyle)
						}
					}
					if siteMenuDesktop, ok := siteMenu["desktop"].(map[string]interface{}); ok {
						if siteMenuDesktopStyle, ok := siteMenuDesktop["style"].(string); ok {
							t.config.Context.SetSpecific("SiteMenuDesktopStyle", siteMenuDesktopStyle)
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
						t.config.Context.SetSpecific("PageStyleSheets", found)
					}

					if stylesheets, ok := sitePage["font-stylesheets"].([]interface{}); ok {
						var found []string
						for _, stylesheet := range stylesheets {
							if href, ok := stylesheet.(string); ok {
								found = append(found, href)
							}
						}
						t.config.Context.SetSpecific("PageFontStyleSheets", found)
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
				t.config.RootStyles = make([]template.CSS, len(rootStyles))
				for idx, rootStyle := range rootStyles {
					t.config.RootStyles[idx] = template.CSS(rootStyle)
				}
			}

			if block, ok := semantic["block"].(map[string]interface{}); ok {
				t.config.BlockStyles = make(map[string][]template.CSS)
				t.config.BlockThemes = make(map[string]map[string]interface{})

				if blockThemes, ok := block["theme"].(map[string]interface{}); ok {
					for k, vv := range blockThemes {
						if blockTheme, ok := vv.(map[string]interface{}); ok {
							results := walkStyles([]string{"style"}, blockTheme)
							resultStyles := make([]template.CSS, len(results))
							for idx, result := range results {
								resultStyles[idx] = template.CSS(result)
							}
							t.config.BlockThemes[k] = blockTheme
							t.config.BlockStyles[k] = resultStyles
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
			t.config.Context[k] = v
			// log.DebugF("%v theme: adding context: %v => %+v", t.ThemeConfig.Name, k, v)
		}
	}
	t.config.Context.CamelizeKeys()
}