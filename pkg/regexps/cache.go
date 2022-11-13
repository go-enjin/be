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

package regexps

import (
	"regexp"
	"sync"
)

var _cache *cache

func init() {
	_cache = &cache{
		data: make(map[string]*cacheEntry),
	}
}

type cache struct {
	data map[string]*cacheEntry

	sync.RWMutex
}

type cacheEntry struct {
	Regexp *regexp.Regexp
	Source string
	Error  error
}

func (c *cache) get(expr string) (rx *regexp.Regexp, err error, ok bool) {
	c.RLock()
	defer c.RUnlock()
	var entry *cacheEntry
	if entry, ok = c.data[expr]; ok {
		rx = entry.Regexp
		err = entry.Error
	}
	return
}

func (c *cache) set(expr string, rx *regexp.Regexp, err error) {
	c.Lock()
	defer c.Unlock()
	c.data[expr] = &cacheEntry{
		Regexp: rx,
		Source: expr,
		Error:  err,
	}
}

func Compile(expr string) (rx *regexp.Regexp, err error) {
	var ok bool
	if rx, err, ok = _cache.get(expr); ok {
		return
	}
	rx, err = regexp.Compile(expr)
	_cache.set(expr, rx, err)
	return
}