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
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"strconv"
	"strings"

	"github.com/go-enjin/be/pkg/log"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

func (re *RenderEnjin) ProcessBlock(blockData map[string]interface{}) (html template.HTML, err error) {
	if typeName, ok := blockData["type"].(string); ok {

		inlineBlocks := re.Njn.InlineBlocks()
		if inlineBlock, ok := inlineBlocks[typeName]; ok {
			if html, err = inlineBlock.ProcessBlock(re, typeName, blockData); err == nil {
				log.DebugF("processed block: %v", typeName)
			}
			return
		}

		log.WarnF("unsupported block type: %v", typeName)
		b, _ := json.MarshalIndent(blockData, "", "  ")
		html, err = re.ProcessBlock(map[string]interface{}{
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

	} else {
		err = fmt.Errorf("missing type property: %+v", blockData)
	}
	return
}

func (re *RenderEnjin) PrepareGenericBlockData(contentData interface{}) (blockDataContent map[string]interface{}, err error) {
	blockDataContent = make(map[string]interface{})
	if contentData == nil {
		err = fmt.Errorf("content not found")
	} else if v, ok := contentData.(map[string]interface{}); ok {
		blockDataContent = v
	} else {
		err = fmt.Errorf("unsupported header content: %T", contentData)
	}
	return
}

func (re *RenderEnjin) PrepareGenericBlock(typeName string, data map[string]interface{}) (preparedData map[string]interface{}) {
	re.blockCount += 1

	var ok bool
	preparedData = make(map[string]interface{})
	preparedData["Context"] = re.ctx
	preparedData["Type"] = typeName
	preparedData["BlockIndex"] = re.blockCount
	if preparedData["Tag"], ok = data["tag"]; !ok {
		preparedData["Tag"] = fmt.Sprintf("%v-%d", typeName, re.blockCount)
	}
	if preparedData["Profile"], ok = data["profile"]; !ok {
		preparedData["Profile"] = "outer--inner"
	}
	if preparedData["Padding"], ok = data["padding"]; !ok {
		preparedData["Padding"] = "both"
	}
	if preparedData["Margins"], ok = data["margins"]; !ok {
		preparedData["Margins"] = "both"
	}
	var v string
	if v, ok = data["jump-top"].(string); ok && beStrings.IsTrue(v) {
		preparedData["JumpTop"] = "true"
	} else {
		preparedData["JumpTop"] = "false"
	}
	if v, ok = data["jump-link"].(string); ok && beStrings.IsTrue(v) {
		preparedData["JumpLink"] = "true"
	} else {
		preparedData["JumpLink"] = "false"
	}

	if re.headingCount == 0 && typeName != "header" {
		// first block on page is not a header, need to ensure that only one
		// h1 tag exists on the page
		switch re.headingLevel {
		case 0, 1:
			// first heading is 0, becomes h1
			// second heading is 1, becomes h2
			re.headingLevel += 1
		}
	}

	preparedData["HeadingLevel"] = re.headingLevel
	preparedData["HeadingCount"] = re.headingCount
	return
}

func (re *RenderEnjin) ParseBlockHeadingLevel(count, current int, blockData map[string]interface{}) (level, headingReset, headingLevel int) {
	headingReset, headingLevel = -255, -255

	if v, ok := blockData["heading-reset"]; ok {

		switch t := v.(type) {
		case string:
			if i, err := strconv.Atoi(t); err == nil {
				headingReset = i
			} else {
				log.ErrorF("error parsing heading-reset integer: %v", err)
			}

		case int:
			headingReset = t

		case float64:
			ti := int(math.Round(t))
			headingReset = ti

		}

	}

	if hl, ok := blockData["heading-level"]; ok {

		switch t := hl.(type) {

		case string:
			switch strings.ToLower(t) {
			case "+", "inc", "increment":
				headingLevel = 1
			case "-", "dec", "decrement":
				headingLevel = -1
			default:
				if i, err := strconv.Atoi(t); err != nil {
					headingLevel = i
				}
			}

		case int:
			headingLevel = t

		case float64:
			headingLevel = int(math.Round(t))

		}

	}

	if headingReset != -255 {
		switch headingReset {
		case 0, 1:
			if count == 0 {
				level = 1
			} else {
				level = 2
			}
		default:
			if headingReset > 0 {
				// positive numbers set literal
				level = headingReset
			} else {
				// add neg is subtraction
				level += headingReset
			}
		}
	} else if headingLevel != -255 {
		level += headingLevel
	}

	if level <= 1 {
		if count == 0 {
			level = 1
		} else {
			level = 2
		}
	}

	return
}

func (re *RenderEnjin) ParseBlockHeader(content map[string]interface{}) (html template.HTML, ok bool) {
	var v []interface{}
	if v, ok = content["header"].([]interface{}); ok {
		if headings, err := re.RenderInlineFields(v); err != nil {
			ok = false
			return
		} else {
			for _, heading := range headings {
				html += heading
			}
		}
	}
	return
}

func (re *RenderEnjin) ParseBlockFooter(content map[string]interface{}) (html template.HTML, ok bool) {
	var v []interface{}
	if v, ok = content["footer"].([]interface{}); ok {
		if footers, err := re.RenderContainerFields(v); err != nil {
			ok = false
			return
		} else {
			for _, footer := range footers {
				html += footer
			}
		}
	}
	return
}