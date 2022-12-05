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
	"strings"

	"github.com/go-enjin/be/pkg/regexps"
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
			if StringInSlices(strings.ToLower(name), nameSuffixes) {
				continue
			}
			lastName = name
			break
		}
	}
	return
}