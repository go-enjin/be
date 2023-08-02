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
	"fmt"
	"strings"

	"golang.org/x/net/html"

	beStrings "github.com/go-enjin/be/pkg/strings"
)

func (re *RenderEnjin) PrepareStringTags(text string) (data []interface{}, err error) {
	if doc, e := html.Parse(strings.NewReader(text)); e != nil {
		err = e
		return
	} else {
		data = re.WalkStringTags(doc)
	}
	return
}

func (re *RenderEnjin) WalkStringTags(doc *html.Node) (prepared []interface{}) {

	var traverse func(depth string, n *html.Node) (*html.Node, []interface{})

	foundBody := false
	traverse = func(depth string, n *html.Node) (*html.Node, []interface{}) {
		var data []interface{}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if !foundBody {
				// log.DebugF("%v[skipping]: %+v - %+v", depth, c.Type, c.Data)
				foundBody = c.Type == html.ElementNode && c.Data == "body"
				if res, childData := traverse(depth+" ", c); res != nil {
					if len(childData) > 0 {
						data = append(data, childData...)
					}
					return res, data
				} else if len(childData) > 0 {
					data = append(data, childData...)
				}
			} else if c.Type == html.TextNode {
				// log.DebugF("%v[storing]: %v", depth, c.Data)

				parsed := re.Enjin.TranslateShortcodes(c.Data, re.ctx)
				if last := len(data) - 1; last >= 0 {
					if v, ok := data[last].(string); ok {
						data[last] = v + parsed
					} else {
						data = append(data, parsed)
					}
				} else {
					data = append(data, parsed)
				}

			} else if c.Type == html.ElementNode {
				if beStrings.StringInStrings(c.Data, re.Njn.StringTags()...) {
					// log.DebugF("%v[shortcode]: %v", depth, c.Data)
					child := make(map[string]interface{})
					child["Type"] = c.Data // tag name for element nodes
					res, childData := traverse(depth+" ", c)
					child["Text"] = childData
					data = append(data, child)
					if res != nil {
						return res, data
					}
				} else {
					// log.DebugF("%v[ignored]: %v", depth, c.Data)
					res, childData := traverse(depth+" ", c)
					childText := "<" + c.Data + ">"
					for _, childDatum := range childData {
						switch typedDatum := childDatum.(type) {
						case string:
							childText += typedDatum
						default:
							// TODO: parse content within StringTags for fields and other oddities
							childText += fmt.Sprintf("(stringtags error: %T)", typedDatum)
						}
					}
					childText += "</" + c.Data + ">"

					if last := len(data) - 1; last >= 0 {
						if v, ok := data[last].(string); ok {
							data[last] = v + childText
						} else {
							data = append(data, childText)
						}
					} else {
						data = append(data, childText)
					}

					if res != nil {
						return res, data
					}
				}
			}
		}

		return nil, data
	}

	_, prepared = traverse("", doc)

	// log.DebugF("returning prepared: %+v", prepared)
	return
}