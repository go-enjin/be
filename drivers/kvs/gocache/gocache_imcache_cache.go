//go:build (driver_kvs_gocache && (driver_kvs_gocache_imcache || imcache)) || drivers_kvs || drivers || all

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
	"sync"
	"time"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
)

var _ feature.KeyValueCache = (*cIMCacheCache)(nil)

type cIMCacheCache struct {
	buckets    map[string]*cIMCacheStore
	expiration time.Duration
	interval   time.Duration
	sync.RWMutex
}

func newIMCacheCache(expiration, interval time.Duration) (cache *cIMCacheCache) {
	cache = &cIMCacheCache{
		expiration: expiration,
		interval:   interval,
		buckets:    make(map[string]*cIMCacheStore),
	}
	return
}

func (c *cIMCacheCache) ListBuckets() (names []string) {
	names = maps.SortedKeys(c.buckets)
	return
}

func (c *cIMCacheCache) MustBucket(name string) (kvs feature.KeyValueStore) {
	if v, err := c.Bucket(name); err != nil {
		log.ErrorDF(1, "error getting required bucket \"%v\": - %v", name, err)
		panic(err)
	} else {
		kvs = v
	}
	return
}

func (c *cIMCacheCache) Bucket(name string) (kvs feature.KeyValueStore, err error) {
	if v, e := c.GetBucket(name); e == nil {
		kvs = v
		return
	}
	kvs, err = c.AddBucket(name)
	return
}

func (c *cIMCacheCache) AddBucket(name string) (kvs feature.KeyValueStore, err error) {
	c.Lock()
	defer c.Unlock()
	if _, exists := c.buckets[name]; exists {
		err = BucketExists
		return
	}
	c.buckets[name] = newIMCacheBucket(c.expiration, c.interval)
	kvs = c.buckets[name]
	return
}

func (c *cIMCacheCache) GetBucket(name string) (kvs feature.KeyValueStore, err error) {
	c.RLock()
	defer c.RUnlock()
	if v, ok := c.buckets[name]; ok {
		kvs = v
	} else {
		err = BucketNotFound
	}
	return
}

func (c *cIMCacheCache) GetBucketSource(name string) (src interface{}) {
	c.RLock()
	defer c.RUnlock()
	if v, ok := c.buckets[name]; ok {
		src = v.cache
	}
	return
}
