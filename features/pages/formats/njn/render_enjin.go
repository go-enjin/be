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

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	path2 "github.com/go-enjin/be/pkg/path"
	"github.com/go-enjin/be/pkg/theme/types"
)

var _ feature.EnjinRenderer = (*RenderEnjin)(nil)

type RenderEnjin struct {
	Njn   feature.EnjinSystem
	Theme types.Theme
	ctx   context.Context

	blockCount   int
	headingLevel int
	headingCount int
	currentDepth int

	cache map[string]string
	data  interface{}

	footnotes map[int][]map[string]interface{}

	sync.RWMutex
}

func renderNjnData(f feature.EnjinSystem, ctx context.Context, t types.Theme, data interface{}) (html template.HTML, err *types.EnjinError) {
	re := new(RenderEnjin)
	re.Njn = f
	re.Theme = t
	re.ctx = ctx
	re.headingLevel = 0
	re.cache = make(map[string]string)
	re.data = data
	re.footnotes = make(map[int][]map[string]interface{}, 0)
	re.currentDepth = 0
	html, err = re.Render(data)
	return
}

func (re *RenderEnjin) Render(data interface{}) (html template.HTML, err *types.EnjinError) {

	switch v := data.(type) {

	case []interface{}:
		for _, c := range v {
			if h, e := re.Render(c); e != nil {
				err = e
				return
			} else {
				html += h
			}
		}

	case map[string]interface{}:
		if processed, e := re.ProcessBlock(v); e != nil {
			content, _ := json.MarshalIndent(v, "", "    ")
			blockType, _ := v["type"]
			err = types.NewEnjinError(fmt.Sprintf("error processing njn %v block", blockType), e.Error(), string(content))
		} else {
			html += processed
		}

	default:
		err = types.NewEnjinError(
			"unsupported njn data type",
			fmt.Sprintf("unsupported njn data type received: %T", v),
			fmt.Sprintf("%+v", v),
		)
	}

	return
}

func (re *RenderEnjin) GetNjnTemplateContent(name string) (contents string, err error) {
	// TODO: use the already prepared templating?
	if v, ok := re.cache[name]; ok {
		log.TraceF("found cached njn template: %v", name)
		contents = v
		return
	}
	path := path2.JoinWithSlashes("layouts", "partials", "njn", name)
	log.TraceF("looking for njn template: %v - %v", name, path)
	var data []byte
	if data, err = re.Theme.FS().ReadFile(path); err == nil {
		contents = string(data)
		re.cache[name] = contents
	} else {
		err = fmt.Errorf("njn template not found: %v", name)
	}
	return
}

func (re *RenderEnjin) RenderNjnTemplate(tag string, data map[string]interface{}) (html template.HTML, err error) {
	var tmplContent string
	if tmplContent, err = re.GetNjnTemplateContent(tag + ".tmpl"); err != nil {
		return
	} else {
		var tt *template.Template
		if tt, err = re.Theme.NewHtmlTemplate(tag).Parse(tmplContent); err == nil {
			var w bytes.Buffer
			data["Depth"] = re.GetCurrentDepth()
			if err = tt.Execute(&w, data); err == nil {
				html = template.HTML(w.Bytes())
			} else {
				err = fmt.Errorf("error rendering template: %v", err)
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

func (re *RenderEnjin) GetFootnotes(blockIndex int) (footnotes []map[string]interface{}) {
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