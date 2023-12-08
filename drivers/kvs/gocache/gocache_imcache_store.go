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
	"time"

	"github.com/erni27/imcache"

	"github.com/go-enjin/be/pkg/feature"
)

var _ feature.KeyValueStore = (*cIMCacheStore)(nil)

type cIMCacheStore struct {
	cache *imcache.Sharded[string, []byte]

	expiration time.Duration
	interval   time.Duration
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

func (c *cIMCacheStore) Get(key string) (value []byte, err error) {
	var ok bool
	var v interface{}

	if v, ok = c.cache.Get(key); !ok {
		err = os.ErrNotExist
		return
	} else if value, ok = v.([]byte); !ok {
		err = fmt.Errorf("invalid value type: %T for key: %v", v, key)
		return
	}

	return
}

func (c *cIMCacheStore) Set(key string, value []byte) (err error) {
	c.cache.Set(key, value, imcache.WithNoExpiration())
	return
}

func (c *cIMCacheStore) Delete(key string) (err error) {
	c.cache.Remove(key)
	return
}