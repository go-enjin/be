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
	"time"

	"github.com/go-enjin/be/pkg/log"
)

type RedisSupport interface {
	AddRedisCache(name string, buckets ...string) MakeFeature
	AddExpiringRedisCache(name string, lifeWindow time.Duration, buckets ...string) MakeFeature
}

func (f *CFeature) AddRedisCache(name string, buckets ...string) MakeFeature {
	return f.AddExpiringRedisCache(name, NoExpiration, buckets...)
}

func (f *CFeature) AddExpiringRedisCache(name string, lifeWindow time.Duration, buckets ...string) MakeFeature {
	f.Lock()
	defer f.Unlock()
	var err error
	var redisCache *cRedisCache
	if redisCache, err = newRedisCache(f, name, lifeWindow); err != nil {
		log.FatalDF(1, "error adding new redis cache: %v", err)
	}
	f.addCache(name, redisCache)
	for _, bucket := range buckets {
		if _, err = f.caches[name].AddBucket(bucket); err != nil {
			log.FatalDF(1, "error adding bucket to cache: %v - %v", name, bucket)
		}
	}
	return f
}
