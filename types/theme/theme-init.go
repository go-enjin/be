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
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/maruel/natural"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/net/headers/policy/csp"
	"github.com/go-enjin/be/pkg/slices"
	"github.com/go-enjin/be/types/theme/layouts"
)

func (t *CTheme) init() (err error) {

	var ctx context.Context
	if ctx, err = t.readToml(); err != nil {
		return
	}
	t.tomlCache = ctx.Copy()

	t.config = t.makeConfig(ctx)
	t.parent = t.config.Parent
	t.layouts, err = layouts.NewLayouts(t)

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

	config.Authors = make([]feature.ThemeAuthor, 0)
	if ctx.Has("author") {
		v := ctx.Get("author")
		switch value := v.(type) {
		case map[string]interface{}:
			authorCtx := context.Context(value)
			author := feature.ThemeAuthor{
				Name:     authorCtx.String("name", ""),
				Homepage: authorCtx.String("homepage", ""),
			}
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
		log.ErrorF("permissions-policy theme config unimplemented")
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

			if siteInfo, ok := semantic["site"].(map[string]interface{}); ok {
				if siteMenu, ok := siteInfo["menu"].(map[string]interface{}); ok {
					if siteMenuMobile, ok := siteMenu["mobile"].(map[string]interface{}); ok {
						if siteMenuMobileStyle, ok := siteMenuMobile["style"].(string); ok {
							config.Context.SetSpecific("SiteMenuMobileStyle", siteMenuMobileStyle)
						}
					}
					if siteMenuDesktop, ok := siteMenu["desktop"].(map[string]interface{}); ok {
						if siteMenuDesktopStyle, ok := siteMenuDesktop["style"].(string); ok {
							config.Context.SetSpecific("SiteMenuDesktopStyle", siteMenuDesktopStyle)
						}
					}
				}

				if sitePage, ok := siteInfo["page"].(map[string]interface{}); ok {

					if found := t.parseListOfStrings(sitePage["early-stylesheets"]); len(found) > 0 {
						config.Context.SetSpecific("PageEarlyStyleSheets", found)
					}

					if found := t.parseListOfStrings(sitePage["stylesheets"]); len(found) > 0 {
						config.Context.SetSpecific("PageStyleSheets", found)
					}

					if found := t.parseListOfStrings(sitePage["font-stylesheets"]); len(found) > 0 {
						config.Context.SetSpecific("PageFontStyleSheets", found)
					}

				}
			}

			if rootStyles, ok := semantic["style"].(map[string]interface{}); ok {
				config.RootStyles = t.parseSemanticStyles(nil, rootStyles)
			}

			if block, ok := semantic["block"].(map[string]interface{}); ok {
				config.BlockStyles = make(map[string][]template.CSS)
				config.BlockThemes = make(map[string]map[string]interface{})

				if blockThemes, ok := block["theme"].(map[string]interface{}); ok {
					for k, vv := range blockThemes {
						if blockTheme, ok := vv.(map[string]interface{}); ok {
							parsedStyles := t.parseSemanticStyles([]string{"style"}, blockTheme)
							config.BlockThemes[k] = blockTheme
							config.BlockStyles[k] = parsedStyles
						}
					}
				}

			}

		} else {
			log.ErrorF("semantic structure is not a map[string]interface{}: %T", v)
		}
	}

	for k, v := range ctx {
		switch k {
		case "author", "styles", "semantic":
		default:
			config.Context[k] = v
		}
	}

	config.Context.CamelizeKeys()
	return
}

func (t *CTheme) parseListOfStrings(input interface{}) (found []string) {
	if stylesheets, ok := input.([]interface{}); ok {
		for _, stylesheet := range stylesheets {
			if value, ok := stylesheet.(string); ok {
				found = append(found, value)
			}
		}
	}
	return
}

var (
	SemanticStylesOrder = []string{
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
)

func (t *CTheme) parseSemanticStyles(keys []string, src map[string]interface{}) (parsed []template.CSS) {
	for _, styleKey := range t.sortedSemanticStyleKeys(len(keys), src) {
		switch styleValue := src[styleKey].(type) {
		case context.Context:
			more := t.parseSemanticStyles(append(keys, styleKey), styleValue)
			parsed = append(parsed, more...)
		case map[string]interface{}:
			more := t.parseSemanticStyles(append(keys, styleKey), styleValue)
			parsed = append(parsed, more...)
		default:
			joined := strings.Join(append(keys, styleKey), "--")
			parsed = append(parsed, template.CSS(fmt.Sprintf("--%v: %v;", joined, styleValue)))
		}
	}
	return
}

func (t *CTheme) sortedSemanticStyleKeys(depth int, src map[string]interface{}) (sorted []string) {
	sorted = maps.Keys(src)

	/*

		Depth of Zero means that the keys in the map are actually the semantic style prefixes, which are to be sorted
		specially:

			An,Bn: natural.Less
			An,By: less = true
			Ay,Bn: less = false
			Ay,By: named order

		Depth greater than zero means that the keys are building up the actual CSS root style name and are sorted naturally.
	*/

	sort.Slice(sorted, func(i, j int) (less bool) {
		a, b := sorted[i], sorted[j]
		if depth > 0 {
			less = natural.Less(a, b)
			return
		}

		aIsNamed, bIsNamed := slices.Within(a, SemanticStylesOrder), slices.Within(b, SemanticStylesOrder)

		if !aIsNamed && !bIsNamed {
			less = natural.Less(a, b)
			return
		} else if !aIsNamed && bIsNamed {
			less = true
			return
		} else if aIsNamed && !bIsNamed {
			less = false
			return
		}

		idx, jdx := slices.IndexOf(SemanticStylesOrder, a), slices.IndexOf(SemanticStylesOrder, b)
		less = idx < jdx
		return
	})
	return
}
