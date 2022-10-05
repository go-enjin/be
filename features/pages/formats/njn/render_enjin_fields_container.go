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

	"github.com/go-enjin/be/pkg/feature"
)

func (re *RenderEnjin) RenderContainerFields(fields []interface{}) (combined []template.HTML, err error) {
	for _, fieldItem := range fields {
		if field, ok := fieldItem.(map[string]interface{}); ok {
			if c, e := re.RenderContainerField(field); e != nil {
				err = e
				return
			} else {
				combined = append(combined, c...)
			}
		} else {
			err = fmt.Errorf("unsupported field structure: %T", fieldItem)
			return
		}
	}
	return
}

func (re *RenderEnjin) RenderContainerField(field map[string]interface{}) (combined []template.HTML, err error) {
	if ft, ok := re.ParseTypeName(field); ok {
		var data map[string]interface{}

		if containerField, ok := re.Njn.FindField(feature.AnyNjnClass, ft); ok {
			if data, err = containerField.PrepareNjnData(re, ft, field); err != nil {
				return
			}
		} else {
			err = fmt.Errorf("njn field not found: %v", ft)
			return
		}

		if html, e := re.RenderNjnTemplate("field/"+ft, data); e != nil {
			err = e
		} else {
			combined = append(combined, html)
		}
	} else {
		err = fmt.Errorf("field missing type")
	}
	return
}

func (re *RenderEnjin) RenderContainerFieldText(field map[string]interface{}) (text template.HTML, err error) {
	if textItem, ok := field["text"]; ok {
		switch t := textItem.(type) {
		case []interface{}:
			text, err = re.RenderContainerFieldList(t)
		case interface{}:
			text, err = re.RenderContainerFieldList([]interface{}{t})
		default:
			err = fmt.Errorf("unsupported field text structure: %T", t)
		}
	} else {
		err = fmt.Errorf("field missing text")
	}
	return
}

func (re *RenderEnjin) RenderContainerFieldList(list []interface{}) (text template.HTML, err error) {
	for idx, listItem := range list {
		if listItemString, ok := listItem.(string); ok {
			if idx > 0 {
				text += " "
			}
			text += template.HTML(listItemString)
		} else if listItemMap, ok := listItem.(map[string]interface{}); ok {
			var rendered []template.HTML
			if rendered, err = re.RenderContainerField(listItemMap); err != nil {
				return
			}
			for _, rend := range rendered {
				text += rend
			}
		} else {
			text = ""
			err = fmt.Errorf("unsupported field structure: %T", listItem)
			return
		}
	}
	return
}