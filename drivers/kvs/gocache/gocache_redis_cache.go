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
	"fmt"
	"sync"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	redis_store "github.com/eko/gocache/store/redis/v4"
	"github.com/iancoleman/strcase"
	"github.com/redis/go-redis/v9"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
)

var (
	_ feature.KeyValueCache        = (*cRedisCache)(nil)
	_ feature.KeyValueCacheFeature = (*cRedisCache)(nil)
)

type cRedisCache struct {
	this       *CFeature
	name       string
	client     *redis.Client
	buckets    map[string]*cRedisStore
	lifeWindow time.Duration

	sync.RWMutex
}

func newRedisCache(this *CFeature, name string, lifeWindow time.Duration) (cache *cRedisCache, err error) {
	cache = &cRedisCache{
		this:       this,
		name:       name,
		lifeWindow: lifeWindow,
		buckets:    make(map[string]*cRedisStore),
	}
	return
}

func (c *cRedisCache) makeFlagName() (category, name string) {
	category = c.this.Tag().Kebab()
	name = strcase.ToKebab(category + "-redis-url-" + c.name)
	return
}

func (c *cRedisCache) Build(b feature.Buildable) (err error) {
	category, flagName := c.makeFlagName()
	b.AddFlags(
		&cli.StringFlag{
			Name:     flagName,
			Usage:    "specify the redis connection url for: " + c.name,
			EnvVars:  b.MakeEnvKeys(flagName),
			Category: category,
		},
	)
	return
}

func (c *cRedisCache) Startup(ctx *cli.Context) (err error) {
	_, flagName := c.makeFlagName()
	var options *redis.Options
	if !ctx.IsSet(flagName) {
		err = fmt.Errorf("required flag not present: --%v", flagName)
		return
	}
	url := ctx.String(flagName)
	if options, err = redis.ParseURL(url); err != nil {
		return
	}
	c.client = redis.NewClient(options)
	return
}

func (c *cRedisCache) Shutdown() {
	if err := c.client.Close(); err != nil {
		log.ErrorF("error closing %v redis client: %v", c.name, err)
	}
}

func (c *cRedisCache) ListBuckets() (names []string) {
	names = maps.SortedKeys(c.buckets)
	return
}

func (c *cRedisCache) MustBucket(name string) (kvs feature.KeyValueStore) {
	if v, err := c.Bucket(name); err != nil {
		log.ErrorDF(1, "error getting required bucket \"%v\": - %v", name, err)
		panic(err)
	} else {
		kvs = v
	}
	return
}

func (c *cRedisCache) Bucket(name string) (kvs feature.KeyValueStore, err error) {
	if v, e := c.GetBucket(name); e == nil {
		kvs = v
		return
	}
	kvs, err = c.AddBucket(name)
	return
}

func (c *cRedisCache) AddBucket(name string) (kvs feature.KeyValueStore, err error) {
	c.Lock()
	defer c.Unlock()
	if v, exists := c.buckets[name]; exists {
		kvs = v
		//err = BucketExists
		return
	}

	var options []store.Option
	if c.lifeWindow > 0 {
		options = append(options, store.WithExpiration(c.lifeWindow))
	}
	redisStore := redis_store.NewRedis(c.client, options...)
	cacheManager := cache.New[string](redisStore)
	c.buckets[name] = &cRedisStore{
		tag:    c.this.Tag().Kebab(),
		name:   name,
		cache:  cacheManager,
		client: c.client,
	}
	kvs = c.buckets[name]
	return
}

func (c *cRedisCache) GetBucket(name string) (kvs feature.KeyValueStore, err error) {
	c.RLock()
	defer c.RUnlock()
	if v, ok := c.buckets[name]; ok {
		kvs = v
	} else {
		err = BucketNotFound
	}
	return
}

func (c *cRedisCache) GetBucketSource(name string) (src interface{}) {
	c.RLock()
	defer c.RUnlock()
	if v, ok := c.buckets[name]; ok {
		src = v.cache
	}
	return
}