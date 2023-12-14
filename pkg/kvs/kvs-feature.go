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

package kvs

import (
	"fmt"

	"github.com/go-enjin/be/pkg/feature"
)

func AsExtended(store feature.KeyValueStore) (extended feature.ExtendedKeyValueStore, err error) {
	var ok bool
	if extended, ok = interface{}(store).(feature.ExtendedKeyValueStore); !ok {
		err = fmt.Errorf("store does not support feature.ExtendedKeyValueStore: %#+v", store)
	}
	return
}

func MustAsExtended(store feature.KeyValueStore) (extended feature.ExtendedKeyValueStore) {
	var err error
	if extended, err = AsExtended(store); err != nil {
		panic(err)
	}
	return
}

func ExtendedBucket(cache feature.KeyValueCache, name string) (store feature.ExtendedKeyValueStore, err error) {
	var bucket feature.KeyValueStore
	if bucket, err = cache.Bucket(name); err == nil {
		store, err = AsExtended(bucket)
	}
	return
}

func MustExtendedBucket(cache feature.KeyValueCache, name string) (store feature.ExtendedKeyValueStore) {
	var err error
	if store, err = ExtendedBucket(cache, name); err != nil {
		panic(err)
	}
	return
}