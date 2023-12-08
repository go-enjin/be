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
	"sync"
	"time"

	"github.com/allegro/bigcache/v3"
	gocache "github.com/eko/gocache/lib/v4/cache"
	store_go_cache "github.com/eko/gocache/store/bigcache/v4"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
)

var _ feature.KeyValueCache = (*cBigCache)(nil)

type cBigCache struct {
	buckets     map[string]*cBigCacheStore
	lifeWindow  time.Duration
	cleanWindow time.Duration

	sync.RWMutex
}

func newBigCache(lifeWindow, cleanWindow time.Duration) (cache *cBigCache) {
	cache = &cBigCache{
		lifeWindow:  lifeWindow,
		cleanWindow: cleanWindow,
		buckets:     make(map[string]*cBigCacheStore),
	}
	return
}

func (c *cBigCache) ListBuckets() (names []string) {
	names = maps.SortedKeys(c.buckets)
	return
}

func (c *cBigCache) MustBucket(name string) (kvs feature.KeyValueStore) {
	if v, err := c.Bucket(name); err != nil {
		log.ErrorDF(1, "error getting required bucket \"%v\": - %v", name, err)
		panic(err)
	} else {
		kvs = v
	}
	return
}

func (c *cBigCache) Bucket(name string) (kvs feature.KeyValueStore, err error) {
	if v, e := c.GetBucket(name); e == nil {
		kvs = v
		return
	}
	kvs, err = c.AddBucket(name)
	return
}

func (c *cBigCache) AddBucket(name string) (kvs feature.KeyValueStore, err error) {
	c.Lock()
	defer c.Unlock()
	if v, exists := c.buckets[name]; exists {
		kvs = v
		//err = BucketExists
		return
	}
	dcfg := bigcache.DefaultConfig(-1)
	cfg := bigcache.Config{
		Shards:             dcfg.Shards,
		LifeWindow:         -1,
		CleanWindow:        -1,
		MaxEntriesInWindow: dcfg.MaxEntriesInWindow,
		MaxEntrySize:       dcfg.MaxEntrySize,
		StatsEnabled:       dcfg.StatsEnabled,
		Verbose:            dcfg.Verbose,
		Hasher:             dcfg.Hasher,
		HardMaxCacheSize:   dcfg.HardMaxCacheSize,
		Logger:             log.PrefixedLogger("bigcache"),
	}
	gocacheClient, _ := bigcache.New(context.Background(), cfg)
	gocacheStore := store_go_cache.NewBigcache(gocacheClient)
	cacheManager := gocache.New[[]byte](gocacheStore)
	c.buckets[name] = &cBigCacheStore{
		cache: cacheManager,
	}
	kvs = c.buckets[name]
	return
}

func (c *cBigCache) GetBucket(name string) (kvs feature.KeyValueStore, err error) {
	c.RLock()
	defer c.RUnlock()
	if v, ok := c.buckets[name]; ok {
		kvs = v
	} else {
		err = BucketNotFound
	}
	return
}

func (c *cBigCache) GetBucketSource(name string) (src interface{}) {
	c.RLock()
	defer c.RUnlock()
	if v, ok := c.buckets[name]; ok {
		src = v.cache
	}
	return
}