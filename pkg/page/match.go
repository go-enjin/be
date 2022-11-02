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

import "sync"

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

func (p *Page) IsTranslation(path string) (ok bool) {
	ok = p.Translates == path
	return
}

func (p *Page) HasTranslation() (ok bool) {
	ok = p.Translates != ""
	return
}