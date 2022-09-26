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
)

func (re *renderEnjin) renderContainerFields(fields []interface{}) (combined []template.HTML, err error) {
	for _, field := range fields {
		if fieldMap, ok := field.(map[string]interface{}); ok {
			if c, e := re.renderContainerField(fieldMap); e != nil {
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

func (re *renderEnjin) renderContainerField(field map[string]interface{}) (combined []template.HTML, err error) {
	if ft, ok := field["type"].(string); ok {
		var data map[string]interface{}
		switch ft {

		case "dl":
			data, err = re.prepareDescriptionListFieldData(field)

		case "div", "dt", "dd", "samp", "blockquote":
			data, err = re.prepareContainerFieldData(ft, field)

		case "p":
			data, err = re.prepareInlineTextFieldData(ft, field)

		case "table":
			data, err = re.prepareTableFieldData(field)

		case "pre":
			data, err = re.preparePreFieldData(field)

		case "code":
			data, err = re.prepareCodeFieldData(field)

		case "ol", "ul":
			data, err = re.prepareListFieldData(ft, field)

		case "hr":
			data, err = re.prepareLiteralFieldData("hr", field)

		case "fieldset":
			data, err = re.prepareFieldsetFieldData(field)

		case "details":
			data, err = re.prepareDetailsFieldData(field)

		default:
			if c, e := re.renderInlineField(field); e == nil {
				combined = c
				return
			} else {
				err = fmt.Errorf("container field type error: %v", e)
			}
			if err == nil {
				err = fmt.Errorf("unsupported container field type: %v", ft)
			}
			return
		}
		if err != nil {
			return
		}
		if html, e := re.renderNjnTemplate(ft, data); e != nil {
			err = e
		} else {
			combined = append(combined, html)
		}
	} else {
		err = fmt.Errorf("container field missing type: %+v", field)
	}
	return
}

func (re *renderEnjin) prepareContainerFieldData(tag string, field map[string]interface{}) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	data["Type"] = tag
	if data["Text"], err = re.renderContainerFieldText(field); err != nil {
		return
	}
	re.finalizeFieldData(data, field, "type", "text")
	return
}

func (re *renderEnjin) renderContainerFieldText(field map[string]interface{}) (text template.HTML, err error) {
	if ti, ok := field["text"].([]interface{}); ok {
		text, err = re.renderContainerFieldList(ti)
	}
	return
}

func (re *renderEnjin) renderContainerFieldList(list []interface{}) (text template.HTML, err error) {
	for idx, tii := range list {
		if tis, ok := tii.(string); ok {
			if idx > 0 {
				text += " "
			}
			text += template.HTML(tis)
		} else if tis, ok := tii.(map[string]interface{}); ok {
			var rendered []template.HTML
			if rendered, err = re.renderContainerField(tis); err != nil {
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