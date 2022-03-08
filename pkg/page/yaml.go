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

	"gopkg.in/yaml.v3"

	"github.com/go-enjin/be/pkg/context"
)

var (
	rxPageYaml = regexp.MustCompile(`(?ms)\A\s*---\s*^(.+?)\s*---\s*^`)
)

func (p *Page) parseYaml(raw string) bool {
	if rxPageYaml.MatchString(raw) {
		m := rxPageYaml.FindStringSubmatch(raw)
		var err error
		var ctx context.Context
		if ctx, err = ParseYaml(m[1]); err != nil {
			return false
		}
		p.parseContext(ctx)
		p.FrontMatter = m[1]
		p.Content = rxPageYaml.ReplaceAllString(raw, "")
		return true
	}
	return false
}

func ParseYaml(content string) (m context.Context, err error) {
	m = context.New()
	err = yaml.Unmarshal([]byte(content), m)
	return
}