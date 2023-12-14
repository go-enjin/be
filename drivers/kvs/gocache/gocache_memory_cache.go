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
	"sync"
	"time"

	gocache "github.com/eko/gocache/lib/v4/cache"
	store_go_cache "github.com/eko/gocache/store/go_cache/v4"
	"github.com/patrickmn/go-cache"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
)

var _ feature.KeyValueCache = (*cMemoryCache)(nil)

type cMemoryCache struct {
	buckets    map[string]*cMemoryStore
	interval   time.Duration
	expiration time.Duration
	sync.RWMutex
}

func newMemoryCache(expiration, interval time.Duration) (cache *cMemoryCache) {
	cache = &cMemoryCache{
		buckets:    make(map[string]*cMemoryStore),
		interval:   interval,
		expiration: expiration,
	}
	return
}

func (c *cMemoryCache) ListBuckets() (names []string) {
	names = maps.SortedKeys(c.buckets)
	return
}

func (c *cMemoryCache) MustBucket(name string) (kvs feature.KeyValueStore) {
	if v, err := c.Bucket(name); err != nil {
		log.ErrorDF(1, "error getting required bucket \"%v\": - %v", name, err)
		panic(err)
	} else {
		kvs = v
	}
	return
}

func (c *cMemoryCache) Bucket(name string) (kvs feature.KeyValueStore, err error) {
	if v, e := c.GetBucket(name); e == nil {
		kvs = v
		return
	}
	kvs, err = c.AddBucket(name)
	return
}

func (c *cMemoryCache) AddBucket(name string) (kvs feature.KeyValueStore, err error) {
	c.Lock()
	defer c.Unlock()
	if _, exists := c.buckets[name]; exists {
		err = BucketExists
		return
	}
	gocacheClient := cache.New(NoExpiration, NoExpiration)
	gocacheStore := store_go_cache.NewGoCache(gocacheClient)
	cacheManager := gocache.New[[]byte](gocacheStore)
	c.buckets[name] = &cMemoryStore{
		client: gocacheClient,
		cache:  cacheManager,
	}
	kvs = c.buckets[name]
	return
}

func (c *cMemoryCache) GetBucket(name string) (kvs feature.KeyValueStore, err error) {
	c.RLock()
	defer c.RUnlock()
	if v, ok := c.buckets[name]; ok {
		kvs = v
	} else {
		err = BucketNotFound
	}
	return
}

func (c *cMemoryCache) GetBucketSource(name string) (src interface{}) {
	c.RLock()
	defer c.RUnlock()
	if v, ok := c.buckets[name]; ok {
		src = v.cache
	}
	return
}