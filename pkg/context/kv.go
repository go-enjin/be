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

package context

import (
	"github.com/iancoleman/strcase"

	"github.com/go-enjin/be/pkg/maps"
)

// GetKV looks for the key as given first and if not found looks for CamelCased, kebab-case and snake_cased variations;
// returning the actual key found and the generic value; returns an empty key and nil value if nothing found at all
func GetKV(c map[string]interface{}, key string) (k string, v interface{}) {
	if key != "" && key[0] == '.' {
		k = key
		v = maps.Get(c, key)
		return
	}

	k = key
	var ok bool
	if v, ok = c[k]; ok {
		return
	}
	k = strcase.ToCamel(key)
	if v, ok = c[k]; ok {
		return
	}
	k = strcase.ToKebab(key)
	if v, ok = c[k]; ok {
		return
	}
	k = strcase.ToSnake(key)
	if v, ok = c[k]; ok {
		return
	}
	return
}

// SetKV allows for .Deep.Variable.Names
func SetKV(c map[string]interface{}, key string, value interface{}) (err error) {
	if key != "" && key[0] == '.' {
		err = maps.Set(c, key, value)
		return
	}
	k := strcase.ToCamel(key)
	c[k] = value
	return
}

// DeleteKV deletes the key allowing for .Deep.Variable.Names
func DeleteKV(c map[string]interface{}, key string) (deleted bool) {
	if key != "" && key[0] == '.' {
		maps.Delete(c, key)
		return
	}
	if k, v := GetKV(c, key); v != nil {
		delete(c, k)
		return true
	}
	return false
}