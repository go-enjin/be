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

package funcmaps

import (
	"regexp"
	"sort"
	"strings"

	"github.com/maruel/natural"

	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

func SplitString(v, delim string) (parts []string) {
	parts = strings.Split(v, delim)
	return
}

func FilterStrings(pattern string, values []string) (filtered []string) {
	if rx, err := regexp.Compile(pattern); err == nil {
		for _, value := range values {
			if rx.MatchString(value) {
				filtered = append(filtered, value)
			}
		}
	} else {
		log.ErrorF("error compiling filterStrings pattern: m!%v! - %v", pattern, err)
	}
	return
}

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

func SortedStrings(values []string) (sorted []string) {
	sorted = values[:]
	sort.Sort(natural.StringSlice(sorted))
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

func Iterate(argv ...int) (ch chan int) {
	ch = make(chan int)
	var from, inc, to int
	switch len(argv) {
	case 0:
		log.ErrorF("template is trying to iterate over nothing")
	case 1:
		from, inc, to = 0, 1, argv[0]
	case 2:
		from, inc, to = argv[0], 1, argv[1]
	default:
		from, inc, to = argv[0], argv[1], argv[2]
	}
	go func() {
		if to > 0 {
			for i := from; i < to; i += inc {
				ch <- i
			}
		}
		close(ch)
	}()
	return
}