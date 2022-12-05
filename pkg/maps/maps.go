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

package maps

import (
	"fmt"
	"html/template"
	"sort"
	"strconv"
	"strings"

	"github.com/fvbommel/sortorder"
	"github.com/iancoleman/strcase"
	"github.com/maruel/natural"

	"github.com/go-enjin/be/pkg/log"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

func ExtractBoolValue(v interface{}) (b bool) {
	switch t := v.(type) {
	case int:
		b = t != 0
	case uint:
		b = t != 0
	case float64:
		b = t != 0.0
	case bool:
		b = t
	case string:
		b = beStrings.IsTrue(t)
	}
	return
}

func ExtractEnumValue(key string, upper bool, enums []string, data map[string]interface{}) (value string, err error) {
	if v, ok := data[key]; ok {
		switch t := v.(type) {
		case string:
			if upper {
				t = strings.ToUpper(t)
			} else {
				t = strings.ToLower(t)
			}
			for _, enum := range enums {
				if t == enum {
					value = enum
					return
				}
			}
		default:
			err = fmt.Errorf("unsupported enum value structure: %T", t)
		}
	}
	return
}

func ExtractIntValue(key string, data map[string]interface{}) (value int, err error) {
	if v, ok := data[key]; ok {
		switch t := v.(type) {
		case string:
			if value, err = strconv.Atoi(t); err != nil {
				err = fmt.Errorf("error parsing %v integer string: %v", key, err)
				return
			}
		case float64:
			value = int(t)
		case float32:
			value = int(t)
		case uint:
			value = int(t)
		case int:
			value = t
		case int64:
			value = int(t)
		default:
			err = fmt.Errorf("unsupported %v integer type: %T %v", key, v, v)
			return
		}
	}
	return
}

func ParseKeyIntValue(key string, data map[string]interface{}) (v int, ok bool) {
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
			if a, e := beStrings.ParseHtmlTagAttributes(attrs); e != nil {
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
		case beStrings.StringInStrings(key, skip...):
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
			attributes = append(attributes, template.HTMLAttr(fmt.Sprintf(`%v="%v"`, k, beStrings.EscapeHtmlAttribute(t))))
		case template.HTMLAttr:
			attributes = append(attributes, template.HTMLAttr(fmt.Sprintf(`%v="%v"`, k, t)))
		default:
			err = fmt.Errorf("unsupported type: %T %+v", t, t)
		}
	}
	return
}

func DebugWalk(thing map[string]interface{}) (results string) {
	var walk func(depth string, tgt map[string]interface{}) (out string)
	walk = func(depth string, tgt map[string]interface{}) (out string) {
		for k, v := range tgt {
			switch t := v.(type) {
			case map[string]interface{}:
				out += walk(fmt.Sprintf("%v%v.", depth, k), t)
			default:
				out += fmt.Sprintf("%v%v", depth, k)
			}
		}
		return
	}
	results = walk("\n * ", thing)
	return
}

func SortedKeyLengths[V interface{}](data map[string]V) (keys []string) {
	for key, _ := range data {
		keys = append(keys, key)
	}
	// longest -> shortest, natsort same lengths
	sort.Slice(keys, func(i, j int) (less bool) {
		if il, jl := len(keys[i]), len(keys[j]); il == jl {
			less = natural.Less(keys[i], keys[j])
		} else {
			less = il > jl
		}
		return
	})
	return
}

func SortedKeys[V interface{}](data map[string]V) (keys []string) {
	for key, _ := range data {
		keys = append(keys, key)
	}
	sort.Sort(natural.StringSlice(keys))
	return
}

func ReverseSortedKeys[V interface{}](data map[string]V) (keys []string) {
	for key, _ := range data {
		keys = append(keys, key)
	}
	sort.Sort(sort.Reverse(sortorder.Natural(keys)))
	return
}

func Keys[V interface{}](data map[string]V) (keys []string) {
	for key, _ := range data {
		keys = append(keys, key)
	}
	return
}

func CamelizeKeys[V interface{}](data map[string]V) (camelized map[string]V) {
	camelized = make(map[string]V)
	for k, v := range data {
		camel := strcase.ToCamel(k)
		camelized[camel] = v
	}
	return
}

func KebabKeys[V interface{}](data map[string]V) (kebabed map[string]V) {
	kebabed = make(map[string]V)
	for k, v := range data {
		kebab := strcase.ToKebab(k)
		kebabed[kebab] = v
	}
	return
}

func IsMap(v interface{}) (ok bool) {
	ok = strings.HasPrefix(fmt.Sprintf("%T", v), "map[")
	return
}