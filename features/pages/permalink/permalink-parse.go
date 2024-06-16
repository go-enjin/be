// Copyright (c) 2024  The Go-Enjin Authors
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

package permalink

import (
	"strings"

	"github.com/go-corelibs/enjinql"
	"github.com/go-corelibs/rxp"
)

var (
	rxPermalinkRoot = rxp.Pattern{
		rxp.Caret(),
		rxp.Text("/"),
		rxp.Or("c",
			rxp.IsUUID(),
			rxp.IsHash10(),
		),
		rxp.Text("/", "??"),
		rxp.Dollar(),
	}

	rxPermalinkedSlug = rxp.Pattern{
		rxp.Text("-"),
		rxp.IsHash10("c"),
		rxp.Text("/", "??"),
		rxp.Dollar(),
	}
)

func ParsePermalink(path string) (id string, short bool, ok bool) {
	switch {
	case rxPermalinkRoot.MatchString(path):
		m := rxPermalinkRoot.FindAllStringSubmatch(path, 1)
		id = m[0][1]
	case rxPermalinkedSlug.MatchString(path):
		m := rxPermalinkedSlug.FindAllStringSubmatch(path, 1)
		id = m[0][1]
	}

	size := len(id)
	if ok = size == enjinql.ShortPermalinkSize || size == enjinql.LongPermalinkSize; ok {
		short = len(id) == enjinql.ShortPermalinkSize
		id = strings.ToLower(id)
	}
	return
}
