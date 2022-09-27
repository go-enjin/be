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
	strings2 "strings"

	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/strings"
)

func (re *renderEnjin) walkTableOfContents(count, level int, data interface{}) (upCount, upLevel int, list []*tocItem) {
	list = make([]*tocItem, 0)

	parseNameTagTitle := func(dt map[string]interface{}) (name, tag, title string, valid bool) {
		var ok bool
		if name, ok = dt["type"].(string); ok {
			if tag, ok = dt["tag"].(string); ok {
				if c, ok := dt["content"].(map[string]interface{}); ok {
					if hi, ok := c["header"]; ok {
						switch ht := hi.(type) {
						case []interface{}:
							for _, hti := range ht {
								if hts, ok := hti.(string); ok {
									if title != "" {
										title += " "
									}
									title += hts
								}
							}
						case string:
							if title != "" {
								title += " "
							}
							title += ht
						}
					}
				}
				valid = true
			}
		}
		return
	}

	switch dt := data.(type) {
	case []interface{}:
		for _, dti := range dt {
			var children []*tocItem
			if count, level, children = re.walkTableOfContents(count, level, dti); children != nil {
				list = append(list, children...)
			}
		}

	case map[string]interface{}:
		// look for tag, title, update level and append any children
		if typeName, tag, title, ok := parseNameTagTitle(dt); ok {

			switch typeName {

			case "carousel", "pair":
				// nop

			case "sidebar":
				if b, ok := dt["blocks"].([]interface{}); ok {
					var children []*tocItem
					if count, level, children = re.walkTableOfContents(count, level, b); children != nil {
						log.TraceF("sidebar found: %v, %v, %v, %v", count, level, tag, title)
						list = append(list, children...)
					}
				}

			case "header":
				log.TraceF("header found: %v, %v, %v, %v", count, level, tag, title)
				var hr, hl int
				level, hr, hl = re.parseHeadingLevel(count, level, dt)
				if level > 1 {
					list = append(list, &tocItem{
						Tag:   tag,
						Title: title,
						Level: level,
					})
				}
				if hr == -255 && hl == -255 {
					level += 1
				}
				count += 1

			default:
				if title != "" {
					list = append(list, &tocItem{
						Tag:   tag,
						Title: title,
						Level: level,
					})
				}

			}
		} else {
			log.TraceF("parse name tag title failed: %+v", dt)
		}

	}
	upCount, upLevel = count, level
	return
}

func (re *renderEnjin) sortTableOfContents(toc []*tocItem) (items []*tocItem) {
	items = make([]*tocItem, 0)
	level := 1
	var last *tocItem
	for _, ti := range toc {
		if ti.Level <= level {
			items = append(items, ti)
		} else /*if ti.level > level*/ {
			if last != nil {
				last.Children = append(last.Children, ti)
			}
		}
		level = ti.Level
		last = ti
	}
	return
}

func (re *renderEnjin) processTableOfContentsBlock(blockData map[string]interface{}) (html template.HTML, err error) {
	// log.DebugF("content received: %v", blockData)

	var blockDataContent map[string]interface{}
	if blockDataContent, err = re.prepareGenericBlockData(blockData["content"]); err != nil {
		if err.Error() != "content not found" {
			return
		}
		err = nil
		blockDataContent = make(map[string]interface{})
	}

	preparedData := re.prepareGenericBlock("toc", blockData)

	var blockTag string
	if v, ok := preparedData["Tag"]; ok {
		blockTag, _ = v.(string)
	}

	pageTitle := false
	if v, ok := blockData["page-title"]; ok {
		switch t := v.(type) {
		case string:
			pageTitle = strings.IsTrue(t)
		case bool:
			pageTitle = t
		case int:
			pageTitle = t > 0
		case float64:
			pageTitle = t > 0
		}
		if pageTitle {
			preparedData["PageTitle"] = "true"
		} else {
			preparedData["PageTitle"] = "false"
		}
	} else {
		preparedData["PageTitle"] = "false"
	}

	var withSelf bool
	if v, ok := blockData["with-self"]; ok {
		switch t := v.(type) {
		case string:
			withSelf = strings.IsTrue(t)
		case bool:
			withSelf = t
		case int:
			withSelf = t > 0
		case float64:
			withSelf = t > 0
		}
		if withSelf {
			preparedData["WithSelf"] = "true"
		} else {
			preparedData["WithSelf"] = "false"
		}
	} else {
		preparedData["WithSelf"] = "false"
	}

	_, _, toc := re.walkTableOfContents(0, 0, re.data)
	items := re.sortTableOfContents(toc)

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

		if withSelf {
			items = append([]*tocItem{
				{
					Tag:   blockTag,
					Title: heading,
				},
			}, items...)
		}
		log.DebugF("have header - %+v", items[0])
	} else {
		log.DebugF("missing header - %+v", blockDataContent)
	}

	if footers, ok := blockDataContent["footer"].([]interface{}); ok {
		if preparedData["Footer"], err = re.renderFooterFields(footers); err != nil {
			return
		}
	}

	preparedData["Items"] = items

	if v, ok := blockData["counter"].(string); ok {
		v = strings2.ToLower(v)
		switch v {
		case "nested", "single":
			preparedData["Counter"] = v
		default:
			err = fmt.Errorf("invalid toc counter value: %v", v)
			return
		}
	} else {
		preparedData["Counter"] = "single"
	}

	// log.DebugF("prepared content: %v", preparedData)
	html, err = re.renderNjnTemplate("block/toc", preparedData)

	return
}

type tocItem struct {
	Tag      string
	Title    string
	Level    int
	Children []*tocItem
}

func (i *tocItem) String() string {
	return fmt.Sprintf(`<a href="#%v">%v</a>`, i.Tag, i.Title)
}