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
	"regexp"

	"github.com/BurntSushi/toml"

	"github.com/go-enjin/be/pkg/context"
)

var (
	rxPageToml = regexp.MustCompile(`(?ms)\A\s*\+\+\+\s*^(.+?)\s*\+\+\+\s*^`)
)

func (p *Page) parseToml(raw string) bool {
	if rxPageToml.MatchString(raw) {
		m := rxPageToml.FindStringSubmatch(raw)
		var err error
		var ctx context.Context
		if ctx, err = ParseToml(m[1]); err != nil {
			return false
		}
		p.parseContext(ctx)
		p.FrontMatter = m[1]
		p.Content = rxPageToml.ReplaceAllString(raw, "")
		return true
	}
	return false
}

func ParseToml(content string) (m context.Context, err error) {
	m = context.New()
	_, err = toml.Decode(content, &m)
	return
}