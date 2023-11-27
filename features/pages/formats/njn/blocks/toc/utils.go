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

package toc

import (
	"html/template"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
)

func parseNameTagTitle(dt map[string]interface{}) (name, tag, title string, valid bool) {
	var ok bool
	if name, ok = dt["type"].(string); ok {
		if tag, ok = dt["tag"].(string); ok {
			if c, ok := dt["content"].(map[string]interface{}); ok {
				if hi, ok := c["header"]; ok {
					// TODO: find a better way of determining a block's "title" heading
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

func walkTableOfContents(re feature.EnjinRenderer, count, level int, data interface{}) (upCount, upLevel int, list []*tocItem) {
	list = make([]*tocItem, 0)

	switch dt := data.(type) {
	case []interface{}:
		for _, dti := range dt {
			var children []*tocItem
			if count, level, children = walkTableOfContents(re, count, level, dti); children != nil {
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
				if title != "" {
					list = append(list, &tocItem{
						Tag:   tag,
						Title: template.HTML(title),
						Level: level,
					})
					log.TraceF("sidebar found: %v, %v, %v, %v", count, level, tag, title)
				}
				if c, ok := dt["content"].(map[string]interface{}); ok {
					if b, ok := c["blocks"].([]interface{}); ok {
						var children []*tocItem
						if count, level, children = walkTableOfContents(re, count, level, b); len(children) > 0 {
							log.TraceF("sidebar children: %v, %v, %v, %v", count, level, tag, children)
							list = append(list, children...)
						}
					}
				}

			case "header":
				var hr, hl int
				level, hr, hl = re.ParseBlockHeadingLevel(count, level, dt)
				log.TraceF("header found: count=%v, level=%v, tag=%v, title=%v (hr=%v,hl=%v)", count, level, tag, title, hr, hl)
				if level > 1 { // skip h1
					list = append(list, &tocItem{
						Tag:   tag,
						Title: template.HTML(title),
						Level: level,
					})
				}
				if hr == -255 /*&& hl == -255*/ {
					level += 1
				}
				count += 1

			default:
				if title != "" {
					list = append(list, &tocItem{
						Tag:   tag,
						Title: template.HTML(title),
						Level: level,
					})
					log.TraceF("default found: %v, %v, %v, %v", count, level, tag, title)
				}

			}
		} else {
			log.TraceF("parse name tag title failed: %+v", dt)
		}

	}
	upCount, upLevel = count, level
	return
}

func sortTableOfContents(toc []*tocItem) (items []*tocItem) {
	level := 1
	var parent *tocItem
	for _, item := range toc {
		if item.Level <= level {
			level = item.Level
			items = append(items, item)
			parent = item
		} else {
			if parent != nil {
				parent.Children = append(parent.Children, item)
			} else {
				level = item.Level
				items = append(items, item)
				parent = item
			}
		}
	}
	return
}
