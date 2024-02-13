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

package strings

import (
	"github.com/go-corelibs/htmlcss"
)

//func TitleCase(input string) (output string) {
//	first := true
//	output = regexps.RxWord.ReplaceAllStringFunc(
//		strings.ToLower(input),
//		func(word string) string {
//			if !first {
//				switch word {
//				case "with", "in", "of", "at", "a", "the":
//					return word
//				}
//			}
//			first = false
//			return strcase.ToCamel(word)
//		},
//	)
//	return
//}

func AddClassNamesToNjnBlock(data map[string]interface{}, classes ...string) map[string]interface{} {
	if v, ok := data["class"]; ok {
		switch t := v.(type) {
		case string:
			data["class"] = htmlcss.AddClassNames(t, classes...)
		case []interface{}:
			var list []interface{}
			for _, iface := range t {
				if vs, ok := iface.(string); ok {
					list = append(list, htmlcss.AddClassNames(vs, classes...))
				}
			}
			data["class"] = list
		}
	}
	return data
}
