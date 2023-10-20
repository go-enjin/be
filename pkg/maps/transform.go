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

	"github.com/iancoleman/strcase"
)

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