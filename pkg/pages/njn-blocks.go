// Copyright (c) 2024  The Go-Enjin Authors
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

package pages

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/iancoleman/strcase"

	"github.com/go-corelibs/htmlcss"
	"github.com/go-corelibs/slices"
	clStrings "github.com/go-corelibs/strings"
	"github.com/go-enjin/be/pkg/log"
)

func AddClassNamesToNjnBlock(data map[string]interface{}, classes ...string) map[string]interface{} {
	if v, ok := data["class"]; ok {
		switch t := v.(type) {
		case string:
			data["class"] = htmlcss.AddClassNames(t, classes...)
		case []interface{}:
			var list []interface{}
			for _, iface := range t {
				if vs, ok := iface.(string); ok {
					list = append(list, htmlcss.AddClassNames(vs, classes...))
				}
			}
			data["class"] = list
		}
	}
	return data
}

func ParseNjnFieldAttributes(field map[string]interface{}) (attributes map[string]interface{}, classes []string, styles map[string]string, err error) {
	attributes = make(map[string]interface{})
	classes = make([]string, 0)
	styles = make(map[string]string)

	if attrsValue, check := field["attributes"]; check {
		switch attrs := attrsValue.(type) {

		case map[string]interface{}:
			if v, found := attrs["class"]; found {
				switch t := v.(type) {
				case string:
					classes = strings.Split(t, " ")
					attributes["class"] = t
				case []interface{}:
					for _, i := range t {
						if name, ok := i.(string); ok {
							classes = append(classes, name)
						}
					}
					attributes["class"] = strings.Join(classes, " ")
				default:
					err = fmt.Errorf("unsupported class type: %T %+v", v, v)
					return
				}
			}

			if v, found := attrs["style"]; found {
				switch t := v.(type) {
				case string:
					attributes["style"] = t
				case map[string]interface{}:
					var list []string
					for k, vi := range t {
						if value, ok := vi.(string); ok {
							styles[k] = value
							list = append(list, fmt.Sprintf(`%v:%v`, k, value))
						}
					}
					attributes["style"] = strings.Join(list, ";")
				default:
					err = fmt.Errorf("unsupported style type: %T %+v", v, v)
					return
				}
			}

			for k, v := range attrs {
				switch k {
				case "class", "style":
				default:
					if vs, ok := v.(string); ok {
						attributes[k] = vs
					} else {
						attributes[k] = fmt.Sprintf("%v", v)
					}
				}
			}

		case []template.HTMLAttr:
			if a, e := htmlcss.ParseHtmlTagAttributes(attrs); e != nil {
				err = fmt.Errorf("error parsing html tag attributes: %v", e)
				log.ErrorF("%v", err)
				return
			} else {
				for k, v := range a {
					attributes[k] = v
				}
			}

		default:
			err = fmt.Errorf("unsupported attributes type: (%T) %+v", attrs, attrs)
		}
	} // no attributes present
	return
}

func FinalizeNjnFieldData(data map[string]interface{}, field map[string]interface{}, skip ...string) (err error) {
	for key, value := range field {
		switch {
		case slices.Present(key, skip...):
		default:
			if key == "attributes" {
				if attrs, _, _, e := ParseNjnFieldAttributes(field); e != nil {
					err = e
					return
				} else if data["Attributes"], e = FinalizeNjnFieldAttributes(attrs); e != nil {
					err = e
					return
				}
				continue
			}
			name := strcase.ToCamel(key)
			switch vv := value.(type) {
			case string:
				data[name] = template.HTML(vv)
			case int, int8, int16, int32, int64, float32, float64, bool:
				data[name] = fmt.Sprintf("%v", vv)
			default:
				data[name] = vv
			}
		}
	}
	return
}

func FinalizeNjnFieldAttributes(attrs map[string]interface{}) (attributes []template.HTMLAttr, err error) {
	for k, v := range attrs {
		switch t := v.(type) {
		case nil:
			attributes = append(attributes, template.HTMLAttr(fmt.Sprintf(`%v`, k)))
		case string:
			attributes = append(attributes, template.HTMLAttr(fmt.Sprintf(`%v=%q`, k, clStrings.EscapeHtmlAttribute(t))))
		case template.HTMLAttr:
			attributes = append(attributes, template.HTMLAttr(fmt.Sprintf(`%v=%q`, k, t)))
		default:
			err = fmt.Errorf("unsupported type: %T %+v", t, t)
		}
	}
	return
}
