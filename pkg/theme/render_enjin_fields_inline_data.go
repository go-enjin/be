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
	"strconv"

	beStrings "github.com/go-enjin/be/pkg/strings"
)

func (re *renderEnjin) prepareAnchorFieldData(field map[string]interface{}) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	data["Type"] = "a"
	if href, ok := field["href"].(string); ok {
		data["Href"] = href
	} else {
		data["Href"] = "#"
	}
	if data["Text"], err = re.renderInlineFieldText(field); err != nil {
		return
	}
	if data["Text"] == "" {
		data["Text"] = data["Href"]
	}

	decorated := false
	if v, ok := field["decorated"].(string); ok {
		if beStrings.IsTrue(v) {
			decorated = true
		}
	}
	data["Decorated"] = decorated
	if attrs, classes, _, ok := re.parseFieldAttributes(field); ok {
		if decorated {
			classes = append(classes, "decorated")
		}
		data["Attributes"] = re.finalizeFieldAttributes(attrs)
	} else if decorated {
		data["Attributes"] = re.finalizeFieldAttributes(map[string]interface{}{
			"class": "decorated",
		})
	}

	re.finalizeFieldData(data, field, "type", "href", "text", "decorated", "attributes")
	return
}

func (re *renderEnjin) prepareFontAwesomeIconFieldData(field map[string]interface{}) (data map[string]interface{}, err error) {
	var ok bool
	data = make(map[string]interface{})
	data["Type"] = "i"
	if data["Class"], ok = field["class"]; !ok {
		data["Class"] = "fa-solid"
	}
	if data["Id"], ok = field["id"]; !ok {
		err = fmt.Errorf("font-awesome icon missing id: %+v", field)
		return
	}
	re.finalizeFieldData(data, field, "type", "class", "id")
	return
}

func (re *renderEnjin) prepareFigureFieldData(field map[string]interface{}) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	data["Type"] = "figure"
	if data["Text"], err = re.renderInlineFieldText(field); err != nil {
		return
	}
	re.finalizeFieldData(data, field, "type", "text")
	return
}

func parseKeyIntVal(key string, data map[string]interface{}) (v int, ok bool) {
	if i, found := data[key]; found {
		switch t := i.(type) {
		case int:
			v = t
			ok = true
			return
		case string:
			if s, err := strconv.Atoi(t); err == nil {
				v = s
				ok = true
				return
			}
		}
	}
	return
}

func (re *renderEnjin) prepareImageFieldData(field map[string]interface{}) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	data["Type"] = "img"
	data["Src"], _ = field["src"].(string)
	data["Alt"], _ = field["alt"].(string)
	var width, height int
	var hasWidth, hasHeight bool
	if width, hasWidth = parseKeyIntVal("width", field); hasWidth {
		data["Width"] = width
	}
	if height, hasHeight = parseKeyIntVal("height", field); hasHeight {
		data["Height"] = height
	}
	if attrs, _, _, ok := re.parseFieldAttributes(field); ok {
		if hasWidth {
			delete(attrs, "width")
		}
		if hasHeight {
			delete(attrs, "height")
		}
		data["Attributes"] = re.finalizeFieldAttributes(attrs)
	}
	re.finalizeFieldData(data, field, "type", "src", "alt", "attributes")
	return
}

func (re *renderEnjin) preparePictureFieldData(field map[string]interface{}) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	data["Type"] = "picture"
	dataDefault := make(map[string]interface{})
	if defaultMap, ok := field["default"].(map[string]interface{}); !ok {
		err = fmt.Errorf("picture field missing default image: %v", field)
		return
	} else {
		dataDefault["Type"] = "img"
		if dataDefault["Src"], ok = defaultMap["src"]; !ok {
			err = fmt.Errorf("picture field missing default img src: %v", defaultMap)
			return
		}
		re.finalizeFieldData(dataDefault, defaultMap, "type", "src")
		data["Default"] = dataDefault
	}
	var dataSources []map[string]interface{}
	if sources, ok := field["sources"].([]interface{}); ok {
		for _, si := range sources {
			if source, ok := si.(map[string]interface{}); ok {
				src := make(map[string]interface{})
				src["Type"] = "source"
				if src["Srcset"], ok = source["srcset"]; !ok {
					err = fmt.Errorf("picture field source missing srcset: %v", field)
					return
				}
				if src["Media"], ok = source["media"]; !ok {
					err = fmt.Errorf("picture field source missing media: %v", field)
					return
				}
				dataSources = append(dataSources, src)
			}
		}
	}
	data["Sources"] = dataSources
	re.finalizeFieldData(data, field, "type", "sources", "default")
	return
}

func (re *renderEnjin) prepareInputFieldData(field map[string]interface{}) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	data["Type"] = "input"
	if input, ok := field["input"]; ok {
		data["Input"] = input
	} else {
		err = fmt.Errorf("input field missing input type: %+v", field)
		return
	}
	data["Value"], _ = field["value"]
	re.finalizeFieldData(data, field, "type", "input", "value")
	return
}

func (re *renderEnjin) prepareSelectFieldData(field map[string]interface{}) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	data["Type"] = "select"

	if list, ok := field["placeholder"].([]interface{}); ok {
		if html, e := re.renderInlineFieldList(list); e != nil {
			err = fmt.Errorf("error rendering select placeholder: %v", e)
			return
		} else {
			data["Placeholder"] = html
		}
	} else if text, ok := field["placeholder"].(string); ok {
		data["Placeholder"] = text
	}

	if list, ok := field["options"].([]interface{}); ok {
		if html, e := re.renderInlineFieldList(list); e != nil {
			err = fmt.Errorf("error rendering select options: %v", e)
			return
		} else {
			data["Options"] = html
		}
	}

	re.finalizeFieldData(data, field, "type", "placeholder", "options")
	return
}

func (re *renderEnjin) prepareOptionGroupFieldData(field map[string]interface{}) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	data["Type"] = "optgroup"
	data["Label"], _ = field["label"].(string)

	if list, ok := field["options"].([]interface{}); ok {
		if html, e := re.renderInlineFieldList(list); e != nil {
			err = fmt.Errorf("error rendering optgroup options: %v", e)
			return
		} else {
			data["Options"] = html
		}
	}

	re.finalizeFieldData(data, field, "type", "label", "options")
	return
}

func (re *renderEnjin) prepareOptionFieldData(field map[string]interface{}) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	data["Type"] = "option"
	data["Value"], _ = field["value"].(string)

	if list, ok := field["text"].([]interface{}); ok {
		if html, e := re.renderInlineFieldList(list); e != nil {
			err = fmt.Errorf("error rendering option: %v", e)
			return
		} else {
			data["Text"] = html
		}
	}

	re.finalizeFieldData(data, field, "type", "value", "text")
	return
}