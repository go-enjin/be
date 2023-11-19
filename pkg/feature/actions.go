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
	"encoding/json"
	"strings"
)

// Actions are the collection of one or more user Action permissions
type Actions []Action

func ParseActions(lines ...string) (actions Actions) {
	for _, newlines := range lines {
		for _, line := range strings.Split(newlines, "\n") {
			for _, part := range strings.Split(line, " ") {
				if part = strings.TrimSpace(part); part != "" {
					actions = append(actions, ParseAction(part))
				}
			}
		}
	}
	return
}

func (a Actions) Len() int {
	return len(a)
}

func (a Actions) String() (s string) {
	for idx, action := range a {
		if idx > 0 {
			s += " "
		}
		s += action.String()
	}
	return
}

func (a Actions) Has(action Action) (present bool) {
	for _, aa := range a {
		if present = aa == action; present {
			return
		}
	}
	return
}

func (a Actions) HasOneOf(actions Actions) (present bool) {
	for _, action := range actions {
		if present = a.Has(action); present {
			return
		}
	}
	return
}

func (a Actions) HasAllOf(actions Actions) (present bool) {
	for _, action := range actions {
		if present = a.Has(action); !present {
			return
		}
	}
	return
}

func (a Actions) HasVerb(verb string) (present bool) {
	for _, aa := range a {
		if present = aa.Verb() == verb; present {
			return
		}
	}
	return
}

func (a Actions) HasSubject(subject string) (present bool) {
	for _, aa := range a {
		if present = aa.Subject() == subject; present {
			return
		}
	}
	return
}

func (a Actions) Append(actions ...Action) (modified Actions) {
	modified = a
	for _, action := range actions {
		if !modified.Has(action) {
			modified = append(modified, action)
		}
	}
	return
}

func (a Actions) AsNewlines() (newlines string) {
	for idx, action := range a {
		if idx > 0 {
			newlines += "\n"
		}
		newlines += action.String()
	}
	return
}

func (a Actions) FilterKnown(other Actions) (known Actions) {
	for _, action := range other {
		if a.Has(action) {
			known = append(known, action)
		}
	}
	return
}

func (a Actions) FilterUnknown(other Actions) (unknown Actions) {
	for _, action := range other {
		if !a.Has(action) {
			unknown = append(unknown, action)
		}
	}
	return
}

func (a Actions) Bytes() (data []byte) {
	var err error
	if data, err = json.MarshalIndent(a, "", "\t"); err != nil {
		panic(err)
	}
	return
}