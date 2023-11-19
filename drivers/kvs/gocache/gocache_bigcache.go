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
	"sync"
	"time"

	"github.com/allegro/bigcache/v3"
	gocache "github.com/eko/gocache/lib/v4/cache"
	store_go_cache "github.com/eko/gocache/store/bigcache/v4"
	"github.com/patrickmn/go-cache"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/gob"
	"github.com/go-enjin/be/pkg/maps"

	"github.com/go-enjin/be/pkg/log"
)

type BigCacheSupport interface {
	AddBigCache(name string, buckets ...string) MakeFeature
	AddExpiringBigCache(name string, lifeWindow, cleanWindow time.Duration, buckets ...string) MakeFeature
}

func (f *CFeature) AddBigCache(name string, buckets ...string) MakeFeature {
	return f.AddExpiringBigCache(name, cache.NoExpiration, cache.NoExpiration, buckets...)
}

func (f *CFeature) AddExpiringBigCache(name string, lifeWindow, cleanWindow time.Duration, buckets ...string) MakeFeature {
	f.Lock()
	defer f.Unlock()
	f.caches[name] = newBigCache(lifeWindow, cleanWindow)
	for _, bucket := range buckets {
		if _, err := f.caches[name].AddBucket(bucket); err != nil {
			log.FatalDF(1, "error adding bucket to cache: %v - %v", name, bucket)
		}
	}
	return f
}

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
	cacheManager := gocache.New[interface{}](gocacheStore)
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

var _ feature.KeyValueStore = (*cBigCacheStore)(nil)

type cBigCacheStore struct {
	cache *gocache.Cache[interface{}]
	sync.RWMutex
}

func (f *cBigCacheStore) Get(key interface{}) (value interface{}, err error) {
	//f.RLock()
	//defer f.RUnlock()
	var ok bool
	var v interface{}
	var data []byte
	if v, err = f.cache.Get(context.Background(), key); err != nil {
		return
	} else if data, ok = v.([]byte); !ok {
		err = fmt.Errorf("value is not []byte: %#+v", v)
		return
	}
	value, err = gob.Decode(data)
	return
}

func (f *cBigCacheStore) Set(key interface{}, value interface{}) (err error) {
	//f.Lock()
	//defer f.Unlock()
	var data []byte
	if data, err = gob.Encode(value); err != nil {
		return
	}
	err = f.cache.Set(context.Background(), key, data)
	return
}

func (f *cBigCacheStore) Delete(key interface{}) (err error) {
	err = f.cache.Delete(context.Background(), key)
	return
}