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

package partials

import (
	"bytes"
	"fmt"
	htmlTemplate "html/template"

	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/slices"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "page-partials"

type Feature interface {
	feature.Feature
	feature.FuncMapProvider
	feature.TemplatePartialsProvider
}

type MakeFeature interface {
	Make() Feature

	Set(block, position, name, tmpl string) MakeFeature
}

type CFeature struct {
	feature.CFeature

	// partials tracks .Block.Position.Name = tmpl
	partials map[string]map[string]map[string]string

	// orders tracks .Block.Position = []names
	orders map[string]map[string][]string
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.FeatureTag = tag
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.partials = map[string]map[string]map[string]string{
		"head": {
			"head": {},
			"tail": {},
		},
		"body": {
			"head": {},
			"tail": {},
		},
	}
	f.orders = map[string]map[string][]string{
		"head": {
			"head": {},
			"tail": {},
		},
		"body": {
			"head": {},
			"tail": {},
		},
	}
	return
}

func (f *CFeature) Set(block, position, name, tmpl string) MakeFeature {
	if err := f.RegisterTemplatePartial(block, position, name, tmpl); err != nil {
		log.FatalDF(1, "error registering partial: %v", err)
	}
	return f
}

func (f *CFeature) Make() (feat Feature) {
	return f
}

func (f *CFeature) Tag() (tag feature.Tag) {
	tag = Tag
	return
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	return
}

func (f *CFeature) Shutdown() {

}

func (f *CFeature) isValidArgs(block, position string) (valid bool) {
	valid = slices.Present(block, "head", "body") &&
		slices.Present(position, "head", "tail")
	return
}

func (f *CFeature) RegisterTemplatePartial(block, position, name, tmpl string) (err error) {
	if !f.isValidArgs(block, position) {
		err = fmt.Errorf(`block must be one of "head" or "body" and position must be one of "head" or "tail"`)
		return
	}
	if _, exists := f.partials[block][position][name]; exists {
		log.DebugDF(1, "overriding feature partials: %s.%s.%s", block, position, name)
	}
	f.partials[block][position][name] = tmpl
	f.orders[block][position] = append(f.orders[block][position], name)
	return
}

func (f *CFeature) ListTemplatePartials(block, position string) (names []string) {
	if !f.isValidArgs(block, position) {
		log.FatalDF(1, `block must be one of "head" or "body" and position must be one of "head" or "tail"`)
		return
	}
	names = f.orders[block][position]
	return
}

func (f *CFeature) GetTemplatePartial(block, position, name string) (tmpl string, ok bool) {
	if !f.isValidArgs(block, position) {
		log.FatalDF(1, `block must be one of "head" or "body" and position must be one of "head" or "tail"`)
		return
	}
	tmpl, ok = f.partials[block][position][name]
	return
}

func (f *CFeature) MakeFuncMap(ctx beContext.Context) (fm feature.FuncMap) {
	fm = feature.FuncMap{
		"featurePartials": func(block, position string) (output string) {
			block = strcase.ToKebab(block)
			position = strcase.ToKebab(position)
			baseName := block + "-" + position + "_"
			innerFM := f.Enjin.MakeFuncMap(ctx).AsHTML()

			for _, name := range f.Enjin.ListTemplatePartials(block, position) {
				tmpl, _ := f.Enjin.GetTemplatePartial(block, position, name)

				if tt, err := htmlTemplate.New(baseName + name).Funcs(innerFM).Parse(tmpl); err != nil {
					log.ErrorF("error parsing feature partials template: %v - %v", name, err)
				} else {
					var buf bytes.Buffer
					if err = tt.Execute(&buf, ctx); err != nil {
						log.ErrorF("error executing feature partials template: %v - %v", name, err)
					} else {
						output += buf.String()
					}
				}

			}
			return
		},
	}
	return
}