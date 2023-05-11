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

package feature

import (
	"sort"

	"github.com/fvbommel/sortorder"
)

type Tags []Tag

// Has returns true if the list of Tags includes the given tag
func (t Tags) Has(tag Tag) bool {
	for _, tt := range t {
		if tag == tt {
			return true
		}
	}
	return false
}

// Append returns a list with the given tag appended
func (t Tags) Append(tag Tag) Tags {
	if !t.Has(tag) {
		return append(t, tag)
	}
	return t
}

// Len returns the number of tags
func (t Tags) Len() int {
	return len(t)
}

// Unique returns a list of unique tags, maintaining their original order
func (t Tags) Unique() (tags Tags) {
	lookup := make(map[string]bool)
	for _, tag := range t {
		lookup[string(tag)] = true
	}
	for _, tag := range t {
		if _, keep := lookup[string(tag)]; keep {
			tags = append(tags, tag)
		}
	}
	return
}

// Strings returns the list of Tags as a string slice
func (t Tags) Strings() (names []string) {
	for _, tag := range t {
		names = append(names, string(tag))
	}
	return
}

// StringsAsTags returns a list of Tags based on the names given
func StringsAsTags(names []string) (tags Tags) {
	for _, name := range names {
		tags = tags.Append(Tag(name))
	}
	return
}

// SortedFeatureTags returns a sortorder.Natural list of Tag keys
func SortedFeatureTags[V interface{}](data map[Tag]V) (tags Tags) {
	var keys []string
	for key, _ := range data {
		keys = append(keys, key.String())
	}
	sort.Sort(sortorder.Natural(keys))
	tags = StringsAsTags(keys)
	return
}