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

	"github.com/go-enjin/be/pkg/log"
)

func (re *renderEnjin) renderInlineFields(fields []interface{}) (combined []template.HTML, err error) {
	for _, field := range fields {
		if fieldMap, ok := field.(map[string]interface{}); ok {
			if c, e := re.renderInlineField(fieldMap); e != nil {
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

func (re *renderEnjin) renderInlineField(field map[string]interface{}) (combined []template.HTML, err error) {
	if ft, ok := field["type"].(string); ok {
		var data map[string]interface{}
		switch ft {
		case "a":
			data, err = re.prepareAnchorFieldData(field)

		case "abbr", "b", "cite", "del", "dfn", "em", "i", "ins", "kbd", "mark", "meter", "progress",
			"q", "s", "small", "strong", "sub", "sup", "u", "var":
			data, err = re.prepareInlineTextFieldData(ft, field)

		case "span":
			data, err = re.prepareInlineTextFieldData(ft, field)

		case "fa-icon":
			data, err = re.prepareFontAwesomeIconFieldData(field)

		case "figure":
			data, err = re.prepareFigureFieldData(field)

		case "br", "hr":
			data, err = re.prepareLiteralFieldData(ft, field)

		case "img":
			data, err = re.prepareImageFieldData(field)

		case "picture":
			data, err = re.preparePictureFieldData(field)

		case "button":
			data, err = re.prepareInlineTextFieldData(ft, field)

		case "input":
			data, err = re.prepareInputFieldData(field)

		case "select":
			data, err = re.prepareSelectFieldData(field)

		case "optgroup":
			data, err = re.prepareOptionGroupFieldData(field)

		case "option":
			data, err = re.prepareOptionFieldData(field)

		default:
			err = fmt.Errorf("unsupported inline field type: %v - %+v", ft, field)
			return
		}
		if err != nil {
			return
		}
		if html, e := re.renderNjnTemplate(ft, data); e != nil {
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

func (re *renderEnjin) prepareInlineFieldData(tag string, field map[string]interface{}) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	data["Type"] = tag
	re.finalizeFieldData(data, field, "type")
	return
}

func (re *renderEnjin) prepareInlineTextFieldData(tag string, field map[string]interface{}) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	data["Type"] = tag
	if attrs, ok := field["attributes"]; ok {
		data["Attributes"] = attrs
		log.DebugF("attrs: %+v", attrs)
	}
	if data["Text"], err = re.renderInlineFieldText(field); err != nil {
		return
	}
	re.finalizeFieldData(data, field, "type", "text", "attributes")
	return
}

func (re *renderEnjin) renderInlineFieldText(field map[string]interface{}) (text template.HTML, err error) {
	if ti, ok := field["text"]; ok {
		switch t := ti.(type) {
		case []interface{}:
			text, err = re.renderInlineFieldList(t)
		case interface{}:
			text, err = re.renderInlineFieldList([]interface{}{t})
		default:
			err = fmt.Errorf("unsupported field text type: %T %+v", t, t)
		}
	} else {
		err = fmt.Errorf("missing field text: %v", field)
	}
	return
}

func (re *renderEnjin) renderInlineFieldList(list []interface{}) (text template.HTML, err error) {
	for idx, tii := range list {
		if tis, ok := tii.(string); ok {
			if idx > 0 {
				text += " "
			}
			text += template.HTML(tis)
		} else if tis, ok := tii.(map[string]interface{}); ok {
			var rendered []template.HTML
			if rendered, err = re.renderInlineField(tis); err != nil {
				return
			}
			for _, rend := range rendered {
				text += rend
			}
		} else if tis, ok := tii.([]interface{}); ok {
			text, err = re.renderInlineFieldList(tis)
		} else {
			text = ""
			err = fmt.Errorf("unsupported text value type: %T %+v", tii, tii)
			return
		}
	}
	return
}