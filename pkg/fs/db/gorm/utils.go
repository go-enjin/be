// Copyright (c) 2023  The Go-Enjin Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this File except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package embed

import "strings"

func sqlEscapeLIKE(input string) (escaped string) {
	escaped = strings.ReplaceAll(input, `%`, `\%`)
	escaped = strings.ReplaceAll(escaped, `_`, `\_`)
	return
}

func isDirectChild(parent, path string) (is bool) {
	pLen := len(parent)
	if pLen > len(path) {
		return
	}
	var trimmed string
	if trimmed = path[0 : pLen-1]; len(trimmed) > 0 && trimmed[0] == '/' {
		// drop parent prefix and remove root slash if present
		trimmed = trimmed[1:]
	}
	is = !strings.Contains(trimmed, "/")
	return
}