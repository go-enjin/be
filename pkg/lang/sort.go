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

package lang

import (
	"sort"

	"github.com/maruel/natural"

	"github.com/go-enjin/golang-org-x-text/language"
)

func SortLanguageTags(tags []language.Tag) (sorted []language.Tag) {
	lookup := make(map[string]language.Tag)
	var keys []string
	for _, tag := range tags {
		lookup[tag.String()] = tag
		keys = append(keys, tag.String())
	}
	sort.Sort(natural.StringSlice(keys))
	for _, key := range keys {
		sorted = append(sorted, lookup[key])
	}
	return
}