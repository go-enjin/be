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

package pages

import (
	"regexp"
)

var (
	permalinkPattern  = `([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}|[0-9a-f]{10})`
	rxPermalinkRoot   = regexp.MustCompile(`^/` + permalinkPattern + `/??$`)
	rxPermalinkedSlug = regexp.MustCompile(`-` + permalinkPattern + `/??$`)
)

func ParsePermalink(path string) (id string, ok bool) {
	if ok = rxPermalinkRoot.MatchString(path); ok {
		m := rxPermalinkRoot.FindAllStringSubmatch(path, 1)
		id = m[0][1]
	} else if ok = rxPermalinkedSlug.MatchString(path); ok {
		m := rxPermalinkedSlug.FindAllStringSubmatch(path, 1)
		id = m[0][1]
	}
	return
}
