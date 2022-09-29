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

	"github.com/BurntSushi/toml"
	"github.com/iancoleman/strcase"
	"github.com/leekchan/gtf"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/log"
	bePath "github.com/go-enjin/be/pkg/path"
	"github.com/go-enjin/be/pkg/tmpl"
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
	Layouts    map[string]*Layout
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

func (t *Theme) init() (err error) {
	t.Name = bePath.Base(t.Path)
	t.Layouts = make(map[string]*Layout)
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
	t.initContext(ctx)
	t.initFuncMap()
	if err = t.initLayouts(); err != nil {
		return
	}
	t.initArchetypes()
	return
}

func (t *Theme) initConfig(ctx context.Context) {
	t.Config = Config{}
	t.Config.Name = ctx.String("Name", t.Name)
	t.Config.Extends = ctx.String("Extends", "")
	t.Config.License = ctx.String("License", "")
	t.Config.LicenseLink = ctx.String("LicenseLink", "")
	t.Config.Description = ctx.String("Description", "")
	t.Config.Homepage = ctx.String("Homepage", "")
	t.Config.Authors = make([]Author, 0)
	if ctx.Has("author") {
		v := ctx.Get("author")
		switch value := v.(type) {
		case map[string]interface{}:
			actx := context.NewFromMap(value)
			author := Author{}
			author.Name = actx.String("Name", "")
			author.Homepage = actx.String("Homepage", "")
			t.Config.Authors = append(t.Config.Authors, author)
		}
	}
}

func (t *Theme) initContext(ctx context.Context) {
	// setup context (global variables for use in templates)
	t.Config.Context = context.New()
	for k, v := range ctx {
		t.Config.Context[k] = v
	}
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

		"asHTML":     tmpl.AsHTML,
		"asHTMLAttr": tmpl.AsHTMLAttr,
		"fsHash":     tmpl.FsHash,
		"fsUrl":      tmpl.FsUrl,
		"fsMime":     tmpl.FsMime,
		"add":        tmpl.Add,
		"sub":        tmpl.Sub,

		"element":           tmpl.Element,
		"elementOpen":       tmpl.ElementOpen,
		"elementClose":      tmpl.ElementClose,
		"elementAttributes": tmpl.ElementAttributes,
	}
	for k, v := range gtf.GtfFuncMap {
		t.FuncMap[k] = v
	}
}

func (t *Theme) initLayouts() (err error) {
	t.Layouts = make(map[string]*Layout)
	var paths []string
	if paths, err = t.FileSystem.ListDirs("layouts"); err != nil {
		err = fmt.Errorf("error listing layouts: %v", err)
		return
	}

	for _, path := range paths {
		var l *Layout
		if lo, e := NewLayout(path, t.FileSystem, t.FuncMap); e != nil {
			err = fmt.Errorf("error creating new layout: %v - %v", path, e)
			return
		} else {
			l = lo
			log.DebugF("including %v theme layout: %v", t.Name, path)
		}
		t.Layouts[bePath.Base(path)] = l
		var files []string
		for _, t := range l.Tmpl.Templates() {
			files = append(files, t.Name())
		}
		log.DebugF("included %v theme %v layout templates: %v", t.Name, l.Name, files)
	}

	if partials, ok := t.Layouts["partials"]; ok {
		for k, layout := range t.Layouts {
			if k == "partials" {
				continue
			}
			for _, tmpl := range partials.Tmpl.Templates() {
				if _, err = layout.Tmpl.AddParseTree(tmpl.Name(), tmpl.Tree); err != nil {
					return
				}
			}
		}
	}
	if defaultLayout, ok := t.Layouts["_default"]; ok {
		for k, layout := range t.Layouts {
			switch k {
			case "partials", "_default":
				continue
			}
			for _, tmpl := range defaultLayout.Tmpl.Templates() {
				if _, err = layout.Tmpl.AddParseTree(tmpl.Name(), tmpl.Tree); err != nil {
					return
				}
			}
		}
	}
	return
}

func (t *Theme) initArchetypes() {
	t.Archetypes = make(map[string]*Archetype)
}