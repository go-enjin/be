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

package tmpl

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/menu"
)

func Element(data map[string]interface{}) (html template.HTML) {
	if eo := ElementOpen(data); eo == "" {
		log.ErrorF("element data missing Type property: %+v", data)
		return
	} else {
		html += eo
	}
	if v, ok := data["Text"]; ok {
		switch t := v.(type) {
		case string:
			html += template.HTML(t)
		case template.HTML:
			html += t
		default:
			log.ErrorF("unknown element text type: %T %+v", t, t)
		}
	} else {
		log.ErrorF("element data missing Text property: %+v", data)
	}
	if ec := ElementClose(data); ec != "" {
		html += ec
	} else {
		log.ErrorF("element failed to close, yet was able to be opened: %+v", data)
	}
	return
}

func ElementAttributes(value interface{}) (html template.HTMLAttr) {
	var parts []string
	switch data := value.(type) {
	case menu.Item:
		if data.Attributes != nil {
			for k, v := range data.Attributes {
				parts = append(parts, fmt.Sprintf(`%v="%v"`, k, v))
			}
		}
	case []template.HTMLAttr:
		for idx, v := range data {
			if idx > 0 {
				html += " "
			}
			html += v
		}
	case map[string]interface{}:
		if ai, found := data["Attributes"]; found && ai != nil {
			switch t := ai.(type) {
			case string:
				parts = append(parts, t)
			case []interface{}:
				for _, i := range t {
					if v, ok := i.(string); ok {
						parts = append(parts, v)
					}
				}
			case map[string]interface{}:
				for k, i := range t {
					if v, ok := i.(string); ok {
						parts = append(parts, fmt.Sprintf(`%v="%v"`, k, v))
					}
				}
			case []template.HTMLAttr:
				for idx, i := range t {
					if idx > 0 {
						html += " "
					}
					html += i
				}
				return
			default:
				log.ErrorF("unknown attributes type: %T %+v", t, ai)
			}
		}
	case nil: // nop
	default:
		log.ErrorF("unknown attributes data type: %T %+v", data, data)
	}
	if len(parts) > 0 {
		html += template.HTMLAttr(strings.Join(parts, " "))
	}
	return
}

func ElementOpen(data map[string]interface{}) (html template.HTML) {
	if dataType, ok := data["Type"]; ok {
		switch dt := dataType.(type) {
		case string:
			html = "<"
			html += template.HTML(dt)
			if attrs := ElementAttributes(data); len(attrs) > 0 {
				html += " "
				html += template.HTML(attrs)
			}
			html += ">"
		case template.HTML:
			html = "<"
			html += dt
			if attrs := ElementAttributes(data); len(attrs) > 0 {
				html += " "
				html += template.HTML(attrs)
			}
			html += ">"
		case template.HTMLAttr:
			html = "<"
			html += template.HTML(dt)
			if attrs := ElementAttributes(data); len(attrs) > 0 {
				html += " "
				html += template.HTML(attrs)
			}
			html += ">"
		default:
			log.ErrorF("element open invalid type property: %T %+v", dt, dt)
		}
	} else {
		log.ErrorF("element open missing type property: %+v", data)
	}
	return
}

func ElementClose(data map[string]interface{}) (html template.HTML) {
	if dataType, ok := data["Type"]; ok {
		switch dt := dataType.(type) {
		case string:
			html = "</"
			html += template.HTML(dt)
			html += ">"
		case template.HTML:
			html = "</" + dt + ">"
		case template.HTMLAttr:
			html = "</"
			html += template.HTML(dt)
			html += ">"
		}
	} else {
		log.ErrorF("element close missing type property: %+v", data)
	}
	return
}