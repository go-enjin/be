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

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/gob"
	"github.com/go-enjin/be/pkg/maths"
)

type privateKey string

const (
	gNilValue privateKey = "nil"
)

type Variables interface {
	maths.Number | byte | string
}

func GetSlice[T Variables](store feature.KeyValueStore, key interface{}) (values []T, err error) {
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

func RemoveFromSlice[T Variables](store feature.KeyValueStore, key interface{}, values ...T) (err error) {
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

func SetSlice[T Variables](store feature.KeyValueStore, key interface{}, values []T) (err error) {
	err = store.Set(key, values)
	return
}

func AppendToSlice[T Variables](store feature.KeyValueStore, key interface{}, values ...T) (err error) {
	var list []T
	if v, e := store.Get(key); e == nil {
		list, _ = v.([]T)
	}
	err = store.Set(key, append(values, list...))
	return
}

func StringSliceEmpty(store feature.KeyValueStore, key interface{}) (empty bool) {
	var err error
	var v interface{}
	if v, err = store.Get(key); err != nil {
		return
	}
	vs, _ := v.(string)
	empty = vs == ""
	return
}

func GetStringSlice(store feature.KeyValueStore, key interface{}) (values []string, err error) {
	var v interface{}
	if v, err = store.Get(key); err != nil {
		return
	}
	if vs, ok := v.(string); !ok {
		err = fmt.Errorf("value of %v is not nl-string", key)
	} else {
		values = strings.Split(vs, "\n")
	}
	return
}

func AppendToStringSlice(store feature.KeyValueStore, key interface{}, values ...string) (err error) {
	var list string
	if v, e := store.Get(key); e == nil {
		list, _ = v.(string)
	}
	combined := strings.Join(values, "\n")
	if list != "" {
		if combined != "" {
			combined += "\n"
		}
		combined += list
	}
	err = store.Set(key, combined)
	return
}

func GetValue[T interface{}](store feature.KeyValueStore, key interface{}) (value T) {
	if v, e := store.Get(key); e == nil {
		if vt, ok := v.(T); ok {
			value = vt
		}
	}
	return
}

func AddToNumber[T maths.Number](store feature.KeyValueStore, key interface{}, increment T) (updated T, err error) {
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

func GetFlatList[T interface{}](store feature.KeyValueStore, key string) (values []T) {
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

func EncodeKeyValue(value interface{}) (valueKey string, err error) {
	var v []byte
	if v, err = gob.Encode(value); err != nil {
		return
	} else {
		valueKey = string(v)
	}
	return
}

func DecodeKeyValue(valueKey string) (value interface{}, err error) {
	value, err = gob.Decode([]byte(valueKey))
	return
}

func DecodeValue[T interface{}](encoded string) (value T, err error) {
	var v interface{}
	if v, err = gob.Decode([]byte(encoded)); err == nil {
		value, _ = v.(T)
	}
	return
}

func FlatListEmpty(store feature.KeyValueStore, key string) (empty bool) {
	empty = CountFlatList(store, key) == 0
	return
}

func CountFlatList(store feature.KeyValueStore, key string) (count uint64) {
	countKey := MakeFlatListKey(key, "count")
	count = GetValue[uint64](store, countKey)
	return
}

func ResetFlatList(store feature.KeyValueStore, key string) (reset bool) {
	endKey := MakeFlatListKey(key, "end")
	endIndex := GetValue[uint64](store, endKey)
	freeKey := MakeFlatListKey(key, "free")
	countKey := MakeFlatListKey(key, "count")

	for i := uint64(0); i < endIndex; i++ {
		idxKey := MakeFlatListKey(key, "idx", fmt.Sprintf("%d", i))
		if err := store.Delete(idxKey); err != nil {
			panic(err)
		}
	}

	if err := store.Delete(freeKey); err != nil {
		panic(err)
	}

	if err := store.Delete(countKey); err != nil {
		panic(err)
	}
	return
}

func ResetFlatListIfEmpty(store feature.KeyValueStore, key string) (reset bool) {
	count := CountFlatList(store, key)
	if reset = count == 0; reset {
		reset = ResetFlatList(store, key)
	}
	return
}

func CountFlatListValues[T comparable](store feature.KeyValueStore, key string) (counts map[T]uint64) {
	counts = make(map[T]uint64)
	for v := range YieldFlatList[interface{}](store, key) {
		if value, ok := v.(T); ok {
			counts[value] += 1
			continue
		}
		if values, ok := v.([]T); ok {
			for _, value := range values {
				counts[value] += 1
			}
		}
	}
	return
}

func CountDistinctFlatListValues[T comparable](store feature.KeyValueStore, key string) (count uint64) {
	track := make(map[T]struct{})
	for v := range YieldFlatList[interface{}](store, key) {
		if value, ok := v.(T); ok {
			track[value] = struct{}{}
			continue
		}
		if values, ok := v.([]T); ok {
			for _, value := range values {
				track[value] = struct{}{}
			}
		}
	}
	count = uint64(len(track))
	return
}

func YieldFlatList[T interface{}](store feature.KeyValueStore, key string) (yield chan T) {
	yield = make(chan T)
	go func(store feature.KeyValueStore, key string, yield chan T) {
		defer close(yield)
		endKey := MakeFlatListKey(key, "end")
		endIdx := GetValue[uint64](store, endKey)

		for i := uint64(0); i < endIdx; i++ {
			idxKey := MakeFlatListKey(key, "idx", fmt.Sprintf("%d", i))
			if v, e := store.Get(idxKey); e == nil {
				if isNil, ok := v.(privateKey); ok {
					if isNil == gNilValue {
						continue
					}
				}
				if vs, ok := v.(string); ok {
					if item, ee := DecodeValue[T](vs); ee == nil {
						yield <- item
						continue
					}
				}
				if item, ok := v.(T); ok {
					yield <- item
				}
			}
		}

	}(store, key, yield)
	return
}

func FirstInFlatList[T comparable](store feature.KeyValueStore, key string) (value T, ok bool) {

	if CountFlatList(store, key) == 0 {
		return
	}

	endKey := MakeFlatListKey(key, "end")
	endIndex := GetValue[uint64](store, endKey)

	for i := uint64(0); i < endIndex; i++ {
		idxKey := MakeFlatListKey(key, "idx", fmt.Sprintf("%d", i))
		if v, e := store.Get(idxKey); e == nil {
			if isNil, present := v.(privateKey); present && isNil == gNilValue {
				continue
			} else if value, ok = v.(T); ok {
				return
			}
		}
	}

	return
}

func LastInFlatList[T comparable](store feature.KeyValueStore, key string) (value T, ok bool) {

	if CountFlatList(store, key) == 0 {
		return
	}

	endKey := MakeFlatListKey(key, "end")
	endIndex := GetValue[uint64](store, endKey)

	for i := endIndex - 1; i > 0; i-- {
		idxKey := MakeFlatListKey(key, "idx", fmt.Sprintf("%d", i))
		if v, e := store.Get(idxKey); e == nil {
			if isNil, present := v.(privateKey); present && isNil == gNilValue {
				continue
			} else if value, ok = v.(T); ok {
				return
			}
		}
	}

	return
}

func AppendToFlatList[T comparable](store feature.KeyValueStore, key string, value T) (err error) {
	endKey := MakeFlatListKey(key, "end")
	endIndex := GetValue[uint64](store, endKey)
	freeKey := MakeFlatListKey(key, "free")
	freeIndexes, _ := GetSlice[uint64](store, freeKey)
	countKey := MakeFlatListKey(key, "count")

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
		return
	}

	if err = store.Set(countKey, GetValue[uint64](store, countKey)+1); err != nil {
		err = fmt.Errorf("error storing flat-list count at key: %v - %v", countKey, err)
		return
	}

	return
}

// TODO: figure out how to shrink kvs flat lists

func RemoveFromFlatList[T comparable](store feature.KeyValueStore, key string, value T) (err error) {
	endKey := MakeFlatListKey(key, "end")
	endIndex := GetValue[uint64](store, endKey)
	freeKey := MakeFlatListKey(key, "free")
	freeIndexes, _ := GetSlice[uint64](store, freeKey)
	countKey := MakeFlatListKey(key, "count")

	var found bool
	var rmIdx uint64
	for idx := uint64(0); idx <= endIndex; idx++ {
		idxKey := MakeFlatListKey(key, "idx", fmt.Sprintf("%d", idx))
		if v, e := store.Get(idxKey); e == nil {
			if t, ok := v.(T); ok {
				if found = t == value; found {
					rmIdx = idx
				}
			}
		}
	}

	if found {
		rmKey := MakeFlatListKey(key, "idx", fmt.Sprintf("%d", rmIdx))
		freeIndexes = append(freeIndexes, rmIdx)
		if err = SetSlice[uint64](store, freeKey, freeIndexes); err != nil {
			return
		}
		if err = store.Set(rmKey, gNilValue); err != nil {
			return
		}

		if err = store.Set(countKey, GetValue[uint64](store, countKey)-1); err != nil {
			err = fmt.Errorf("error storing flat-list count at key: %v - %v", countKey, err)
			return
		}
	}

	return
}
