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
	"fmt"
	"html/template"
	"regexp"
)

var (
	rxIsEmpty = regexp.MustCompile(`(?msi)\A\s*\z`)
)

func IsEmptyString(input interface{}) (empty bool) {
	var content string
	switch t := input.(type) {
	case string:
		content = t
	case template.HTML:
		content = string(t)
	case template.CSS:
		content = string(t)
	case template.JS:
		content = string(t)
	case template.HTMLAttr:
		content = string(t)
	case template.JSStr:
		content = string(t)
	case template.URL:
		content = string(t)
	case template.Srcset:
		content = string(t)
	default:
		content = fmt.Sprintf("%v", input)
	}
	empty = rxIsEmpty.MatchString(content)
	return
}