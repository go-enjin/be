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
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/iancoleman/strcase"
	"github.com/leekchan/gtf"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/fs/local"
	"github.com/go-enjin/be/pkg/log"
	bePath "github.com/go-enjin/be/pkg/path"
	"github.com/go-enjin/be/pkg/theme/funcs"
)

type Author struct {
	Name     string
	Homepage string
}

type Config struct {
	Name        string
	License     string
	LicenseLink string
	Description string
	Homepage    string
	Authors     []Author
	Extends     string
	RootStyles  []template.CSS
	BlockStyles map[string][]template.CSS
	BlockThemes map[string]map[string]interface{}
	Context     context.Context
}

type Archetype struct {
}

type Theme struct {
	Name   string
	Path   string
	Config Config
	Minify bool

	FuncMap    template.FuncMap
	Layouts    *Layouts
	Archetypes map[string]*Archetype

	FileSystem fs.FileSystem
}

func New(path string, fs fs.FileSystem) (t *Theme, err error) {
	path = bePath.TrimSlashes(path)
	t = new(Theme)
	t.Path = path
	t.FileSystem = fs
	err = t.init()
	return
}

func NewLocal(path string) (theme *Theme, err error) {
	if !bePath.IsDir(path) {
		err = bePath.ErrorDirNotFound
		return
	}
	theme = new(Theme)
	theme.Path = bePath.TrimSlashes(path)
	if theme.FileSystem, err = local.New(path); err != nil {
		return
	}
	if staticFs, e := local.New(path + "/static"); e == nil {
		fs.RegisterFileSystem("/", staticFs)
	}
	err = theme.init()
	return
}

func (t *Theme) init() (err error) {
	t.Name = bePath.Base(t.Path)
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
	return
}

func (t *Theme) FS() fs.FileSystem {
	return t.FileSystem
}

func (t *Theme) initConfig(ctx context.Context) {
	t.Config = Config{
		Name:        ctx.String("name", t.Name),
		Extends:     ctx.String("extends", ""),
		License:     ctx.String("license", ""),
		LicenseLink: ctx.String("licenselink", ""),
		Description: ctx.String("description", ""),
		Homepage:    ctx.String("homepage", ""),
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

	// if v := ctx.Get("styles"); v != nil {
	// 	rootStyles := make([]string, 0)
	// 	switch vt := v.(type) {
	// 	case map[string]interface{}:
	// 		for section, values := range vt {
	// 			switch section {
	//
	// 			case "color", "primary", "secondary", "accent", "highlight", "alternate", "overlay", "style", "page":
	// 				if subSections, ok := values.(map[string]interface{}); ok {
	// 					for subSection, subSectionValue := range subSections {
	// 						if valueString, ok := subSectionValue.(string); ok {
	// 							key := "--" + section + "--" + subSection
	// 							rootStyles = append(rootStyles, key+": "+valueString+";")
	// 							log.DebugF("%v theme: root style: %v => %v", t.Config.Name, key, valueString)
	// 						} else if subValues, ok := subSectionValue.(map[string]interface{}); ok {
	// 							for subKey, subValue := range subValues {
	// 								if valueString, ok := subValue.(string); ok {
	// 									key := "--" + section + "--" + subSection + "--" + subKey
	// 									rootStyles = append(rootStyles, key+": "+valueString+";")
	// 									log.DebugF("%v theme: root style: %v => %v", t.Config.Name, key, valueString)
	// 								} else {
	// 									log.DebugF("%v theme: unknown root style value type: (%T) %+v", t.Config.Name, subValue, subValue)
	// 								}
	// 							}
	// 						} else {
	// 							log.DebugF("%v theme: unknown root style type: (%T) %+v", t.Config.Name, values, values)
	// 						}
	// 					}
	// 				} else {
	// 					log.DebugF("%v theme: unknown root type: (%T) %+v", t.Config.Name, values, values)
	// 				}
	//
	// 			default:
	// 				log.DebugF("%v theme: unknown root key: %v", t.Config.Name, section)
	//
	// 			}
	// 		}
	// 	default:
	// 		log.ErrorF("%v theme: invalid styles type: (%T) %+v", t.Config.Name, vt, vt)
	// 	}
	// 	sort.Sort(sortorder.Natural(rootStyles))
	// 	t.Config.RootStyles = make([]template.CSS, len(rootStyles))
	// 	for idx, rootStyle := range rootStyles {
	// 		t.Config.RootStyles[idx] = template.CSS(rootStyle)
	// 	}
	// }

	/*
		semantic
			|- styles
			|	|- root:  --*
			|	`- icons: --icon--*--content
			`- themes
				|- primary:  --theme--primary--*
	*/

	if v := ctx.Get("semantic"); v != nil {
		if semantic, ok := v.(map[string]interface{}); ok {
			// log.DebugF("found semantic configuration: %T %+v", v, maps.DebugWalk(semantic))
			var walkStyles func(keys []string, src map[string]interface{}) (styles []string)
			walkStyles = func(keys []string, src map[string]interface{}) (styles []string) {
				for k, s := range src {
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
		log.DebugF("no semantic enjin configuration found")
	}

	// setup context (global variables for use in templates)
	t.Config.Context = context.New()
	for k, v := range ctx {
		switch k {
		case "author", "styles", "semantic":
		default:
			t.Config.Context[k] = v
			log.DebugF("%v theme: adding context: %v => %+v", t.Config.Name, k, v)
		}
	}
	t.Config.Context.CamelizeKeys()
}

func (t *Theme) initFuncMap() {
	t.FuncMap = template.FuncMap{
		"toCamel":              strcase.ToCamel,
		"toLowerCamel":         strcase.ToLowerCamel,
		"toDelimited":          strcase.ToDelimited,
		"toScreamingDelimited": strcase.ToScreamingDelimited,
		"toKebab":              strcase.ToKebab,
		"toScreamingKebab":     strcase.ToScreamingKebab,
		"toSnake":              strcase.ToSnake,
		"toScreamingSnake":     strcase.ToScreamingSnake,

		"asHTML":     funcs.AsHTML,
		"asHTMLAttr": funcs.AsHTMLAttr,
		"asCSS":      funcs.AsCSS,
		"asJS":       funcs.AsJS,
		"fsHash":     funcs.FsHash,
		"fsUrl":      funcs.FsUrl,
		"fsMime":     funcs.FsMime,
		"add":        funcs.Add,
		"sub":        funcs.Sub,

		"mergeClassNames": funcs.MergeClassNames,

		"element":           funcs.Element,
		"elementOpen":       funcs.ElementOpen,
		"elementClose":      funcs.ElementClose,
		"elementAttributes": funcs.ElementAttributes,

		"DebugF": funcs.LogDebug,
		"WarnF":  funcs.LogWarn,
		"ErrorF": funcs.LogError,
	}
	for k, v := range gtf.GtfFuncMap {
		t.FuncMap[k] = v
	}
}

func (t *Theme) initLayouts() (err error) {
	t.Layouts, err = NewLayouts(t)
	return
}

func (t *Theme) initArchetypes() {
	t.Archetypes = make(map[string]*Archetype)
}