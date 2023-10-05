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
	"sync"

	"github.com/patrickmn/go-cache"

	gocache "github.com/eko/gocache/lib/v4/cache"
	store_go_cache "github.com/eko/gocache/store/go_cache/v4"

	"github.com/go-enjin/be/pkg/kvs"
	"github.com/go-enjin/be/pkg/log"
)

type MemorySupport interface {
	AddMemoryCache(name string, buckets ...string) MakeFeature
}

func (f *CFeature) AddMemoryCache(name string, buckets ...string) MakeFeature {
	f.Lock()
	defer f.Unlock()
	f.caches[name] = newLocalCache()
	for _, bucket := range buckets {
		if _, err := f.caches[name].AddBucket(bucket); err != nil {
			log.FatalDF(1, "error adding bucket to cache: %v - %v", name, bucket)
		}
	}
	return f
}

var _ kvs.KeyValueCache = (*cLocalCache)(nil)

type cLocalCache struct {
	buckets map[string]*cLocalStore
	sync.RWMutex
}

func newLocalCache() (cache *cLocalCache) {
	cache = &cLocalCache{
		buckets: make(map[string]*cLocalStore),
	}
	return
}

func (c *cLocalCache) MustBucket(name string) (kvs kvs.KeyValueStore) {
	if v, err := c.Bucket(name); err != nil {
		log.FatalDF(1, "error getting required bucket \"%v\": - %v", name, err)
	} else {
		kvs = v
	}
	return
}

func (c *cLocalCache) Bucket(name string) (kvs kvs.KeyValueStore, err error) {
	if v, e := c.GetBucket(name); e == nil {
		kvs = v
		return
	}
	kvs, err = c.AddBucket(name)
	return
}

func (c *cLocalCache) AddBucket(name string) (kvs kvs.KeyValueStore, err error) {
	c.Lock()
	defer c.Unlock()
	if _, exists := c.buckets[name]; exists {
		err = BucketExists
		return
	}
	gocacheClient := cache.New(cache.NoExpiration, cache.NoExpiration)
	gocacheStore := store_go_cache.NewGoCache(gocacheClient)
	cacheManager := gocache.New[interface{}](gocacheStore)
	c.buckets[name] = &cLocalStore{
		cache: cacheManager,
	}
	kvs = c.buckets[name]
	return
}

func (c *cLocalCache) GetBucket(name string) (kvs kvs.KeyValueStore, err error) {
	c.RLock()
	defer c.RUnlock()
	if v, ok := c.buckets[name]; ok {
		kvs = v
	} else {
		err = BucketNotFound
	}
	return
}

func (c *cLocalCache) GetBucketSource(name string) (src interface{}) {
	c.RLock()
	defer c.RUnlock()
	if v, ok := c.buckets[name]; ok {
		src = v.cache
	}
	return
}

var _ kvs.KeyValueStore = (*cLocalStore)(nil)

type cLocalStore struct {
	cache *gocache.Cache[interface{}]
	sync.RWMutex
}

func (f *cLocalStore) Get(key interface{}) (value interface{}, err error) {
	//f.RLock()
	//defer f.RUnlock()
	value, err = f.cache.Get(context.Background(), key)
	// var data []byte
	// if data, err = f.cache.Get(context.Background(), key); err != nil {
	// 	return
	// }
	// value, err = gob.Decode(data)
	return
}

func (f *cLocalStore) Set(key interface{}, value interface{}) (err error) {
	//f.Lock()
	//defer f.Unlock()
	// var data []byte
	// if data, err = gob.Encode(value); err != nil {
	// 	return
	// }
	err = f.cache.Set(context.Background(), key, value)
	return
}

func (f *cLocalStore) Delete(key interface{}) (err error) {
	//f.Lock()
	//defer f.Unlock()
	// var data []byte
	// if data, err = gob.Encode(value); err != nil {
	// 	return
	// }
	err = f.cache.Delete(context.Background(), key)
	return
}