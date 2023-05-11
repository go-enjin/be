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
	"strings"
)

type Numbers interface {
	uint | uint8 | uint16 | uint32 | uint64 |
		int | int8 | int16 | int32 | int64 |
		float32 | float64
}

type Contents interface {
	byte | string
}

type Variables interface {
	Numbers | Contents
}

func GetSlice[T Variables](store KeyValueStore, key interface{}) (values []T, err error) {
	var v interface{}
	if v, err = store.Get(key); err != nil {
		return
	}
	var ok bool
	if values, ok = v.([]T); !ok {
		err = fmt.Errorf("value of %v is not %T", key, ([]T)(nil))
	}
	return
}

func RemoveFromSlice[T Variables](store KeyValueStore, key interface{}, values ...T) (err error) {
	var list []T
	lookup := make(map[T]bool)
	for _, value := range values {
		lookup[value] = true
	}
	if v, e := store.Get(key); e == nil {
		if items, ok := v.([]T); ok {
			for _, item := range items {
				if _, remove := lookup[item]; !remove {
					list = append(list, item)
				}
			}
		}
	}
	err = store.Set(key, list)
	return
}

func AppendToSlice[T Variables](store KeyValueStore, key interface{}, values ...T) (err error) {
	var list []T
	unique := make(map[T]bool)
	if v, e := store.Get(key); e == nil {
		if items, ok := v.([]T); ok {
			for _, item := range items {
				if _, present := unique[item]; !present {
					unique[item] = true
					list = append(list, item)
				}
			}
		}
	}
	list = append(values, list...)
	err = store.Set(key, list)
	return
}

func GetValue[T interface{}](store KeyValueStore, key interface{}) (value T) {
	if v, e := store.Get(key); e == nil {
		if vt, ok := v.(T); ok {
			value = vt
		}
	}
	return
}

func AddToNumber[T Numbers](store KeyValueStore, key interface{}, increment T) (updated T, err error) {
	if v, e := store.Get(key); e == nil {
		if vt, ok := v.(T); ok {
			updated = vt + increment
		}
	}
	err = store.Set(key, updated)
	return
}

func MakeFlatListKey(key string, suffixes ...string) (name string) {
	name = key
	if len(suffixes) > 0 {
		name += "__" + strings.Join(suffixes, "__")
	}
	return
}

func GetFlatList[T interface{}](store KeyValueStore, key string, value T) (values []T) {
	endKey := MakeFlatListKey(key, "end")
	endIdx := GetValue[uint64](store, endKey)

	for i := uint64(0); i < endIdx; i++ {
		idxKey := MakeFlatListKey(key, "idx", fmt.Sprintf("%d", i))
		if v, e := store.Get(idxKey); e == nil {
			if item, ok := v.(T); ok {
				values = append(values, item)
			}
		}
	}

	return
}

func YieldFlatList[T interface{}](store KeyValueStore, key string) (yield chan T) {
	yield = make(chan T)
	go func() {
		defer close(yield)
		endKey := MakeFlatListKey(key, "end")
		endIdx := GetValue[uint64](store, endKey)

		for i := uint64(0); i < endIdx; i++ {
			idxKey := MakeFlatListKey(key, "idx", fmt.Sprintf("%d", i))
			if v, e := store.Get(idxKey); e == nil {
				if item, ok := v.(T); ok {
					yield <- item
				}
			}
		}

	}()
	return
}

func AppendToFlatList(store KeyValueStore, key string, value interface{}) (err error) {
	endKey := MakeFlatListKey(key, "end")
	endIndex := GetValue[uint64](store, endKey)
	freeKey := MakeFlatListKey(key, "free")
	freeIndexes, _ := GetSlice[uint64](store, freeKey)

	var dstIdx uint64

	if len(freeIndexes) > 0 {
		if err = store.Set(freeKey, freeIndexes[1:]); err != nil {
			err = fmt.Errorf("error recovering free index: %v - %v", freeKey, err)
			return
		}
		dstIdx = freeIndexes[0]
	} else if err = store.Set(endKey, endIndex+1); err != nil {
		err = fmt.Errorf("error incremented end index: %v - %v", endKey, err)
		return
	} else {
		dstIdx = endIndex
	}

	dstKey := MakeFlatListKey(key, "idx", fmt.Sprintf("%d", dstIdx))
	if err = store.Set(dstKey, value); err != nil {
		err = fmt.Errorf("error storing value at key: %v - %v", dstKey, err)
	}
	return
}