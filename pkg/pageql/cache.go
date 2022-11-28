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
	Expr  *Statement
}

func (c *cache) get(query string) (stmnt *Statement, ok bool) {
	c.RLock()
	defer c.RUnlock()
	var entry *cacheEntry
	if entry, ok = c.data[query]; ok {
		stmnt = entry.Expr
		return
	}
	return
}

func (c *cache) set(query string, expr *Statement) {
	c.Lock()
	defer c.Unlock()
	c.data[query] = &cacheEntry{
		Query: query,
		Expr:  expr,
	}
	return
}

func Compile(query string) (stmnt *Statement, err *ParseError) {
	query = SanitizeQuery(query)
	var ok bool
	if stmnt, ok = _cache.get(query); ok {
		return
	}
	if stmnt, err = parseString(query); err == nil {
		_cache.set(query, stmnt)
	}
	return
}