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

package funcs

import (
	"strings"

	"github.com/go-enjin/be/pkg/maps"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

func StringsAsList(v ...string) (list []string) {
	list = v
	return
}

func ReverseStrings(v []string) (reversed []string) {
	for i := len(v) - 1; i >= 0; i-- {
		reversed = append(reversed, v[i])
	}
	return
}

func SortedFirstLetters(values []interface{}) (firsts []string) {
	cache := make(map[string]bool)
	for _, v := range values {
		if value, ok := v.(string); ok && value != "" {
			char := strings.ToLower(string(value[0]))
			cache[char] = true
		}
	}
	firsts = maps.SortedKeys(cache)
	return
}

func SortedLastNameFirstLetters(values []interface{}) (firsts []string) {
	cache := make(map[string]bool)
	for _, v := range values {
		if value, ok := v.(string); ok && value != "" {
			if word := beStrings.LastName(value); word != "" {
				char := strings.ToLower(string(word[0]))
				cache[char] = true
			}
		}
	}
	firsts = maps.SortedKeys(cache)
	return
}