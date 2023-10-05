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
	"cmp"
	"fmt"
	"html/template"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/maruel/natural"

	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/regexps"
	"github.com/go-enjin/be/pkg/slices"
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

func ValuesSortedByKeys[V interface{}](data map[string]V) (values []V) {
	for _, k := range SortedKeys(data) {
		values = append(values, data[k])
	}
	return
}

// SortedKeyLengths returns the list of keys natural sorted and from longest to
// shortest
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
	sort.Sort(sort.Reverse(natural.StringSlice(keys)))
	return
}

func OrderedKeys[K cmp.Ordered, V interface{}](data map[K]V) (keys []K) {
	for key, _ := range data {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) (less bool) {
		less = cmp.Less(keys[i], keys[j]) // i vs j
		return
	})
	return
}

func ReverseOrderedKeys[K cmp.Ordered, V interface{}](data map[K]V) (keys []K) {
	for key, _ := range data {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) (less bool) {
		less = cmp.Less(keys[j], keys[i]) // j vs i
		return
	})
	return
}

func SortedKeysByLastKeyword[V interface{}](data map[string]V) (keys []string) {
	lookup := make(map[string]string)
	for key, _ := range data {
		keywords := regexps.RxKeywords.FindAllString(key, -1)
		lookup[key] = keywords[len(keywords)-1]
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) (less bool) {
		less = natural.Less(lookup[keys[i]], lookup[keys[j]])
		return less
	})
	return
}

func SortedKeysByLastName[V interface{}](data map[string]V) (keys []string) {
	lookup := make(map[string]string)
	for key, _ := range data {
		lookup[key] = beStrings.LastName(key)
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) (less bool) {
		less = natural.Less(lookup[keys[i]], lookup[keys[j]])
		return less
	})
	return
}

func Keys[V interface{}](data map[string]V) (keys []string) {
	for key, _ := range data {
		keys = append(keys, key)
	}
	return
}

func AnyKeys[V interface{}](data map[interface{}]V) (keys []interface{}) {
	for key, _ := range data {
		keys = append(keys, key)
	}
	return
}

func TypedKeys[T comparable, V interface{}](data map[T]V) (keys []T) {
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

func ParseKeySlice(input string) (key string, idx int, ok bool) {
	var err error
	if ok = regexps.RxKeySlice.MatchString(input); ok {
		km := regexps.RxKeySlice.FindStringSubmatch(input)
		if km[2] == "" {
			idx = -1
		} else {
			if idx, err = strconv.Atoi(km[2]); err != nil {
				ok = false
				idx = -1
				return
			}
		}
		key = km[1]
	}
	return
}

func Set(key string, value interface{}, m map[string]interface{}) (err error) {
	if key == "" {
		return
	} else if key != "" && key[0] != '.' {
		m[key] = value
		return
	}
	keys := strings.Split(key[1:], ".")
	keysLen := len(keys)
	switch keysLen {
	case 0: // nop
	case 1:
		if name, idx, ok := ParseKeySlice(keys[0]); ok {
			if vs := Get(name, m); vs != nil {
				if slice, ok := vs.([]interface{}); ok {
					count := len(slice)
					if idx == -1 || idx == count {
						// append
						slice = append(slice, value)
					} else if idx > count {
						for i := count; i < idx; i++ {
							slice = append(slice, nil)
						}
						slice = append(slice, value)
					} else {
						// overwrite
						slice[idx] = value
					}
					err = Set(name, slice, m)
				} else {
					// hmm
				}
			} else {
				var slice []interface{}
				if idx <= 0 {
					slice = append(slice, value)
				} else {
					for i := 0; i < idx; i++ {
						slice = append(slice, nil)
					}
					slice = append(slice, value)
				}
				err = Set(name, slice, m)
			}
		} else {
			err = Set(keys[0], value, m)
		}
		return
	default:
		if name, idx, ok := ParseKeySlice(keys[0]); ok {
			var list []interface{}
			if v, ok := m[name]; ok {
				if vl, ok := v.([]interface{}); ok {
					list = vl
				} else {
					err = fmt.Errorf("unexpected sub-context list type: %T", v)
					return
				}
			} else {
				list = []interface{}{}
				if err = Set(name, list, m); err != nil {
					return
				}
			}

			var mm map[string]interface{}
			count := len(list)
			if idx == -1 || idx >= count {
				// append
				if idx >= count {
					// append more
					for i := count - 1; i < idx; i++ {
						list = append(list, map[string]interface{}{})
					}
				}
				mm = list[idx].(map[string]interface{})
				if err = Set(name, list, m); err != nil {
					return
				}
			} else {
				// existing
				if vm, ok := list[idx].(map[string]interface{}); ok {
					mm = vm
				} else {
					err = fmt.Errorf("unexpected sub-context value type: %T", list[idx])
					return
				}
			}

			err = Set("."+strings.Join(keys[1:], "."), value, mm)
			return
		}

		var mm map[string]interface{}
		if v, ok := m[keys[0]]; ok {
			if t, ok := v.(map[string]interface{}); ok {
				mm = t
			} else {
				err = fmt.Errorf("unexpected sub-context value type: %T", v)
				return
			}
		} else {
			mm = make(map[string]interface{})
			if err = Set(keys[0], mm, m); err != nil {
				return
			}
		}
		err = Set("."+strings.Join(keys[1:], "."), value, mm)
	}
	return
}

func Get(key string, m map[string]interface{}) (value interface{}) {
	if key == "" {
		return
	} else if key != "" && key[0] != '.' {
		if v, ok := m[key]; ok {
			value = v
		}
		return
	}
	keys := strings.Split(key[1:], ".")
	switch len(keys) {
	case 0: // nop
	case 1:
		if name, idx, ok := ParseKeySlice(keys[0]); ok {
			if vs := Get(name, m); vs != nil {
				if slice, ok := vs.([]interface{}); ok {
					if len(slice) < idx {
						return
					}
					value = slice[idx]
				}
			}
		} else {
			value = Get(keys[0], m)
		}
	default:
		if v, ok := m[keys[0]]; ok {
			if ms, ok := v.(map[string]string); ok {
				value, _ = ms[keys[1]]
			} else if mm, ok := v.(map[string]interface{}); ok {
				value = Get("."+strings.Join(keys[1:], "."), mm)
			}
		}
	}
	return
}

func Delete(key string, m map[string]interface{}) {
	if key == "" {
		return
	} else if key != "" && key[0] != '.' {
		if _, ok := m[key]; ok {
			delete(m, key)
		}
		return
	}
	keys := strings.Split(key[1:], ".")
	switch len(keys) {
	case 0: // nop
	case 1:
		if name, idx, ok := ParseKeySlice(keys[0]); ok {
			if vs := Get(name, m); vs != nil {
				if slice, ok := vs.([]interface{}); ok {
					count := len(slice)
					if idx > -1 && idx < count {
						newSlice := make([]interface{}, 0)
						newSlice = append(newSlice, slice[:idx]...)
						if idx < count-1 {
							newSlice = append(newSlice, slice[idx+1:]...)
						}
						_ = Set(name, newSlice, m)
					}
				}
			}
		} else {
			Delete(keys[0], m)
		}
	default:
		if v, ok := m[keys[0]]; ok {
			if ms, ok := v.(map[string]string); ok {
				delete(ms, keys[1])
			} else if mm, ok := v.(map[string]interface{}); ok {
				Delete("."+strings.Join(keys[1:], "."), mm)
			}
		}
	}
	return
}

func TransformMapAnyToString(input map[string]interface{}) (output map[string]string) {
	output = make(map[string]string)
	for k, v := range input {
		switch t := v.(type) {
		case string:
			output[k] = t
		default:
			output[k] = fmt.Sprintf("%v", t)
		}
	}
	return
}

func TransformAnyToStringSlice(input interface{}) (output []string, ok bool) {
	ok = true
	switch t := input.(type) {
	case string:
		output = append(output, t)
	case []string:
		output = append(output, t...)
	case []interface{}:
		for _, v := range t {
			if s, check := v.(string); check {
				output = append(output, s)
			}
		}
	default:
		ok = false
	}
	return
}

type SliceValues interface {
	~bool |
		~string | ~[]byte |
		~float32 | ~float64 |
		~complex64 | ~complex128 |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

func DeepCopySlice[V SliceValues](src []V) (dst []V) {
	dst = append(dst, src...)
	return
}

func DeepCopy(src map[string]interface{}) (dst map[string]interface{}) {
	dst = map[string]interface{}{}
	for k, v := range src {
		switch t := v.(type) {

		case map[string]interface{}:
			dst[k] = DeepCopy(t)
		case map[string]bool:
			m := make(map[string]bool)
			for tk, tv := range t {
				m[tk] = tv
			}
			dst[k] = m
		case map[string]int:
			m := make(map[string]int)
			for tk, tv := range t {
				m[tk] = tv
			}
			dst[k] = m
		case map[string]string:
			m := make(map[string]string)
			for tk, tv := range t {
				m[tk] = tv
			}
			dst[k] = m

		case []bool:
			dst[k] = DeepCopySlice(t)
		case []string:
			dst[k] = DeepCopySlice(t)
		case []int:
			dst[k] = DeepCopySlice(t)
		case []int8:
			dst[k] = DeepCopySlice(t)
		case []int16:
			dst[k] = DeepCopySlice(t)
		case []int32:
			dst[k] = DeepCopySlice(t)
		case []int64:
			dst[k] = DeepCopySlice(t)
		case []uint:
			dst[k] = DeepCopySlice(t)
		case []uint8:
			dst[k] = DeepCopySlice(t)
		case []uint16:
			dst[k] = DeepCopySlice(t)
		case []uint32:
			dst[k] = DeepCopySlice(t)
		case []uint64:
			dst[k] = DeepCopySlice(t)
		case []float32:
			dst[k] = DeepCopySlice(t)
		case []float64:
			dst[k] = DeepCopySlice(t)

		case bool, string,
			float32, float64,
			complex64, complex128,
			int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64:
			dst[k] = t

		default:
			dst[k] = t
		}
	}
	return
}

func MakeTypedKey[K comparable, L comparable, V interface{}, M map[L]V](key K, m map[K]M) (made bool) {
	if _, present := m[key]; !present {
		var l L
		var v V
		kt, vt := reflect.TypeOf(l), reflect.TypeOf(v)
		mt := reflect.MapOf(kt, vt)
		mv := reflect.MakeMapWithSize(mt, 0)
		mi := mv.Interface()
		m[key], _ = mi.(M)
		return true
	}
	return
}