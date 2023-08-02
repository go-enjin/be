/*
 * Copyright (c) 2023  The Go-Enjin Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package funcmaps

import (
	"fmt"
	"strings"

	beContext "github.com/go-enjin/be/pkg/context"
)

// functions based on hugo documentation for theme portability

// Dict mimics the hugo "dict" template function
//   - accepts a list of key/value argument pairs, errors if list is odd
//   - keys can be strings or string slices:
//   - string slices set deep values, ie: ["one","two","three"] produces map[one]map[two]map[three]=value
//   - string keys also support enjin context deep keys, ie: .One.Two.Three is the same as the slice example above
func Dict(argv ...interface{}) (dict beContext.Context, err error) {
	var argc int
	if argc = len(argv); argc%2 != 0 {
		err = fmt.Errorf("odd number of dictionary key/value arguments")
		return
	}

	dict = beContext.Context{}

	for i := 0; i < argc; i += 2 {

		switch argt := argv[i].(type) {

		case string:
			if err = dict.SetKV(argt, argv[i+1]); err != nil {
				return
			}

		case []string:
			if err = dict.SetKV("."+strings.Join(argt, "."), argv[i+1]); err != nil {
				return
			}

		default:
			err = fmt.Errorf("invalid key argument type: %T", argt)
			return

		}

	}

	return
}

func MakeSlice(values ...interface{}) (output []string) {
	for _, value := range values {
		output = append(output, ToString(value))
	}
	return
}