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
		queryData:  make(map[string]*queryDataEntry),
		selectData: make(map[string]*selectDataEntry),
	}
}

type cache struct {
	queryData  map[string]*queryDataEntry
	selectData map[string]*selectDataEntry

	sync.RWMutex
}

type queryDataEntry struct {
	Query string
	Expr  *Statement
}

func (c *cache) getQuery(query string) (stmnt *Statement, ok bool) {
	c.RLock()
	defer c.RUnlock()
	var entry *queryDataEntry
	if entry, ok = c.queryData[query]; ok {
		stmnt = entry.Expr
		return
	}
	return
}

func (c *cache) setQuery(query string, expr *Statement) {
	c.Lock()
	defer c.Unlock()
	c.queryData[query] = &queryDataEntry{
		Query: query,
		Expr:  expr,
	}
	return
}

func CompileQuery(query string) (stmnt *Statement, err *ParseError) {
	err = nil
	query = SanitizeQuery(query)
	var ok bool
	if stmnt, ok = _cache.getQuery(query); ok {
		return
	}
	if stmnt, err = parseQueryString(query); err == nil {
		_cache.setQuery(query, stmnt)
	}
	return
}

type selectDataEntry struct {
	Query string
	Sel   *Selection
}

func (c *cache) getSelect(query string) (stmnt *Selection, ok bool) {
	c.RLock()
	defer c.RUnlock()
	var entry *selectDataEntry
	if entry, ok = c.selectData[query]; ok {
		stmnt = entry.Sel
		return
	}
	return
}

func (c *cache) setSelect(query string, sel *Selection) {
	c.Lock()
	defer c.Unlock()
	c.selectData[query] = &selectDataEntry{
		Query: query,
		Sel:   sel,
	}
	return
}

func CompileSelect(query string) (sel *Selection, err *ParseError) {
	err = nil
	query = SanitizeQuery(query)
	var ok bool
	if sel, ok = _cache.getSelect(query); ok {
		return
	}
	if sel, err = parseSelectString(query); err == nil {
		_cache.setSelect(query, sel)
	}
	return
}