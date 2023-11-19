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
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/maruel/natural"

	"github.com/go-enjin/be/pkg/regexps"
	"github.com/go-enjin/be/pkg/slices"
)

var nameSuffixes = []string{
	"jr", "sr",
	"i", "ii", "iii", "iv", "v", "vi", "vii", "viii", "ix", "x",
	"xi", "xii", "xiii", "xiv", "xv", "xvi", "xvii", "xviii", "xix", "xx",
	"xxi", "xxii", "xxiii", "xxiv", "xxv", "xxvi", "xxvii", "xxviii", "xxix", "xxx",
	"xxxi", "xxxii", "xxxiii", "xxxiv",
}

// TODO: figure out a better way of decoding arbitrary "full name" strings, similarly to date/time language

func FirstName(fullName string) (firstName string) {
	if names := regexps.RxKeywords.FindAllString(fullName, -1); len(names) > 0 {
		for i := len(names) - 1; i >= 0; i-- {
			firstName = names[i]
			switch strings.ToLower(firstName) {
			case "dr", "mr":
				continue
			}
			break
		}
	}
	return
}

func LastName(fullName string) (lastName string) {
	if names := regexps.RxKeywords.FindAllString(fullName, -1); len(names) > 0 {
		for i := len(names) - 1; i >= 0; i-- {
			name := names[i]
			if slices.Within(strings.ToLower(name), nameSuffixes) {
				continue
			}
			lastName = name
			break
		}
	}
	return
}

func SortedByLastName(data []string) (keys []string) {
	lookup := make(map[string]string)
	for _, key := range data {
		lookup[key] = LastName(key)
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) (less bool) {
		less = natural.Less(lookup[keys[i]], lookup[keys[j]])
		return less
	})
	return
}

func NameFromEmail(email string) (name string) {
	if before, after, found := strings.Cut(email, "@"); found {
		name = strcase.ToCamel(strcase.ToDelimited(before, ' '))
		name += " @"
		if parts := strings.Split(after, "."); len(parts) > 1 {
			name += strcase.ToCamel(parts[len(parts)-2])
		} else {
			name += strcase.ToCamel(after)
		}
	}
	return
}

var (
	RxNameSuffix = regexp.MustCompile(`\((\d+)\)\s*$`)
)

// IncrementFileName will return the given label with a specific number suffix incremented. If the label does not end with
// the pattern `\(\d+\)\s*$`, one is appended with a space, if the pattern matches no spaces are added and the
// digit is modified and any trailing space removed
func IncrementFileName(name string) (modified string) {
	if RxNameSuffix.MatchString(name) {
		m := RxNameSuffix.FindAllStringSubmatch(name, 1)
		d, _ := strconv.Atoi(m[0][1])
		modified = RxNameSuffix.ReplaceAllString(name, fmt.Sprintf("(%d)", d+1))
	} else {
		modified = name + " (1)"
	}
	return
}