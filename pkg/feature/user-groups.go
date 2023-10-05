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
	"strings"
)

// Groups are the collection of one or more groups
type Groups []Group

func NewGroupsFromStringNL(newlines string) (groups Groups) {
	groups = groups.AppendString(strings.Split(newlines, "\n")...)
	return
}

func NewGroupsFromStrings(slice ...string) (groups Groups) {
	groups = groups.AppendString(slice...)
	return
}

func (g Groups) Len() int {
	return len(g)
}

func (g Groups) String() (s string) {
	for idx, group := range g {
		if idx > 0 {
			s += " "
		}
		s += group.String()
	}
	return
}

func (g Groups) Has(group Group) (present bool) {
	for _, gg := range g {
		if present = gg == group; present {
			return
		}
	}
	return
}

func (g Groups) Remove(groups ...Group) (modified Groups) {
	remove := Groups(groups)
	for _, maybeKeep := range g {
		if !remove.Has(maybeKeep) {
			modified = append(modified, maybeKeep)
		}
	}
	return
}

func (g Groups) Append(groups ...Group) (modified Groups) {
	modified = g
	for _, group := range groups {
		if group != "" && !g.Has(group) {
			modified = append(modified, group)
		}
	}
	return
}

func (g Groups) AppendString(names ...string) (modified Groups) {
	modified = g
	for _, name := range names {
		if name != "" {
			if namedGroup := NewGroup(name); !modified.Has(namedGroup) {
				modified = append(modified, namedGroup)
			}
		}
	}
	return
}

func (g Groups) AsNewlines() (newlines string) {
	for idx, group := range g {
		if idx > 0 {
			newlines += "\n"
		}
		newlines += group.String()
	}
	return
}