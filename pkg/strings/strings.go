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
	"regexp"
	"strings"

	"github.com/iancoleman/strcase"
)

func StringInStrings(src string, dst ...string) bool {
	for _, v := range dst {
		if src == v {
			return true
		}
	}
	return false
}

func StringIndexInStrings(src string, dst ...string) int {
	for i, v := range dst {
		if src == v {
			return i
		}
	}
	return -1
}

func RemoveIndexFromStrings(idx int, slice []string) []string {
	if idx >= 0 && idx < len(slice) {
		if idx == 0 {
			return slice[1:]
		}
		return append(slice[:idx], slice[idx+1:]...)
	}
	return slice
}

var RxWord = regexp.MustCompile(`\w+`)

func TitleCase(input string) (output string) {
	first := true
	output = RxWord.ReplaceAllStringFunc(
		strings.ToLower(input),
		func(word string) string {
			if !first {
				switch word {
				case "with", "in", "of", "at", "a", "the":
					return word
				}
			}
			first = false
			return strcase.ToCamel(word)
		},
	)
	return
}

var RxBasicMimeType = regexp.MustCompile(`^\s*([^\s;]*)\s*.+?\s*$`)

func GetBasicMime(mime string) (basic string) {
	if RxBasicMimeType.MatchString(mime) {
		m := RxBasicMimeType.FindAllStringSubmatch(mime, 1)
		basic = m[0][1]
		return
	}
	basic = mime
	return
}