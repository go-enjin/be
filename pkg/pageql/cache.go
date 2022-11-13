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

package pageql

import "sync"

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
	Query string
	Expr  *Expression
	Error error
}

func (c *cache) get(query string) (expr *Expression, err error, ok bool) {
	c.RLock()
	defer c.RUnlock()
	var entry *cacheEntry
	if entry, ok = c.data[query]; ok {
		expr = entry.Expr
		err = entry.Error
		return
	}
	return
}

func (c *cache) set(query string, expr *Expression, err error) {
	c.Lock()
	defer c.Unlock()
	c.data[query] = &cacheEntry{
		Query: query,
		Expr:  expr,
		Error: err,
	}
	return
}

func Compile(query string) (expr *Expression, err error) {
	var ok bool
	if expr, err, ok = _cache.get(query); ok {
		return
	}
	expr, err = parser.ParseString("pageql", query)
	return
}