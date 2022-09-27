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

	"github.com/go-enjin/be/pkg/log"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

func (re *renderEnjin) processNoticeBlock(blockData map[string]interface{}) (html template.HTML, err error) {
	// log.DebugF("content received: %v", blockData)

	var blockDataContent map[string]interface{}
	if blockDataContent, err = re.prepareGenericBlockData(blockData["content"]); err != nil {
		return
	}
	preparedData := re.prepareGenericBlock("notice", blockData)

	if v, ok := blockData["dismiss"].(string); ok && beStrings.IsTrue(v) {
		preparedData["Dismiss"] = "true"
	} else {
		preparedData["Dismiss"] = "false"
	}

	if v, ok := blockData["open"].(string); ok && beStrings.IsTrue(v) {
		preparedData["Open"] = "true"
	} else {
		preparedData["Open"] = "false"
	}

	if v, ok := blockData["notice-type"].(string); ok {
		v = strings.ToLower(v)
		switch v {
		case "info", "warn", "error":
			preparedData["NoticeType"] = v
		default:
			err = fmt.Errorf("invalid notice type: %v", v)
			return
		}
	} else {
		preparedData["NoticeType"] = "info"
	}

	var sectionHtml template.HTML
	if sections, ok := blockDataContent["section"].([]interface{}); ok {
		var renderedSectionFields []template.HTML
		if renderedSectionFields, err = re.renderSectionFields(sections); err != nil {
			err = fmt.Errorf("error rendering notice section field: %v", err)
			return
		} else {
			preparedData["Section"] = renderedSectionFields
			for _, s := range renderedSectionFields {
				sectionHtml += s
			}
		}
	}

	if v, ok := blockDataContent["summary"].([]interface{}); ok {
		var summary string
		for idx, vv := range v {
			if vs, ok := vv.(string); ok {
				if idx > 0 {
					summary += " "
				}
				summary += vs
			}
		}
		preparedData["Summary"] = summary
	} else if sectionHtml != "" {
		preparedData["Summary"] = sectionHtml
		delete(preparedData, "Section")
	} else {
		err = fmt.Errorf("notice block missing summary: %v - %v", blockData, sectionHtml)
		return
	}

	if footers, ok := blockDataContent["footer"].([]interface{}); ok {
		if preparedData["Footer"], err = re.renderFooterFields(footers); err != nil {
			return
		}
	}

	log.DebugF("notice block: %+v", preparedData)
	// log.DebugF("prepared content: %v", preparedData)
	html, err = re.renderNjnTemplate("block/notice", preparedData)

	return
}

func (re *renderEnjin) processLinkListBlock(blockData map[string]interface{}) (html template.HTML, err error) {
	// log.DebugF("content received: %v", blockData)

	var blockDataContent map[string]interface{}
	if blockDataContent, err = re.prepareGenericBlockData(blockData["content"]); err != nil {
		return
	}
	preparedData := re.prepareGenericBlock("link-list", blockData)

	if v, ok := blockDataContent["header"].([]interface{}); ok {
		var heading string
		for idx, vv := range v {
			if vs, ok := vv.(string); ok {
				if idx > 0 {
					heading += " "
				}
				heading += vs
			}

		}
		preparedData["Heading"] = heading
	}

	if sections, ok := blockDataContent["section"].([]interface{}); ok {
		sectionFields := make([]interface{}, 0)
		for _, si := range sections {
			switch st := si.(type) {
			case map[string]interface{}:
				if itype, ok := st["type"].(string); ok {
					if itype == "a" {
						st["decorated"] = "true"
						if attrs, ok := st["attributes"].(map[string]interface{}); ok {
							if ci, ok := attrs["class"]; ok {
								switch ct := ci.(type) {
								case string:
									attrs["class"] = ct + " decorated"
								case []interface{}:
									var classes []string
									for _, cli := range ct {
										if cls, ok := cli.(string); ok {
											classes = append(classes, cls)
										}
									}
									classes = append(classes, "decorated")
								}
								st["attributes"] = attrs
							} else {
								st["class"] = "decorated"
							}
						} else {
							st["class"] = "decorated"
						}
						sectionFields = append(sectionFields, st)
					} else {
						err = fmt.Errorf("link-list block has more than just anchor tags: %+v", st)
						return
					}
				}
			}
		}

		if preparedData["Section"], err = re.renderSectionFields([]interface{}{
			map[string]interface{}{
				"type": "ul",
				"list": sectionFields,
			},
		}); err != nil {
			return
		}
	}

	if footers, ok := blockDataContent["footer"].([]interface{}); ok {
		if preparedData["Footer"], err = re.renderFooterFields(footers); err != nil {
			return
		}
	}

	// log.DebugF("prepared content: %v", preparedData)
	html, err = re.renderNjnTemplate("block/link-list", preparedData)

	return
}

func (re *renderEnjin) processContentBlock(blockData map[string]interface{}) (html template.HTML, err error) {
	// log.DebugF("content received: %v", blockData)

	var blockDataContent map[string]interface{}
	if blockDataContent, err = re.prepareGenericBlockData(blockData["content"]); err != nil {
		return
	}
	preparedData := re.prepareGenericBlock("content", blockData)

	if v, ok := blockDataContent["header"].([]interface{}); ok {
		var heading string
		for idx, vv := range v {
			if vs, ok := vv.(string); ok {
				if idx > 0 {
					heading += " "
				}
				heading += vs
			}

		}
		preparedData["Heading"] = heading
	}

	if sections, ok := blockDataContent["section"].([]interface{}); ok {
		if preparedData["Section"], err = re.renderSectionFields(sections); err != nil {
			return
		}
	}

	if footers, ok := blockDataContent["footer"].([]interface{}); ok {
		if preparedData["Footer"], err = re.renderFooterFields(footers); err != nil {
			return
		}
	}

	// log.DebugF("prepared content: %v", preparedData)
	html, err = re.renderNjnTemplate("block/content", preparedData)

	return
}