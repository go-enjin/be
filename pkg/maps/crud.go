// Copyright (c) 2023  The Go-Enjin Authors
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
	"reflect"
	"strings"
)

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
				list = make([]interface{}, 0)
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
						list = append(list, make(map[string]interface{}))
					}
				}
				mm, _ = list[idx].(map[string]interface{})
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