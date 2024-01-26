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

	"github.com/maruel/natural"
)

type Tags []Tag

// StringsAsTags returns a list of Tags based on the names given
func StringsAsTags(names []string) (tags Tags) {
	for _, name := range names {
		tags = tags.Append(Tag(name))
	}
	return
}

// SortedFeatureTags returns a natural.StringSlice list of Tag keys
func SortedFeatureTags[V interface{}](data map[Tag]V) (tags Tags) {
	var keys []string
	for key := range data {
		keys = append(keys, key.String())
	}
	sort.Sort(natural.StringSlice(keys))
	tags = StringsAsTags(keys)
	return
}

// Has returns true if the list of Tags includes the given tag
func (t Tags) Has(tag Tag) (present bool) {
	for _, tt := range t {
		if present = tag == tt; present {
			return
		}
	}
	return
}

// Find returns the first tag matching the given name exactly, or matching by kebab-cased comparison
func (t Tags) Find(name string) (found Tag, ok bool) {
	tag := Tag(name)
	kebab := tag.Kebab()
	for _, found = range t {
		if ok = tag == found; ok {
			return
		} else if ok = kebab == found.Kebab(); ok {
			return
		}
	}
	found = NilTag
	return
}

// Append returns a (unique) list with the given tags appended
func (t Tags) Append(tags ...Tag) (list Tags) {
	list = t[:]
	for _, tag := range tags {
		if !list.Has(tag) {
			list = append(list, tag)
		}
	}
	return
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
