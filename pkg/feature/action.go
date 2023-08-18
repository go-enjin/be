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

	"github.com/iancoleman/strcase"
)

// Action is a kebab-cased name consisting of a verb, a subject (feature.Tag)
// and one or more additional details, separated by periods. For example: the
// action `view_fs-content_page` has a verb of "view", a subject of "fs-content"
// and one additional detail of "page". Details are individually converted to
// kebab-case and joined with hyphens, ie: `view_fs-content_page.search`
type Action string

func ParseAction(line string) (action Action) {
	var subject, verb string
	var details []string

	if parts := strings.SplitN(line, "_", 2); len(parts) >= 2 {
		verb = parts[0]
		if more := strings.SplitN(parts[1], "_", 2); len(more) == 1 {
			subject = parts[1]
		} else {
			subject = more[0]
			details = strings.Split(more[1], ".")
		}
	}

	action = NewAction(subject, verb, details...)
	return
}

func NewAction(subject string, verb string, details ...string) Action {
	action := strcase.ToKebab(verb) + "_" + subject
	if len(details) > 0 {
		action += "_"
		for idx, detail := range details {
			if idx > 0 {
				action += "."
			}
			action += strcase.ToKebab(detail)
		}
	}
	return Action(action)
}

func (a Action) String() string {
	return string(a)
}

func (a Action) Verb() string {
	parts := strings.SplitN(string(a), "_", 2)
	return parts[0]
}

func (a Action) Subject() string {
	parts := strings.SplitN(string(a), "_", 2)
	if more := strings.SplitN(parts[1], "_", 2); len(more) > 1 {
		return more[0]
	}
	return parts[1]
}

func (a Action) Details() (details []string) {
	parts := strings.SplitN(string(a), "_", 3)
	if len(parts) == 3 {
		details = strings.Split(parts[2], ".")
		return
	}
	return
}