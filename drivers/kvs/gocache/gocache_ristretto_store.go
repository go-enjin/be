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

	"github.com/dgraph-io/ristretto"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
)

var _ feature.KeyValueStore = (*cRistrettoStore)(nil)

type cRistrettoStore struct {
	cache *ristretto.Cache
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

func (c *cRistrettoStore) Get(key string) (value []byte, err error) {
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

func (c *cRistrettoStore) Set(key string, value []byte) (err error) {
	if !c.cache.Set(key, value, 0) {
		log.FatalF("ristretto set dropped for key: %v", key)
	}
	return
}

func (c *cRistrettoStore) Delete(key string) (err error) {
	c.cache.Del(key)
	return
}