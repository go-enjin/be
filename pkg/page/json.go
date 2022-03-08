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

package page

import (
	"encoding/json"
	"regexp"

	"github.com/go-enjin/be/pkg/context"
)

var (
	rxPageJson = regexp.MustCompile(`(?ms)\A\s*(\{.+?\})\s*^`)
)

func (p *Page) parseJson(raw string) bool {
	if rxPageJson.MatchString(raw) {
		m := rxPageJson.FindStringSubmatch(raw)
		var err error
		var ctx context.Context
		if ctx, err = ParseJson(m[1]); err != nil {
			return false
		}
		p.parseContext(ctx)
		p.FrontMatter = m[1]
		p.Content = rxPageJson.ReplaceAllString(raw, "")
		return true
	}
	return false
}

func ParseJson(content string) (m context.Context, err error) {
	m = context.New()
	err = json.Unmarshal([]byte(content), &m)
	return
}