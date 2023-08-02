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
	"strings"
)

func IsEmptyString(input interface{}) (empty bool) {
	empty = strings.TrimSpace(ToString(input)) == ""
	return
}

func ToString(input interface{}) (output string) {
	switch t := input.(type) {
	case string:
		output = t
	case template.HTML:
		output = string(t)
	case template.CSS:
		output = string(t)
	case template.JS:
		output = string(t)
	case template.HTMLAttr:
		output = string(t)
	case template.JSStr:
		output = string(t)
	case template.URL:
		output = string(t)
	case template.Srcset:
		output = string(t)
	default:
		output = fmt.Sprintf("%v", input)
	}
	return
}