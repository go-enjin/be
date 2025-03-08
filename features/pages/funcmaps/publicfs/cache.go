//go:build page_funcmaps || pages || all

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

package publicfs

import (
	"sync"

	clContext "github.com/go-corelibs/context"
	"github.com/go-corelibs/maps"
)

type pfsCacheItem struct {
	url      string
	shasum   string
	mimeType string
}

func (i *pfsCacheItem) revUrl() (rev string) {
	return i.url + "?rev=" + i.shasum
}

type pfsCache struct {
	d map[string]*pfsCacheItem
	m *sync.RWMutex
}

func makePfsCache() (c *pfsCache) {
	return &pfsCache{
		d: make(map[string]*pfsCacheItem),
		m: &sync.RWMutex{},
	}
}

func (c *pfsCache) get(url string) *pfsCacheItem {
	c.m.RLock()
	defer c.m.RUnlock()
	if item, ok := c.d[url]; ok {
		return &pfsCacheItem{
			url:      item.url,
			shasum:   item.shasum,
			mimeType: item.mimeType,
		}
	}
	return nil
}

func (c *pfsCache) set(url, shasum, mimeType string) {
	c.m.Lock()
	defer c.m.Unlock()
	c.d[url] = &pfsCacheItem{
		url:      url,
		shasum:   shasum,
		mimeType: mimeType,
	}
}

func (c *pfsCache) walk(fn func(item *pfsCacheItem) bool) {
	c.m.RLock()
	defer c.m.RUnlock()
	for _, key := range maps.SortedKeys(c.d) {
		if !fn(c.d[key]) {
			return
		}
	}
}

func getPfsCache(ctx clContext.Context, key string) *pfsCache {
	if cache, ok := ctx[key].(*pfsCache); ok && cache != nil {
		return cache
	}
	ctx[key] = makePfsCache()
	return ctx[key].(*pfsCache)
}
