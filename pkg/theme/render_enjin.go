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
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
	"sync"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/log"
	bePath "github.com/go-enjin/be/pkg/path"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

type renderEnjin struct {
	theme *Theme
	ctx   context.Context

	headingLevel int
	headingCount int

	cache map[string]string

	sync.RWMutex
}

func newNjnRenderer(ctx context.Context, t *Theme) (re *renderEnjin) {
	re = new(renderEnjin)
	re.theme = t
	re.ctx = ctx
	re.headingLevel = 0
	re.cache = make(map[string]string)
	return
}

func (re *renderEnjin) getNjnTemplateContent(name string) (contents string, err error) {
	// TODO: use the already prepared templating?
	if v, ok := re.cache[name]; ok {
		log.TraceF("found cached njn template: %v", name)
		contents = v
		return
	}
	path := bePath.JoinWithSlashes("layouts", "partials", "njn", name)
	log.TraceF("looking for njn template: %v - %v", name, path)
	var data []byte
	if data, err = re.theme.FileSystem.ReadFile(path); err == nil {
		contents = string(data)
		re.cache[name] = contents
	} else {
		err = fmt.Errorf("njn template not found: %v", name)
	}
	return
}

func (re *renderEnjin) render(ctx context.Context, data interface{}) (html template.HTML, err error) {

	switch v := data.(type) {

	case []interface{}:
		for _, c := range v {
			if h, e := re.render(ctx, c); e != nil {
				err = e
				return
			} else {
				html += h
			}
		}

	case map[string]interface{}:
		html, err = re.processBlock(ctx, v)

	default:
		err = fmt.Errorf("unsupported njn data received: %T", v)
	}

	return
}

func (re *renderEnjin) processBlock(ctx context.Context, blockData map[string]interface{}) (html template.HTML, err error) {
	if typeName, ok := blockData["type"]; ok {

		switch typeName {
		case "header":
			html, err = re.processHeaderBlock(ctx, blockData)

		case "content":
			html, err = re.processContentBlock(ctx, blockData)

		default:
			log.WarnF("unsupported block type: %v", typeName)
			b, _ := json.MarshalIndent(blockData, "", "  ")
			html, err = re.processBlock(ctx, map[string]interface{}{
				"type": "content",
				"content": map[string]interface{}{
					"section": []interface{}{
						map[string]interface{}{
							"type":    "details",
							"summary": fmt.Sprintf("Unexpected block type: %v", typeName),
							"text": []interface{}{
								map[string]interface{}{
									"type": "pre",
									"text": string(b),
								},
							},
						},
					},
				},
			})
		}

	} else {
		err = fmt.Errorf("missing type property: %+v", blockData)
	}
	return
}

func (re *renderEnjin) prepareGenericBlock(typeName string, blockData map[string]interface{}) (fieldData map[string]interface{}) {
	fieldData = make(map[string]interface{})
	fieldData["Type"] = typeName
	fieldData["Tag"], _ = blockData["tag"]
	fieldData["Profile"], _ = blockData["profile"]
	fieldData["Padding"], _ = blockData["padding"]
	fieldData["Margins"], _ = blockData["margins"]
	fieldData["JumpTop"], _ = blockData["jump-top"]
	fieldData["JumpLink"], _ = blockData["jump-link"]
	switch re.headingLevel {
	case 0, 1:
		re.headingLevel += 1
	}
	fieldData["HeadingLevel"] = re.headingLevel
	fieldData["HeadingCount"] = re.headingCount
	return
}

func (re *renderEnjin) prepareGenericBlockData(contentData interface{}) (blockDataContent map[string]interface{}, err error) {
	blockDataContent = make(map[string]interface{})
	if contentData == nil {
		err = fmt.Errorf("header without any content: %+v", contentData)
	} else if v, ok := contentData.(map[string]interface{}); ok {
		blockDataContent = v
	} else {
		err = fmt.Errorf("unsupported header content: %T", contentData)
	}
	return
}

func (re *renderEnjin) renderNjnTemplate(tag string, data map[string]interface{}) (html template.HTML, err error) {
	var tmplContent string
	if tmplContent, err = re.getNjnTemplateContent(tag + ".tmpl"); err != nil {
		return
	} else {
		var tt *template.Template
		if tt, err = re.theme.NewHtmlTemplate(tag).Parse(tmplContent); err == nil {
			var w bytes.Buffer
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

func (re *renderEnjin) parseFieldAttributes(field map[string]interface{}) (attrs map[string]interface{}, classes []string, styles map[string]string, ok bool) {
	classes = make([]string, 0)
	styles = make(map[string]string)
	if attrs, ok = field["attributes"].(map[string]interface{}); ok {

		if v, found := attrs["class"]; found {
			switch t := v.(type) {
			case string:
				classes = strings.Split(t, " ")
				attrs["class"] = t
			case []interface{}:
				for _, i := range t {
					if name, ok := i.(string); ok {
						classes = append(classes, name)
					}
				}
				attrs["class"] = strings.Join(classes, " ")
			}
		}

		if v, found := attrs["style"]; found {
			switch t := v.(type) {
			case string:
				attrs["style"] = t
			case map[string]interface{}:
				var list []string
				for k, vi := range t {
					if value, ok := vi.(string); ok {
						styles[k] = value
						list = append(list, fmt.Sprintf(`%v="%v"`, k, beStrings.EscapeHtmlAttribute(value)))
					}
				}
				attrs["style"] = strings.Join(list, ";")
			}
		}
	}
	return
}