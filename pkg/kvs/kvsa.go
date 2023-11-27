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
	"github.com/go-enjin/be/pkg/feature"
)

type cKVSA struct {
	kvs feature.KeyValueStore
}

func NewKVSA(kvs feature.KeyValueStore) (kvsa feature.KeyValueStoreAny) {
	kvsa = &cKVSA{
		kvs: kvs,
	}
	return
}

func (w *cKVSA) Get(key interface{}) (value interface{}, ok bool) {
	v, err := w.kvs.Get(key)
	if ok = err == nil; ok {
		value = v
	}
	return
}

func (w *cKVSA) Set(key interface{}, value interface{}) {
	_ = w.kvs.Set(key, value)
	return
}
