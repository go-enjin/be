//go:build (driver_kvs_gocache && (driver_kvs_gocache_memshard || memshard)) || drivers_kvs || drivers || all

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
	"hash/maphash"
	"os"
	"sync"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
)

var (
	MemShardShardCount int = 100
)

type MemShardSupport interface {
	AddMemShardCache(name string, buckets ...string) MakeFeature
}

func (f *CFeature) AddMemShardCache(name string, buckets ...string) MakeFeature {
	f.Lock()
	defer f.Unlock()
	f.caches[name] = newMemShardCache()
	for _, bucket := range buckets {
		if _, err := f.caches[name].AddBucket(bucket); err != nil {
			log.FatalDF(1, "error adding bucket to cache: %v - %v", name, bucket)
		}
	}
	return f
}

var _ feature.KeyValueCache = (*cMemShardCache)(nil)

type cMemShardCache struct {
	buckets map[string]*cMemShardStore
	sync.RWMutex
}

func newMemShardCache() (cache *cMemShardCache) {
	cache = &cMemShardCache{
		buckets: make(map[string]*cMemShardStore),
	}
	return
}

func (c *cMemShardCache) ListBuckets() (names []string) {
	names = maps.SortedKeys(c.buckets)
	return
}

func (c *cMemShardCache) MustBucket(name string) (kvs feature.KeyValueStore) {
	if v, err := c.Bucket(name); err != nil {
		log.ErrorDF(1, "error getting required bucket \"%v\": - %v", name, err)
		panic(err)
	} else {
		kvs = v
	}
	return
}

func (c *cMemShardCache) Bucket(name string) (kvs feature.KeyValueStore, err error) {
	if v, e := c.GetBucket(name); e == nil {
		kvs = v
		return
	}
	kvs, err = c.AddBucket(name)
	return
}

func (c *cMemShardCache) AddBucket(name string) (kvs feature.KeyValueStore, err error) {
	c.Lock()
	defer c.Unlock()
	if _, exists := c.buckets[name]; exists {
		err = BucketExists
		return
	}
	c.buckets[name] = &cMemShardStore{
		shards: makeMemShardShard(MemShardShardCount),
	}
	kvs = c.buckets[name]
	return
}

func (c *cMemShardCache) GetBucket(name string) (kvs feature.KeyValueStore, err error) {
	c.RLock()
	defer c.RUnlock()
	if v, ok := c.buckets[name]; ok {
		kvs = v
	} else {
		err = BucketNotFound
	}
	return
}

func (c *cMemShardCache) GetBucketSource(name string) (src interface{}) {
	return
}

var _ feature.KeyValueStore = (*cMemShardStore)(nil)

type cMemShardStore struct {
	//cache map[string][]byte
	shards cMemShardStoreCaches
	sync.RWMutex
}

func (f *cMemShardStore) Get(k interface{}) (value interface{}, err error) {
	f.RLock()
	defer f.RUnlock()
	var ok bool
	var key string
	if key, ok = k.(string); !ok {
		err = fmt.Errorf("not a string key: %#+v", k)
		return
	}

	//var data []byte

	if value, ok = f.shards.Get(key); !ok {
		err = os.ErrNotExist
		return
	}

	//if data, ok = f.cache[key]; !ok {
	//	err = os.ErrNotExist
	//	return
	//}

	//value, err = gob.Decode(data)
	return
}

func (f *cMemShardStore) Set(k interface{}, value interface{}) (err error) {
	f.Lock()
	defer f.Unlock()
	var ok bool
	var key string
	if key, ok = k.(string); !ok {
		err = fmt.Errorf("not a string key: %#+v", k)
		return
	}
	//var data []byte
	//if data, err = gob.Encode(value); err != nil {
	//	return
	//}

	f.shards.Set(key, value)

	//f.cache[key] = data
	return
}

func (f *cMemShardStore) Delete(k interface{}) (err error) {
	f.Lock()
	defer f.Unlock()
	var ok bool
	var key string
	if key, ok = k.(string); !ok {
		err = fmt.Errorf("not a string key: %#+v", k)
		return
	}
	err = f.shards.Delete(key)
	return
}

type cMemShardStoreCache[V interface{}] struct {
	m map[string]V
	sync.RWMutex
}

type cMemShardStoreValue struct {
	v interface{}
}

type cMemShardStoreCaches []*cMemShardStoreCache[*cMemShardStoreValue]

func makeMemShardShard(size int) cMemShardStoreCaches {
	m := make([]*cMemShardStoreCache[*cMemShardStoreValue], size)
	for i := 0; i < size; i++ {
		m[i] = &cMemShardStoreCache[*cMemShardStoreValue]{
			m: make(map[string]*cMemShardStoreValue),
		}
	}
	return m
}

func (m cMemShardStoreCaches) getShardKey(key string) int {
	/* sha256 */
	//hash := sha256.Sum256([]byte(key))
	//return int(hash[31]) % len(m)

	/* sha1 */
	//hash := sha1.Sum([]byte(key))
	//return int(hash[19]) % len(m)

	/* fnv */
	//hasher := fnv.New64a()
	//_, _ = hasher.Write([]byte(key))
	//hash := hasher.Sum64()
	//return int(hash) % len(m)

	hasher := maphash.Hash{}
	return int(hasher.Sum([]byte(key))[0]) % len(m)
}

func (m cMemShardStoreCaches) GetShard(key string) *cMemShardStoreCache[*cMemShardStoreValue] {
	sk := m.getShardKey(key)
	return m[sk]
}

func (m cMemShardStoreCaches) Get(key string) (value interface{}, ok bool) {
	shard := m.GetShard(key)
	shard.RLock()
	defer shard.RUnlock()
	var v *cMemShardStoreValue
	if v, ok = shard.m[key]; ok {
		return v.v, ok
	}
	return
}

func (m cMemShardStoreCaches) Set(key string, val interface{}) {
	shard := m.GetShard(key)
	shard.Lock()
	defer shard.Unlock()
	shard.m[key] = &cMemShardStoreValue{v: val}
}

func (m cMemShardStoreCaches) Delete(key string) (err error) {
	shard := m.GetShard(key)
	shard.Lock()
	defer shard.Unlock()
	if _, ok := shard.m[key]; ok {
		delete(shard.m, key)
	}
	return
}