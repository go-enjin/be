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

package forms

import (
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

func Sanitize(input string) (sanitized string) {
	p := bluemonday.UGCPolicy()
	sanitized = p.Sanitize(input)
	return
}

func StripTags(input string) (sanitized string) {
	p := bluemonday.StripTagsPolicy()
	sanitized = p.Sanitize(input)
	return
}

func SanitizeRequestPath(path string) (cleaned string) {
	cleaned = TrimQueryParams(path)
	if cleaned = strings.TrimSuffix(cleaned, "/"); cleaned == "" {
		cleaned = "/"
	} else if cleaned[0] != '/' {
		cleaned = "/" + cleaned
	}
	cleaned = Sanitize(cleaned)
	return
}