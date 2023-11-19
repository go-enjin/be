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
	"strings"

	"github.com/go-enjin/be/pkg/values"
)

func IsMap(v interface{}) (ok bool) {
	ok = strings.HasPrefix(values.TypeOf(v), "map[")
	return
}

func PrettyMap[T comparable, V interface{}](m map[T]V) (pretty string) {
	value := fmt.Sprintf("%#v", m)
	last := len(value) - 1
	pretty += "{"
	var keeping bool
	for i := 0; i < last; i++ {
		if !keeping {
			this, next := value[i], value[i+1]
			keeping = this == '{' && next != '}'
			continue
		}
		pretty += string(value[i])
	}
	pretty += "}"
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