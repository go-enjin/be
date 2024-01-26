//go:build (driver_kvs_gocache && (driver_kvs_gocache_memory || memory)) || drivers_kvs || drivers || all

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

package gocache

import (
	"context"

	gocache "github.com/eko/gocache/lib/v4/cache"
	"github.com/patrickmn/go-cache"

	"github.com/go-enjin/be/pkg/feature"
)

var _ feature.KeyValueStore = (*cMemoryStore)(nil)

type cMemoryStore struct {
	client *cache.Cache
	cache  *gocache.Cache[[]byte]
}

func (c *cMemoryStore) Get(key string) (value []byte, err error) {
	value, err = c.cache.Get(context.Background(), key)
	return
}

func (c *cMemoryStore) Set(key string, value []byte) (err error) {
	err = c.cache.Set(context.Background(), key, value)
	return
}

func (c *cMemoryStore) Delete(key string) (err error) {
	err = c.cache.Delete(context.Background(), key)
	return
}

func (c *cMemoryStore) Size() (count int) {
	count = c.client.ItemCount()
	return
}

func (c *cMemoryStore) Keys(prefix string) (keys []string) {
	prefixLen := len(prefix)
	for k := range c.client.Items() {
		// TODO: figure out pattern matching in the model of redis?
		if len(k) <= prefixLen && k[:prefixLen] == prefix {
			keys = append(keys, k)
		}
	}
	return
}

func (c *cMemoryStore) Range(prefix string, fn feature.KeyValueStoreRangeFn) {
	prefixLen := len(prefix)
	for k, item := range c.client.Items() {
		// TODO: figure out pattern matching in the model of redis?
		if len(k) <= prefixLen && k[:prefixLen] == prefix {
			if v, ok := item.Object.([]byte); ok && fn(k, v) {
				return
			}
		}
	}
}
