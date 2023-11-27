// Copyright (c) 2022  The Go-Enjin Authors
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

package kvm

import (
	"fmt"
	"regexp"
	"strconv"
)

// TODO: consolidate how page indexing and other features use key-value-stores

type HashMap struct {
	store KeyValueStore
	name  string
	keys  map[string]int
}

func NewHashMap(store KeyValueStore, name string, keys ...string) (h *HashMap, err error) {
	h = new(HashMap)
	h.store = store
	h.name = name
	h.keys = make(map[string]int)
	for _, key := range keys {
		h.keys[key] = -1
		p := h.makePattern(key)
		if err = h.store.Iterate(p, func(k string, v string) (ok bool) {
			_, idx := h.parseStoreKey(k)
			if idx > h.keys[key] {
				h.keys[key] = idx
			}
			return
		}); err != nil {
			return
		}
		h.keys[key] += 1
	}
	return
}

var RxStoreKey = regexp.MustCompile(`^([^:]+):(\d+):(.*)$`)

func (h *HashMap) parseStoreKey(storeKey string) (key string, index int) {
	index = -1
	if RxStoreKey.MatchString(storeKey) {
		m := RxStoreKey.FindAllStringSubmatch(storeKey, 1)
		if v, err := strconv.Atoi(m[0][1]); err == nil {
			key = m[0][2]
			index = v
		}
	}
	return
}

func (h *HashMap) makeKey(key string, index int) (built string) {
	built = fmt.Sprintf("%v:%d:%v", h.name, index, key)
	return
}

func (h *HashMap) makePattern(key string) (pattern string) {
	pattern = fmt.Sprintf("%v:*:%v", h.name, key)
	return
}

func (h *HashMap) Get(key string, index int) (value interface{}, err error) {
	var vs string
	k := h.makeKey(key, index)
	if vs, err = h.store.Get(k); err == nil {
		v := new(Value)
		if err = v.UnmarshalBinary([]byte(vs)); err == nil {
			value = v.Get()
		}
	}
	return
}

func (h *HashMap) Set(key string, index int, value interface{}) (err error) {
	v := NewValue(value)
	k := h.makeKey(key, index)
	var data []byte
	if data, err = v.MarshalBinary(); err != nil {
		return
	}
	err = h.store.Set(k, string(data))
	return
}

func (h *HashMap) Append(key string, value interface{}) (err error) {
	if next, ok := h.keys[key]; ok {
		if err = h.Set(key, next, value); err == nil {
			h.keys[key] += 1
		}
	}
	return
}

func (h *HashMap) Iterate(key string, iterator KeyValueIteratorFunc) (err error) {
	pattern := h.makePattern(key)
	err = h.store.Iterate(pattern, iterator)
	return
}
