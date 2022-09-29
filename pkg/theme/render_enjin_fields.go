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

	"github.com/iancoleman/strcase"

	"github.com/go-enjin/be/pkg/log"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

func (re *renderEnjin) renderSectionFields(fields []interface{}) (combined []template.HTML, err error) {
	combined, err = re.renderContainerFields(fields)
	return
}

func (re *renderEnjin) renderFooterFields(fields []interface{}) (combined []template.HTML, err error) {
	combined, err = re.renderContainerFields(fields)
	return
}

func (re *renderEnjin) finalizeFieldData(data map[string]interface{}, field map[string]interface{}, skip ...string) {
	for key, value := range field {
		switch {
		case beStrings.StringInStrings(key, skip...):
		default:
			if key == "attributes" {
				if attrs, _, _, ok := re.parseFieldAttributes(field); ok {
					data["Attributes"] = re.finalizeFieldAttributes(attrs)
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

func (re *renderEnjin) parseFieldAttributes(field map[string]interface{}) (attributes map[string]interface{}, classes []string, styles map[string]string, ok bool) {
	attributes = make(map[string]interface{})
	classes = make([]string, 0)
	styles = make(map[string]string)

	if attrs, check := field["attributes"].(map[string]interface{}); check {

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
				log.ErrorF("error parsing class attribute: %T %+v", v, v)
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
				log.ErrorF("error parsing style attribute: %T %+v", v, v)
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

		ok = true
	} else if v, found := field["attributes"]; found {
		log.ErrorF("invalid attributes type found: %T %+v", v, v)
	}

	return
}

func (re *renderEnjin) finalizeFieldAttributes(attrs map[string]interface{}) (attributes []template.HTMLAttr) {
	for k, v := range attrs {
		switch t := v.(type) {
		case string:
			attributes = append(attributes, template.HTMLAttr(fmt.Sprintf(`%v="%v"`, k, beStrings.EscapeHtmlAttribute(t))))
		case template.HTMLAttr:
			attributes = append(attributes, template.HTMLAttr(fmt.Sprintf(`%v="%v"`, k, t)))
		default:
			log.ErrorF("unknown attributes value type: %T %+v", t, t)
		}
	}
	return
}