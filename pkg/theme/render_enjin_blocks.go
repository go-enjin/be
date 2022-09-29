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
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/strings"
)

func (re *renderEnjin) processBlock(blockData map[string]interface{}) (html template.HTML, err error) {
	if typeName, ok := blockData["type"]; ok {

		switch typeName {
		case "header":
			html, err = re.processHeaderBlock(blockData)

		case "notice":
			html, err = re.processNoticeBlock(blockData)

		case "link-list":
			html, err = re.processLinkListBlock(blockData)

		case "toc":
			html, err = re.processTableOfContentsBlock(blockData)

		case "image":
			html, err = re.processImageBlock(blockData)

		case "icon":
			html, err = re.processIconBlock(blockData)

		case "content":
			html, err = re.processContentBlock(blockData)

		default:
			log.WarnF("unsupported block type: %v", typeName)
			b, _ := json.MarshalIndent(blockData, "", "  ")
			html, err = re.processBlock(map[string]interface{}{
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

func (re *renderEnjin) prepareGenericBlockData(contentData interface{}) (blockDataContent map[string]interface{}, err error) {
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

func (re *renderEnjin) prepareGenericBlock(typeName string, data map[string]interface{}) (preparedData map[string]interface{}) {
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
	if v, ok = data["jump-top"].(string); ok && strings.IsTrue(v) {
		preparedData["JumpTop"] = "true"
	} else {
		preparedData["JumpTop"] = "false"
	}
	if v, ok = data["jump-link"].(string); ok && strings.IsTrue(v) {
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

func (re *renderEnjin) parseBlockHeader(content map[string]interface{}) (html template.HTML, ok bool) {
	var v []interface{}
	if v, ok = content["header"].([]interface{}); ok {
		if headings, err := re.renderInlineFields(v); err != nil {
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

func (re *renderEnjin) parseBlockFooter(content map[string]interface{}) (html template.HTML, ok bool) {
	var v []interface{}
	if v, ok = content["footer"].([]interface{}); ok {
		if footers, err := re.renderContainerFields(v); err != nil {
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