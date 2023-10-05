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
	"github.com/go-enjin/golang-org-x-text/language"
)

func (t *CTheme) init() (err error) {

	var ctx context.Context
	if ctx, err = t.readToml(); err != nil {
		return
	}

	t.layouts, err = layouts.NewLayouts(t)
	t.config = t.makeConfig(ctx)

	return
}

func (t *CTheme) readToml() (ctx context.Context, err error) {
	if t.fs == nil {
		err = fmt.Errorf(`missing filesystem`)
		return
	}

	var cfg []byte
	ctx = context.New()
	if cfg, err = t.fs.ReadFile("theme.toml"); err != nil {
		return
	} else if _, err = toml.Decode(string(cfg), &ctx); err != nil {
		return
	}
	return
}

func (t *CTheme) makeConfig(ctx context.Context) (config *feature.ThemeConfig) {
	config = &feature.ThemeConfig{
		Name:             ctx.String("name", t.Name()),
		Parent:           ctx.String("parent", ""),
		Extends:          ctx.String("extends", ""),
		License:          ctx.String("license", ""),
		LicenseLink:      ctx.String("licenselink", ""),
		Description:      ctx.String("description", ""),
		Homepage:         ctx.String("homepage", ""),
		FontawesomeLinks: make(map[string]string),
		Context:          context.New(),
		Supports: feature.ThemeSupports{
			Archetypes: map[string]context.Fields{},
		},
	}

	if ctx.Has("supports") {
		if v, ok := ctx.Get("supports").(map[string]interface{}); ok {
			supCtx := context.Context(v)

			config.Supports.Menus = feature.ParseMenuSupports(supCtx.Get("menus"))

			if codes, ok := supCtx.Get("locales").([]string); ok {
				for _, code := range codes {
					if tag, err := language.Parse(code); err != nil {
						log.ErrorF("%v theme.toml error - invalid language code: \"%v\"", config.Name, code)
					} else {
						config.Supports.Locales = append(config.Supports.Locales, tag)
					}
				}
			}

			if len(config.Supports.Locales) == 0 {
				config.Supports.Locales = []language.Tag{language.English}
			}

			if archetypes, ok := supCtx.Get("archetypes").(map[string]interface{}); ok {
				for archetype, vv := range archetypes {
					if list, ok := vv.(map[string]interface{}); ok {
						for _, vvv := range list {
							if item, ok := vvv.(map[string]interface{}); ok {
								field := context.ParseField(item)
								if _, present := config.Supports.Archetypes[archetype]; !present {
									config.Supports.Archetypes[archetype] = context.Fields{}
								}
								config.Supports.Archetypes[archetype][field.Key] = field
							}
						}
					}
				}
			}
		}
	}

	if ctx.Has("static") {
		if static, ok := ctx.Get("static").(map[string]interface{}); ok {
			if cacheControl, ok := static["cache-control"].(string); ok {
				config.CacheControl = cacheControl
			}
		}
	}

	config.Authors = make([]feature.Author, 0)
	if ctx.Has("author") {
		v := ctx.Get("author")
		switch value := v.(type) {
		case map[string]interface{}:
			actx := context.NewFromMap(value)
			author := feature.Author{}
			author.Name = actx.String("name", "")
			author.Homepage = actx.String("homepage", "")
			config.Authors = append(config.Authors, author)
		}
	}

	if ctx.Has("google-analytics") {
		if ga, ok := ctx.Get("google-analytics").(map[string]interface{}); ok {
			if gtm, ok := ga["gtm"].(string); ok {
				_ = config.Context.SetKV(".GoogleAnalytics.GTM", gtm)
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
			if config.ContentSecurityPolicy, ee = csp.ParseContentSecurityPolicyConfig(ctxCsp); ee != nil {
				log.ErrorF("%v theme errors:\n%v", t.Name(), ee)
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
								config.FontawesomeClasses = append(config.FontawesomeClasses, strings.ToLower(class))
							} else {
								log.ErrorF("error parsing fontawesome config: expected string, found: %T", vvv)
							}
						}
					} else {
						log.ErrorF("error parsing fontawesome config: expected []interface{}, found: %T", vv)
					}
				default:
					if vv, ok := v.(string); ok {
						config.FontawesomeLinks[key] = vv
					} else {
						log.ErrorF("error parsing fontawesome config: expected string, found: %T", v)
					}
				}
			}
		}
	}

	config.Context.SetSpecific("SiteMenuMobileStyle", "side")
	config.Context.SetSpecific("SiteMenuDesktopStyle", "menu")

	if v := ctx.Get("semantic"); v != nil {
		if semantic, ok := v.(map[string]interface{}); ok {
			// log.DebugF("semantic configuration: %T %+v", v, maps.DebugWalk(semantic))

			if siteInfo, ok := semantic["site"].(map[string]interface{}); ok {
				if siteMenu, ok := siteInfo["menu"].(map[string]interface{}); ok {
					if siteMenuMobile, ok := siteMenu["mobile"].(map[string]interface{}); ok {
						if siteMenuMobileStyle, ok := siteMenuMobile["style"].(string); ok {
							config.Context.SetSpecific("SiteMenuMobileStyle", siteMenuMobileStyle)
							//log.DebugF("site menu mobile style: %v", siteMenuMobileStyle)
						}
					}
					if siteMenuDesktop, ok := siteMenu["desktop"].(map[string]interface{}); ok {
						if siteMenuDesktopStyle, ok := siteMenuDesktop["style"].(string); ok {
							config.Context.SetSpecific("SiteMenuDesktopStyle", siteMenuDesktopStyle)
							//log.DebugF("site menu desktop style: %v", siteMenuDesktopStyle)
						}
					}
				}

				if sitePage, ok := siteInfo["page"].(map[string]interface{}); ok {

					if stylesheets, ok := sitePage["early-stylesheets"].([]interface{}); ok {
						var found []string
						for _, stylesheet := range stylesheets {
							if href, ok := stylesheet.(string); ok {
								found = append(found, href)
							}
						}
						config.Context.SetSpecific("PageEarlyStyleSheets", found)
					}

					if stylesheets, ok := sitePage["stylesheets"].([]interface{}); ok {
						var found []string
						for _, stylesheet := range stylesheets {
							if href, ok := stylesheet.(string); ok {
								found = append(found, href)
							}
						}
						config.Context.SetSpecific("PageStyleSheets", found)
					}

					if stylesheets, ok := sitePage["font-stylesheets"].([]interface{}); ok {
						var found []string
						for _, stylesheet := range stylesheets {
							if href, ok := stylesheet.(string); ok {
								found = append(found, href)
							}
						}
						config.Context.SetSpecific("PageFontStyleSheets", found)
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
				config.RootStyles = make([]template.CSS, len(rootStyles))
				for idx, rootStyle := range rootStyles {
					config.RootStyles[idx] = template.CSS(rootStyle)
				}
			}

			if block, ok := semantic["block"].(map[string]interface{}); ok {
				config.BlockStyles = make(map[string][]template.CSS)
				config.BlockThemes = make(map[string]map[string]interface{})

				if blockThemes, ok := block["theme"].(map[string]interface{}); ok {
					for k, vv := range blockThemes {
						if blockTheme, ok := vv.(map[string]interface{}); ok {
							results := walkStyles([]string{"style"}, blockTheme)
							resultStyles := make([]template.CSS, len(results))
							for idx, result := range results {
								resultStyles[idx] = template.CSS(result)
							}
							config.BlockThemes[k] = blockTheme
							config.BlockStyles[k] = resultStyles
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
			config.Context[k] = v
			// log.DebugF("%v theme: adding context: %v => %+v", t.ThemeConfig.Name, k, v)
		}
	}

	config.Context.CamelizeKeys()
	return
}