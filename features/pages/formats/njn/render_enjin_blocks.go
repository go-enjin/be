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

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

func (re *RenderEnjin) RenderErrorBlock(summary string, data ...interface{}) (html template.HTML, err error) {
	b, _ := json.MarshalIndent(data, "", "  ")
	html, err = re.ProcessBlock(map[string]interface{}{
		"type": "content",
		"content": map[string]interface{}{
			"section": []interface{}{
				map[string]interface{}{
					"type":    "details",
					"summary": summary,
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
	return
}

func (re *RenderEnjin) PrepareBlock(data map[string]interface{}) (block map[string]interface{}, err error) {
	if name, ok := re.ParseTypeName(data); ok {
		if njnBlock, ok := re.Njn.FindBlock(feature.AnyNjnClass, name); ok {
			if block, err = njnBlock.PrepareBlock(re, name, data); err == nil {
				log.TraceF("prepared block type: %v", name)
			}
			return
		}
		err = fmt.Errorf("unsupported block type: %v", name)
	} else {
		err = fmt.Errorf("missing block type")
	}
	return
}

func (re *RenderEnjin) RenderPreparedBlock(block map[string]interface{}) (html template.HTML, err error) {
	if name, ok := re.ParseTypeName(block); ok {
		if njnBlock, ok := re.Njn.FindBlock(feature.AnyNjnClass, name); ok {
			if html, err = njnBlock.RenderPreparedBlock(re, block); err == nil {
				log.TraceF("rendered prepared block type: %v (depth=%v)", name, re.GetCurrentDepth())
			}
			return
		}
		err = fmt.Errorf("unsupported block type: %v", name)
	} else {
		err = fmt.Errorf("missing block type")
	}
	return
}

func (re *RenderEnjin) ProcessBlock(block map[string]interface{}) (html template.HTML, err error) {
	if name, ok := re.ParseTypeName(block); ok {
		if njnBlock, ok := re.Njn.FindBlock(feature.AnyNjnClass, name); ok {
			if html, err = njnBlock.ProcessBlock(re, name, block); err == nil {
				log.DebugF("processed block type: %v (depth=%v)", name, re.GetCurrentDepth())
			}
			return
		}
		log.ErrorDF(1, "unsupported block type: %v", name)
		html, err = re.RenderErrorBlock(fmt.Sprintf("unsupported block type: %v", name), block)
	} else {
		html, err = re.RenderErrorBlock("missing block type", block)
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
	preparedData["Depth"] = re.GetCurrentDepth()
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
	if v, ok = data["class"].(string); ok {
		preparedData["Class"] = v
	}
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

	if linkHref, ok := data["link-href"].(string); ok {
		preparedData["LinkHref"] = linkHref
		if linkText, ok := data["link-text"].(string); ok {
			preparedData["LinkText"] = linkText
		} else {
			preparedData["LinkText"] = linkHref
		}
		if linkTarget, ok := data["link-target"].(string); ok {
			switch linkTarget {
			case "_self", "_blank", "_parent", "_top":
				preparedData["LinkTarget"] = linkTarget
			default:
				log.ErrorF(`invalid block link target: - "%v"`, linkTarget)
			}
		}
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
			log.ErrorDF(1, "error rendering inline fields: %v", err)
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
			log.ErrorDF(1, "error rendering container fields: %v", err)
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