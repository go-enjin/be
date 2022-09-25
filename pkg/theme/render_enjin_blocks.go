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

	"github.com/go-enjin/be/pkg/context"
)

func (re *renderEnjin) processHeaderBlock(ctx context.Context, blockData map[string]interface{}) (html template.HTML, err error) {
	// log.DebugF("header received: %v", blockData)

	var blockDataContent map[string]interface{}
	if blockDataContent, err = re.prepareGenericBlockData(blockData["content"]); err != nil {
		return
	}
	preparedData := re.prepareGenericBlock("header", blockData)

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

	if list, ok := blockDataContent["nav"].([]interface{}); ok {
		var navItems []map[string]interface{}
		for _, item := range list {
			var navItem map[string]interface{}
			if v, ok := item.(map[string]interface{}); ok {
				if vType, ok := v["type"].(string); ok {

					switch vType {
					case "a":
						if navItem, err = re.prepareAnchorFieldData(v); err != nil {
							return
						}

					default:
						err = fmt.Errorf("unsupported heading nav item type: %+v", v)
					}

				} else {
					err = fmt.Errorf("heading nav item missing type: %+v", v)
					return
				}
			}
			navItems = append(navItems, navItem)
		}
		preparedData["Nav"] = navItems
	}

	// log.DebugF("prepared header: %v", preparedData)
	html, err = re.renderNjnTemplate("header", preparedData)

	return
}

func (re *renderEnjin) processContentBlock(ctx context.Context, blockData map[string]interface{}) (html template.HTML, err error) {
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
	html, err = re.renderNjnTemplate("content", preparedData)

	return
}