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

	"github.com/eko/gocache/lib/v4/cache"

	"github.com/go-enjin/be/pkg/feature"
)

var _ feature.KeyValueStore = (*cRedisStore)(nil)

type cRedisStore struct {
	tag   string
	name  string
	cache *cache.Cache[string]
}

func (f *cRedisStore) makeKey(k string) (key string) {
	key = f.tag + "__" + f.name + "__" + k
	return
}

func (f *cRedisStore) Get(key string) (value []byte, err error) {
	var data string
	if data, err = f.cache.Get(context.Background(), f.makeKey(key)); err != nil {
		return
	}
	value = []byte(data)
	return
}

func (f *cRedisStore) Set(key string, value []byte) (err error) {
	key = f.makeKey(key)
	err = f.cache.Set(context.Background(), key, string(value))
	return
}

func (f *cRedisStore) Delete(key string) (err error) {
	key = f.makeKey(key)
	err = f.cache.Delete(context.Background(), key)
	return
}