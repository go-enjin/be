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

package strings

import (
	"strings"
	"unicode"

	"github.com/iancoleman/strcase"
)

func PathToSnake(path string) (snake string) {
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	path = strings.ReplaceAll(path, "/", "--")
	snake = strcase.ToSnake(path)
	return
}

func ToCamelWords(text string) (capitalized string) {
	// inspired by: https://stackoverflow.com/a/70284562
	var within bool
	characters := make([]rune, len(text))
	for idx, character := range []rune(text) {
		if isLetter := unicode.IsLetter(character); isLetter && !within {
			characters[idx] = unicode.ToTitle(character)
			within = true
		} else if isLetter {
			characters[idx] = character
		} else if within = unicode.IsNumber(character); within {
			characters[idx] = character
		} else {
			characters[idx] = character
			within = false
		}
	}
	capitalized = string(characters)
	return
}

func ToSpaced(s string) (spaced string) {
	spaced = strcase.ToDelimited(s, ' ')
	return
}

func ToSpacedCamel(s string) (spacedCamel string) {
	spaced := strcase.ToDelimited(s, ' ')
	spacedCamel = ToCamelWords(spaced)
	return
}

func ToDeepKey(s string) (deepKey string) {
	parts := strings.Split(strings.TrimPrefix(s, "."), ".")
	for _, part := range parts {
		deepKey += "." + strcase.ToKebab(part)
	}
	return
}

func ToDeepVar(s string) (deepKey string) {
	parts := strings.Split(strings.TrimPrefix(s, "."), ".")
	for _, part := range parts {
		deepKey += "." + strcase.ToKebab(part)
	}
	return
}
