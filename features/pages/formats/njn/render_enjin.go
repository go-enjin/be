//go:build !exclude_pages_formats && !exclude_pages_format_njn

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

package njn

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"sync"

	"github.com/iancoleman/strcase"

	clPath "github.com/go-corelibs/path"
	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/request/argv"
)

var _ feature.EnjinRenderer = (*RenderEnjin)(nil)

type RenderEnjin struct {
	Njn   feature.EnjinSystem
	Theme feature.Theme
	Enjin feature.Internals
	ctx   context.Context

	blockCount   int
	headingLevel int
	headingCount int
	currentDepth int
	withinAside  bool

	cache map[string]string
	data  interface{}

	footnotes map[int][]map[string]interface{}

	sync.RWMutex
}

func renderNjnData(f *CFeature, ctx context.Context, data interface{}) (html template.HTML, redirect string, err error) {
	re := new(RenderEnjin)
	re.Njn = f
	re.Enjin = f.Enjin
	re.Theme = f.Enjin.MustGetTheme()
	re.ctx = ctx
	re.headingLevel = 0
	re.cache = make(map[string]string)
	re.data = data
	re.footnotes = make(map[int][]map[string]interface{}, 0)
	re.currentDepth = 0
	html, redirect, err = re.Render(data)
	return
}

func (re *RenderEnjin) RequestArgv() (reqArgv *argv.Argv) {
	var ok bool
	if reqArgv, ok = re.ctx.Get(argv.RequestKey.String()).(*argv.Argv); ok {
		//reqArgv.Request, _ = re.ctx.Get("R").(*http.Request)
	} else {
		reqArgv = &argv.Argv{}
	}
	return
}

func (re *RenderEnjin) RequestContext() (ctx context.Context) {
	ctx = re.ctx
	return
}

func (re *RenderEnjin) Render(data interface{}) (html template.HTML, redirect string, err error) {

	if prepared, redir, e := re.PreparePageData(data); e != nil {
		err = e
		return
	} else if redir != "" {
		redirect = redir
		return
	} else {
		if h, ee := re.RenderNjnTemplateList("block-list", prepared); ee != nil {
			content, _ := json.MarshalIndent(prepared, "", "    ")
			err = errors.NewEnjinError(
				"error rendering njn template list",
				ee.Error(),
				string(content),
			)
		} else {
			html = h
		}
	}

	return
}

func (re *RenderEnjin) PreparePageData(data interface{}) (blocks []interface{}, redirect string, err *errors.EnjinError) {

	switch typedData := data.(type) {

	case []interface{}:
		for _, c := range typedData {
			if preparedBlocks, redir, e := re.PreparePageData(c); e != nil {
				err = e
				return
			} else if redir != "" {
				redirect = redir
				return
			} else {
				blocks = append(blocks, preparedBlocks...)
			}
		}

	case map[string]interface{}:
		if prepared, redir, e := re.PrepareBlock(typedData); e != nil {
			preparedBlock, _ := re.PrepareErrorBlock(e.Error(), typedData)
			blocks = append(blocks, preparedBlock)
		} else if redir != "" {
			redirect = redir
			return
		} else {
			blocks = append(blocks, prepared)
		}

	default:
		err = errors.NewEnjinError(
			"unsupported njn data type",
			fmt.Sprintf("unsupported njn data type received: %T", typedData),
			fmt.Sprintf("%+v", typedData),
		)
	}

	return
}

func (re *RenderEnjin) GetNjnTemplateContent(name string) (contents string, err error) {
	if v, ok := re.cache[name]; ok {
		log.TraceF("found cached njn template: %v", name)
		contents = v
		return
	}
	path := clPath.JoinWithSlashes("layouts", globals.PartialThemeLayoutName, "njn", name)
	log.TraceF("looking for njn template: %v - %v", name, path)
	var data []byte
	if data, err = re.Theme.ThemeFS().ReadFile(path); err == nil {
		contents = string(data)
		re.cache[name] = contents
		log.TraceF("caching new njn template: %v - %v", name, path)
	} else if parent := re.Theme.GetParent(); parent != nil && parent.ThemeFS() != nil {
		if data, err = parent.ThemeFS().ReadFile(path); err == nil {
			contents = string(data)
			re.cache[name] = contents
			log.TraceF("caching new parent njn template: %v - %v", name, path)
		} else {
			err = fmt.Errorf("parent njn template not found: %v, expected path: %v", name, path)
		}
	} else {
		err = fmt.Errorf("njn template not found: %v, expected path: %v", name, path)
	}
	return
}

func (re *RenderEnjin) RenderNjnTemplateList(tag string, data []interface{}) (html template.HTML, err error) {
	var tmplContent string
	if tmplContent, err = re.GetNjnTemplateContent(tag + ".tmpl"); err != nil {
		return
	} else {
		var tt *template.Template
		if tt, err = re.Theme.NewHtmlTemplate(re.Enjin, "render-enjin--"+tag, re.ctx); err == nil {
			if tt, err = tt.Parse(tmplContent); err == nil {
				var w bytes.Buffer
				if err = tt.Execute(&w, data); err == nil {
					html = template.HTML(w.Bytes())
				} else {
					err = fmt.Errorf("error rendering template: %v", err)
				}
			}
		} else {
			err = fmt.Errorf("error parsing template: %v", err)
		}
	}
	return
}

func (re *RenderEnjin) RenderNjnTemplate(tag string, data map[string]interface{}) (html template.HTML, err error) {
	var tmplContent string
	if tmplContent, err = re.GetNjnTemplateContent(tag + ".tmpl"); err != nil {
		return
	} else {
		var tt *template.Template
		if tt, err = re.Theme.NewHtmlTemplate(re.Enjin, "render-enjin--"+tag, re.ctx); err == nil {
			if tt, err = tt.Parse(tmplContent); err == nil {
				var w bytes.Buffer
				if err = tt.Execute(&w, data); err == nil {
					html = template.HTML(w.Bytes())
				} else {
					err = fmt.Errorf("error rendering template: %v", err)
				}
			}
		} else {
			err = fmt.Errorf("error parsing template: %v", err)
		}
	}
	return
}

func (re *RenderEnjin) GetData() (data interface{}) {
	return re.data
}

func (re *RenderEnjin) GetHeadingLevel() (level int) {
	level = re.headingLevel
	return
}

func (re *RenderEnjin) IncHeadingLevel() {
	re.headingLevel += 1
	return
}

func (re *RenderEnjin) DecHeadingLevel() {
	re.headingLevel -= 1
	return
}

func (re *RenderEnjin) SetHeadingLevel(level int) {
	re.headingLevel = level
	return
}

func (re *RenderEnjin) GetHeadingCount() (count int) {
	count = re.headingCount
	return
}

func (re *RenderEnjin) IncHeadingCount() {
	re.headingCount += 1
	return
}

func (re *RenderEnjin) SetHeadingCount(count int) {
	re.headingCount = count
	return
}

func (re *RenderEnjin) AddFootnote(blockIndex int, field map[string]interface{}) (index int) {
	if _, ok := re.footnotes[blockIndex]; !ok {
		re.footnotes[blockIndex] = make([]map[string]interface{}, 0)
	}
	re.footnotes[blockIndex] = append(re.footnotes[blockIndex], field)
	index = len(re.footnotes[blockIndex]) - 1
	return
}

func (re *RenderEnjin) PrepareFootnotes(blockIndex int) (footnotes []map[string]interface{}, err error) {
	footnotes, _ = re.footnotes[blockIndex]
	return
}

func (re *RenderEnjin) GetBlockIndex() (index int) {
	index = re.blockCount - 1
	return
}

func (re *RenderEnjin) ParseTypeName(data map[string]interface{}) (name string, ok bool) {
	if name, ok = data["type"].(string); ok {
		name = strcase.ToKebab(name)
	} else if name, ok = data["Type"].(string); ok {
		name = strcase.ToKebab(name)
	}
	return
}

func (re *RenderEnjin) ParseFieldAndTypeName(data interface{}) (field map[string]interface{}, name string, ok bool) {
	if field, ok = data.(map[string]interface{}); ok {
		name, ok = re.ParseTypeName(field)
	}
	return
}

func (re *RenderEnjin) GetWithinAside() (within bool) {
	within = re.withinAside
	return
}

func (re *RenderEnjin) SetWithinAside(within bool) {
	re.withinAside = within
	return
}

func (re *RenderEnjin) GetCurrentDepth() (depth int) {
	depth = re.currentDepth
	return
}

func (re *RenderEnjin) IncCurrentDepth() (depth int) {
	re.currentDepth += 1
	return re.currentDepth
}

func (re *RenderEnjin) DecCurrentDepth() (depth int) {
	re.currentDepth -= 1
	return re.currentDepth
}
