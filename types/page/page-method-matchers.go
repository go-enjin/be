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

package page

import (
	"strings"

	"github.com/go-corelibs/slices"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/pageql"
)

func (p *CPage) Match(path string) (found string, ok bool) {
	if ok = p.fields.Url == path; ok {
		found = p.fields.Url
	} else if ok = p.IsTranslation(path); ok {
		found = p.fields.Translates
	} else if ok = p.IsRedirection(path); ok {
		found = p.fields.Url
	} else {
		for _, matcher := range feature.GetPageMatcherFuncs() {
			if found, ok = matcher(path, p); ok {
				return
			}
		}
	}
	return
}
func (p *CPage) MatchPrefix(prefix string) (found string, ok bool) {
	if ok = strings.HasPrefix(p.fields.Url, prefix); ok {
		found = p.fields.Url
	}
	return
}

func (p *CPage) Redirections() (redirects []string) {
	if redirect := p.fields.Context.Get("Redirect"); redirect != nil {
		switch t := redirect.(type) {
		case string:
			redirects = append(redirects, t)
		case []interface{}:
			for _, v := range t {
				if r, ok := v.(string); ok {
					redirects = append(redirects, r)
				}
			}
		}
	}
	return
}

func (p *CPage) IsRedirection(path string) (ok bool) {
	ok = slices.Present(path, p.Redirections()...)
	return
}

func (p *CPage) IsTranslation(path string) (ok bool) {
	ok = p.fields.Translates == path
	return
}

func (p *CPage) HasTranslation() (ok bool) {
	ok = p.fields.Translates != ""
	return
}

func (p *CPage) MatchQL(query string) (ok bool, err error) {
	ok, err = pageql.Match(query, p.fields.Context.Copy())
	return
}
