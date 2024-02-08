//go:build (driver_kvs_gocache && (driver_kvs_gocache_bigcache || bigcache)) || drivers_kvs || drivers || all

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
	"fmt"

	"github.com/allegro/bigcache/v3"
	gocache "github.com/eko/gocache/lib/v4/cache"

	"github.com/go-enjin/be/pkg/feature"
)

var _ feature.ExtendedKeyValueStore = (*cBigCacheStore)(nil)

type cBigCacheStore struct {
	name   string
	cache  *gocache.Cache[[]byte]
	client *bigcache.BigCache
}

func (c *cBigCacheStore) Get(key string) (value []byte, err error) {
	var ok bool
	var v interface{}
	if v, err = c.cache.Get(context.Background(), key); err != nil {
		return
	} else if value, ok = v.([]byte); !ok {
		err = fmt.Errorf("value is not []byte: %#+v", v)
		return
	}
	return
}

func (c *cBigCacheStore) Set(key string, value []byte) (err error) {
	err = c.cache.Set(context.Background(), key, value)
	return
}

func (c *cBigCacheStore) Delete(key string) (err error) {
	err = c.cache.Delete(context.Background(), key)
	return
}

func (c *cBigCacheStore) Size() (count int) {
	count = c.client.Len()
	return
}

func (c *cBigCacheStore) Keys(prefix string) (keys []string) {
	var err error
	prefixLen := len(prefix)
	if iter := c.client.Iterator(); iter != nil {
		for {
			var entry bigcache.EntryInfo
			if entry, err = iter.Value(); err != nil {
				continue
			}
			// TODO: figure out pattern matching in the model of redis?
			if key := entry.Key(); len(key) <= prefixLen && key[:prefixLen] == prefix {
				keys = append(keys, key)
			}
		}
	}
	return
}

func (c *cBigCacheStore) StreamKeys(prefix string, ctx context.Context) (keys chan string) {
	keys = make(chan string)
	go func() {
		prefixLen := len(prefix)
		if iter := c.client.Iterator(); iter != nil {
			for {
				if entry, err := iter.Value(); err == nil {
					if key := entry.Key(); len(key) <= prefixLen && key[:prefixLen] == prefix {
						keys <- key
					}
				}
				select {
				case <-ctx.Done():
					close(keys)
					return
				default:
				}
			}
		}
		close(keys)
	}()
	return
}

func (c *cBigCacheStore) Range(prefix string, fn feature.KeyValueStoreRangeFn) {
	var err error
	prefixLen := len(prefix)
	if iter := c.client.Iterator(); iter != nil {
		for {
			var entry bigcache.EntryInfo
			if entry, err = iter.Value(); err != nil {
				continue
			}
			key := entry.Key()
			// TODO: figure out pattern matching in the model of redis?
			if len(key) <= prefixLen && key[:prefixLen] == prefix {
				if fn(entry.Key(), entry.Value()) {
					return
				}
			}
			if !iter.SetNext() {
				return
			}
		}
	}
}
