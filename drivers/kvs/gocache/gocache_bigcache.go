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
	"time"

	"github.com/go-enjin/be/pkg/log"
)

type BigCacheSupport interface {
	AddBigCache(name string, buckets ...string) MakeFeature
	AddExpiringBigCache(name string, lifeWindow, cleanWindow time.Duration, buckets ...string) MakeFeature
}

func (f *CFeature) AddBigCache(name string, buckets ...string) MakeFeature {
	return f.AddExpiringBigCache(name, NoExpiration, NoExpiration, buckets...)
}

func (f *CFeature) AddExpiringBigCache(name string, lifeWindow, cleanWindow time.Duration, buckets ...string) MakeFeature {
	f.Lock()
	defer f.Unlock()
	f.addCache(name, newBigCache(lifeWindow, cleanWindow))
	for _, bucket := range buckets {
		if _, err := f.caches[name].AddBucket(bucket); err != nil {
			log.FatalDF(1, "error adding bucket to cache: %v - %v", name, bucket)
		}
	}
	return f
}