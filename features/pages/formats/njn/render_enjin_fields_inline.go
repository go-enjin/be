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
	"html/template"
	"strings"
)

func (re *RenderEnjin) RenderInlineFields(fields []interface{}) (combined []template.HTML, err error) {
	for _, field := range fields {
		if fieldMap, ok := field.(map[string]interface{}); ok {
			if c, e := re.RenderInlineField(fieldMap); e != nil {
				err = e
				return
			} else {
				combined = append(combined, c...)
			}
		} else if fieldString, ok := field.(string); ok {
			combined = append(combined, template.HTML(fieldString))
		} else {
			err = fmt.Errorf("unsupported inline field structure: %T", field)
			return
		}
	}
	return
}

func (re *RenderEnjin) RenderInlineField(field map[string]interface{}) (combined []template.HTML, err error) {
	if ft, ok := field["type"].(string); ok {
		ft = strings.ToLower(ft)
		var data map[string]interface{}

		inlineFields := re.Njn.InlineFields()
		if inlineField, ok := inlineFields[ft]; ok {
			// log.DebugF("preparing inline field %v: %+v", ft, inlineField.Tag())
			if data, err = inlineField.PrepareNjnData(re, ft, field); err != nil {
				return
			}
		} else {
			err = fmt.Errorf("unsupported field type: %v", ft)
			return
		}

		// log.DebugF("rendering inline field %v: %+v", ft, data)
		if html, e := re.RenderNjnTemplate("field/"+ft, data); e != nil {
			err = e
			return
		} else {
			combined = append(combined, html)
		}
	} else {
		err = fmt.Errorf("inline field missing type: %+v", field)
	}
	return
}

func (re *RenderEnjin) RenderInlineFieldText(field map[string]interface{}) (text template.HTML, err error) {
	if ti, ok := field["text"]; ok {
		switch t := ti.(type) {
		case []interface{}:
			text, err = re.RenderInlineFieldList(t)
		case interface{}:
			text, err = re.RenderInlineFieldList([]interface{}{t})
		default:
			err = fmt.Errorf("unsupported field text type: %T %+v", t, t)
		}
	} else {
		err = fmt.Errorf("missing field text: %v", field)
	}
	return
}

func (re *RenderEnjin) RenderInlineFieldList(list []interface{}) (html template.HTML, err error) {
	for idx, tii := range list {
		if tis, ok := tii.(string); ok {
			if idx > 0 {
				if _, ok := list[idx-1].(string); ok {
					html += " "
				}
			}
			html += template.HTML(tis)
		} else if tis, ok := tii.(map[string]interface{}); ok {
			var rendered []template.HTML
			if rendered, err = re.RenderInlineField(tis); err != nil {
				return
			}
			for _, rend := range rendered {
				html += rend
			}
		} else if tis, ok := tii.([]interface{}); ok {
			html, err = re.RenderInlineFieldList(tis)
		} else {
			html = ""
			err = fmt.Errorf("unsupported text value type: %T %+v", tii, tii)
			return
		}
	}
	return
}