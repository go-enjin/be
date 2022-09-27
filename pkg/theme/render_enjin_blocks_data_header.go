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
	"math"
	"strconv"
	"strings"

	"github.com/go-enjin/be/pkg/log"
)

func (re *renderEnjin) parseHeadingLevel(count, current int, blockData map[string]interface{}) (level, headingReset, headingLevel int) {
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

func (re *renderEnjin) processHeaderBlock(blockData map[string]interface{}) (html template.HTML, err error) {
	// log.DebugF("header received: %v", blockData)

	var blockDataContent map[string]interface{}
	if blockDataContent, err = re.prepareGenericBlockData(blockData["content"]); err != nil {
		return
	}

	var hr /*, hl*/ int
	re.headingLevel, hr, _ /*hl*/ = re.parseHeadingLevel(re.headingCount, re.headingLevel, blockData)
	re.headingCount += 1 // total number of header blocks on the page

	preparedData := re.prepareGenericBlock("header", blockData)

	// tag, _ := blockData["tag"]
	// log.DebugF("tag=%v, count=%v, level=%v, hr=%v, hl=%v", tag, re.headingCount, re.headingLevel, hr, hl)
	if hr == -255 /*&& hl == -255*/ {
		re.headingLevel += 1 // header blocks cause further blocks to be level+1
	}

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
	html, err = re.renderNjnTemplate("block/header", preparedData)

	return
}