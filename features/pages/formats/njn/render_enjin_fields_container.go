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

func (re *RenderEnjin) RenderContainerFields(fields []interface{}) (combined []template.HTML, err error) {
	for _, field := range fields {
		if fieldMap, ok := field.(map[string]interface{}); ok {
			if c, e := re.RenderContainerField(fieldMap); e != nil {
				err = e
				return
			} else {
				combined = append(combined, c...)
			}
		} else {
			err = fmt.Errorf("unsupported container field structure: %T", field)
			return
		}
	}
	return
}

func (re *RenderEnjin) RenderContainerField(field map[string]interface{}) (combined []template.HTML, err error) {

	if ft, ok := field["type"].(string); ok {
		ft = strings.ToLower(ft)
		var data map[string]interface{}

		containerFields := re.Njn.ContainerFields()
		if containerField, ok := containerFields[ft]; ok {
			// log.DebugF("preparing container field %v: %+v", ft, containerField.Tag())
			if data, err = containerField.PrepareNjnData(re, ft, field); err != nil {
				return
			}
		} else {
			combined, err = re.RenderInlineField(field)
			return
		}

		// log.DebugF("rendering container field %v: %+v", ft, data)
		if html, e := re.RenderNjnTemplate(ft, data); e != nil {
			err = e
		} else {
			combined = append(combined, html)
		}
	} else {
		err = fmt.Errorf("container field missing type: %+v", field)
	}
	return
}

func (re *RenderEnjin) RenderContainerFieldText(field map[string]interface{}) (text template.HTML, err error) {
	if ti, ok := field["text"].([]interface{}); ok {
		text, err = re.RenderContainerFieldList(ti)
	}
	return
}

func (re *RenderEnjin) RenderContainerFieldList(list []interface{}) (text template.HTML, err error) {
	for idx, tii := range list {
		if tis, ok := tii.(string); ok {
			if idx > 0 {
				text += " "
			}
			text += template.HTML(tis)
		} else if tis, ok := tii.(map[string]interface{}); ok {
			var rendered []template.HTML
			if rendered, err = re.RenderContainerField(tis); err != nil {
				return
			}
			for _, rend := range rendered {
				text += rend
			}
		} else {
			text = ""
			err = fmt.Errorf("unsupported anchor text value type: %T %+v", tii, tii)
			return
		}
	}
	return
}