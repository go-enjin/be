//go:build (driver_kvs_gocache && (driver_kvs_gocache_redis || redis)) || drivers_kvs || drivers || all

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
	"strings"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/redis/go-redis/v9"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
)

var _ feature.ExtendedKeyValueStore = (*cRedisStore)(nil)

type cRedisStore struct {
	tag    string
	name   string
	cache  *cache.Cache[string]
	client *redis.Client
}

func (c *cRedisStore) MakeKey(prefix string) (key string) {
	key = prefix + ":" + c.name + ":" + c.tag
	return
}

func (c *cRedisStore) makePattern(prefix string) (pattern string) {
	pattern = c.MakeKey(prefix + "*")
	return
}

func (c *cRedisStore) Get(key string) (value []byte, err error) {
	var data string
	if data, err = c.cache.Get(context.Background(), c.MakeKey(key)); err != nil {
		return
	}
	value = []byte(data)
	return
}

func (c *cRedisStore) Set(key string, value []byte) (err error) {
	key = c.MakeKey(key)
	err = c.cache.Set(context.Background(), key, string(value))
	return
}

func (c *cRedisStore) Delete(key string) (err error) {
	key = c.MakeKey(key)
	err = c.cache.Delete(context.Background(), key)
	return
}

func (c *cRedisStore) DoScan(ctx context.Context, pattern string, fn DoScanRedisFn) {
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	defer func() {
		if err := iter.Err(); err != nil {
			log.ErrorF("error scanning %q (%q) for range %q: %v", c.tag, c.name, pattern, err)
		}
	}()
	for iter.Next(ctx) {
		if stopped := fn(iter); stopped {
			return
		}
	}
}

func (c *cRedisStore) Size() (count int) {
	ctx := context.Background()
	pattern := c.makePattern("")
	c.DoScan(ctx, pattern, func(iter *redis.ScanIterator) (stop bool) {
		count += 1
		return
	})
	return
}

func (c *cRedisStore) Keys(prefix string) (keys []string) {
	ctx := context.Background()
	bucket := c.MakeKey("")
	pattern := c.makePattern(prefix)
	c.DoScan(ctx, pattern, func(iter *redis.ScanIterator) (stop bool) {
		key := strings.TrimSuffix(iter.Val(), bucket)
		keys = append(keys, key)
		return
	})
	return
}

func (c *cRedisStore) StreamKeys(prefix string, ctx context.Context) (keys chan string) {
	keys = make(chan string)
	go func() {
		if ctx == nil {
			ctx = context.Background()
		}
		bucket := c.MakeKey("")
		pattern := c.makePattern(prefix)
		c.DoScan(ctx, pattern, func(iter *redis.ScanIterator) (stop bool) {
			keys <- strings.TrimSuffix(iter.Val(), bucket)
			return
		})
		close(keys)
	}()
	return
}

func (c *cRedisStore) Range(prefix string, fn feature.KeyValueStoreRangeFn) {
	ctx := context.Background()
	bucket := c.MakeKey("")
	pattern := c.makePattern(prefix)
	c.DoScan(ctx, pattern, func(iter *redis.ScanIterator) (stop bool) {
		key := iter.Val()
		if v := c.client.Get(ctx, key); v != nil {
			k := strings.TrimSuffix(key, bucket)
			if stop = fn(k, []byte(v.Val())); stop {
				return
			}
		}
		return
	})
}

type DoScanRedisFn func(iter *redis.ScanIterator) (stop bool)

type UnsafeRedisStore interface {
	feature.ExtendedKeyValueStore

	MakeKey(prefix string) (key string)
	DoScan(ctx context.Context, pattern string, fn DoScanRedisFn)
	Client() (client *redis.Client)
}

func (c *cRedisStore) Client() (client *redis.Client) {
	client = c.client
	return
}
