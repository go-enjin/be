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
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/erni27/imcache"
	"github.com/patrickmn/go-cache"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/gob"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
)

var IMCacheShardCount = 50

type IMCacheSupport interface {
	AddIMCacheCache(name string, buckets ...string) MakeFeature
}

func (f *CFeature) AddIMCacheCache(name string, buckets ...string) MakeFeature {
	return f.AddExpiringIMCacheCache(name, cache.NoExpiration, cache.NoExpiration, buckets...)
}

func (f *CFeature) AddExpiringIMCacheCache(name string, expiration, interval time.Duration, buckets ...string) MakeFeature {
	f.Lock()
	defer f.Unlock()
	f.caches[name] = newIMCacheCache(expiration, interval)
	for _, bucket := range buckets {
		if _, err := f.caches[name].AddBucket(bucket); err != nil {
			log.FatalDF(1, "error adding bucket to cache: %v - %v", name, bucket)
		}
	}
	return f
}

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
	return
}

var _ feature.KeyValueStore = (*cIMCacheStore)(nil)

type cIMCacheStore struct {
	cache *imcache.Sharded[string, []byte]

	expiration time.Duration
	interval   time.Duration

	sync.RWMutex
}

func newIMCacheBucket(expiration, interval time.Duration) (store *cIMCacheStore) {
	var options []imcache.Option[string, []byte]
	if expiration > 1 {
		options = append(options, imcache.WithDefaultExpirationOption[string, []byte](expiration))
	}
	if interval > 1 {
		options = append(options, imcache.WithCleanerOption[string, []byte](interval))
	}
	store = &cIMCacheStore{
		expiration: expiration,
		interval:   interval,
		cache: imcache.NewSharded[string, []byte](
			IMCacheShardCount,
			imcache.DefaultStringHasher64{},
			options...,
		),
	}
	return
}

func (f *cIMCacheStore) Get(k interface{}) (value interface{}, err error) {
	f.RLock()
	defer f.RUnlock()
	var ok bool
	var key string
	if key, ok = k.(string); !ok {
		err = fmt.Errorf("not a string key: %#+v", k)
		return
	}

	var data []byte
	var v interface{}

	if v, ok = f.cache.Get(key); !ok {
		err = os.ErrNotExist
		return
	} else if data, ok = v.([]byte); !ok {
		err = fmt.Errorf("invalid value type: %T for key: %v", v, key)
		return
	}

	value, err = gob.Decode(data)
	return
}

func (f *cIMCacheStore) Set(k interface{}, value interface{}) (err error) {
	f.Lock()
	defer f.Unlock()
	var ok bool
	var key string
	if key, ok = k.(string); !ok {
		err = fmt.Errorf("not a string key: %#+v", k)
		return
	}
	var data []byte
	if data, err = gob.Encode(value); err != nil {
		return
	}

	f.cache.Set(key, data, imcache.WithNoExpiration())
	return
}

func (f *cIMCacheStore) Delete(k interface{}) (err error) {
	f.Lock()
	defer f.Unlock()
	var ok bool
	var key string
	if key, ok = k.(string); !ok {
		err = fmt.Errorf("not a string key: %#+v", k)
		return
	}
	f.cache.Remove(key)
	return
}
