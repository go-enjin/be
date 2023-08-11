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
	"strings"
	"sync"

	"github.com/go-enjin/be/pkg/pageql"
	"github.com/go-enjin/be/pkg/slices"
)

type MatcherFn func(path string, pg *Page) (found string, ok bool)

var (
	_knownMatcherFns     []MatcherFn
	_knownMatcherFnMutex = sync.RWMutex{}
)

func RegisterMatcherFn(matcher MatcherFn) {
	_knownMatcherFnMutex.Lock()
	defer _knownMatcherFnMutex.Unlock()
	_knownMatcherFns = append(_knownMatcherFns, matcher)
}

func (p *Page) Match(path string) (found string, ok bool) {
	if ok = p.Url == path; ok {
		found = p.Url
	} else if ok = p.IsTranslation(path); ok {
		found = p.Translates
	} else if ok = p.IsRedirection(path); ok {
		found = p.Url
	} else {
		_knownMatcherFnMutex.RLock()
		defer _knownMatcherFnMutex.RUnlock()
		for _, matcher := range _knownMatcherFns {
			if found, ok = matcher(path, p); ok {
				return
			}
		}
	}
	return
}

func (p *Page) MatchPrefix(prefix string) (found string, ok bool) {
	if ok = strings.HasPrefix(p.Url, prefix); ok {
		found = p.Url
	}
	return
}

func (p *Page) Redirections() (redirects []string) {
	if redirect := p.Context.Get("Redirect"); redirect != nil {
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

func (p *Page) IsRedirection(path string) (ok bool) {
	ok = slices.Present(path, p.Redirections()...)
	return
}

func (p *Page) IsTranslation(path string) (ok bool) {
	ok = p.Translates == path
	return
}

func (p *Page) HasTranslation() (ok bool) {
	ok = p.Translates != ""
	return
}

func (p *Page) MatchQL(query string) (ok bool, err error) {
	ok, err = pageql.Match(query, p.Context.Copy())
	return
}