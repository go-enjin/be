//go:build (driver_kvs_gocache && (driver_kvs_gocache_ristretto || ristretto)) || drivers_kvs || drivers || all

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

	"github.com/dgraph-io/ristretto"

	"github.com/go-enjin/be/pkg/gob"
	"github.com/go-enjin/be/pkg/kvs"
	"github.com/go-enjin/be/pkg/log"
)

type RistrettoSupport interface {
	AddRistrettoCache(name string, buckets ...string) MakeFeature
}

func (f *CFeature) AddRistrettoCache(name string, buckets ...string) MakeFeature {
	f.Lock()
	defer f.Unlock()
	f.caches[name] = newRistrettoCache()
	for _, bucket := range buckets {
		if _, err := f.caches[name].AddBucket(bucket); err != nil {
			log.FatalDF(1, "error adding bucket to cache: %v - %v", name, bucket)
		}
	}
	return f
}

var _ kvs.KeyValueCache = (*cRistrettoCache)(nil)

type cRistrettoCache struct {
	buckets map[string]*cRistrettoStore
	sync.RWMutex
}

func newRistrettoCache() (cache *cRistrettoCache) {
	cache = &cRistrettoCache{
		buckets: make(map[string]*cRistrettoStore),
	}
	return
}

func (c *cRistrettoCache) MustBucket(name string) (kvs kvs.KeyValueStore) {
	if v, err := c.Bucket(name); err != nil {
		log.FatalDF(1, "error getting required bucket \"%v\": - %v", name, err)
	} else {
		kvs = v
	}
	return
}

func (c *cRistrettoCache) Bucket(name string) (kvs kvs.KeyValueStore, err error) {
	if v, e := c.GetBucket(name); e == nil {
		kvs = v
		return
	}
	kvs, err = c.AddBucket(name)
	return
}

func (c *cRistrettoCache) AddBucket(name string) (kvs kvs.KeyValueStore, err error) {
	c.Lock()
	defer c.Unlock()
	if _, exists := c.buckets[name]; exists {
		err = BucketExists
		return
	}
	c.buckets[name] = newRistrettoBucket()
	kvs = c.buckets[name]
	return
}

func (c *cRistrettoCache) GetBucket(name string) (kvs kvs.KeyValueStore, err error) {
	c.RLock()
	defer c.RUnlock()
	if v, ok := c.buckets[name]; ok {
		kvs = v
	} else {
		err = BucketNotFound
	}
	return
}

func (c *cRistrettoCache) GetBucketSource(name string) (src interface{}) {
	return
}

var _ kvs.KeyValueStore = (*cRistrettoStore)(nil)

type cRistrettoStore struct {
	//cache map[string][]byte
	//shards cRistrettoStoreCaches

	cache *ristretto.Cache
	sync.RWMutex
}

func newRistrettoBucket() (store *cRistrettoStore) {
	var err error
	store = &cRistrettoStore{}
	if store.cache, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	}); err != nil {
		log.FatalF("error constructing ristretto cache instance: %v", err)
	}
	return
}

func (f *cRistrettoStore) Get(k interface{}) (value interface{}, err error) {
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

	//if data, ok = f.cache[key]; !ok {
	//	err = os.ErrNotExist
	//	return
	//}

	value, err = gob.Decode(data)
	return
}

func (f *cRistrettoStore) Set(k interface{}, value interface{}) (err error) {
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

	if !f.cache.Set(key, data, 0) {
		log.FatalF("ristretto set dropped for key: %v", key)
	}

	//f.cache[key] = data
	return
}