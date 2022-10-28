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

package page

import (
	"sort"
	"strings"

	beStrings "github.com/go-enjin/be/pkg/strings"
	"github.com/go-enjin/be/pkg/types/theme-types"
)

var (
	knownFormats = make(map[string]types.Format)
	Extensions   = make([]string, 0)
)

func MatchFormatExtension(filename string) (format types.Format, match string) {
	for _, extension := range Extensions {
		if strings.HasSuffix(filename, "."+extension) {
			match = extension
			format, _ = knownFormats[extension]
			break
		}
	}
	return
}

func GetFormat(name string) (format types.Format) {
	if v, ok := knownFormats[name]; ok {
		format = v
	}
	return
}

func AddFormat(format types.Format) {
	extensions := format.Extensions()
	for _, extension := range extensions {
		knownFormats[extension] = format
		if !beStrings.StringInStrings(extension, Extensions...) {
			Extensions = append(Extensions, extension)
		}
	}
	sort.Sort(beStrings.SortByLengthDesc(Extensions))
}

func RemoveFormat(name string) {
	if format, ok := knownFormats[name]; ok {
		for _, extension := range format.Extensions() {
			delete(knownFormats, extension)
			if idx := beStrings.StringIndexInStrings(extension, Extensions...); idx >= 0 {
				Extensions = beStrings.RemoveIndexFromStrings(idx, Extensions)
			}
		}
	}
	sort.Sort(beStrings.SortByLengthDesc(Extensions))
}