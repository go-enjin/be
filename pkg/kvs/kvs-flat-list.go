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
)

func MakeFlatListKey(key string, suffixes ...string) (name string) {
	name = key
	if len(suffixes) > 0 {
		name += "__" + strings.Join(suffixes, "__")
	}
	return
}

func GetFlatList[T comparable](store feature.KeyValueStore, key string) (values []T) {
	endKey := MakeFlatListKey(key, "end")
	endIdx := GetValue[uint64](store, endKey)

	for i := uint64(0); i < endIdx; i++ {
		idxKey := MakeFlatListKey(key, "idx", fmt.Sprintf("%d", i))
		var value T
		if e := GetUnmarshal(store, idxKey, &value); e == nil {
			values = append(values, value)
		}
	}

	return
}

func SetFlatList[T comparable](store feature.KeyValueStore, key string, values []T) (err error) {
	ResetFlatList(store, key)
	for _, value := range values {
		if err = AppendToFlatList[T](store, key, value); err != nil {
			return
		}
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
	RangeFlatList(store, key, func(item T) (stop bool) {
		counts[item] += 1
		return
	})
	return
}

func CountDistinctFlatListValues[T comparable](store feature.KeyValueStore, key string) (count uint64) {
	track := make(map[T]struct{})
	RangeFlatList(store, key, func(item T) (stop bool) {
		track[item] = struct{}{}
		return
	})
	count = uint64(len(track))
	return
}

func RangeFlatList[T comparable](store feature.KeyValueStore, key string, fn func(item T) (stop bool)) {
	extended := MustAsExtended(store)
	extended.Range(key+"__idx__", func(key string, value []byte) (stop bool) {
		if GetIsNil(extended, key) {
			return
		}
		var item T
		if e := Unmarshal(value, &item); e != nil {
			return
		} else if stop = fn(item); stop {
			return
		}
		return
	})

	return
}

func YieldFlatList[T comparable](store feature.KeyValueStore, key string) (yield chan T) {
	yield = make(chan T)
	go func(store feature.KeyValueStore, key string, yield chan T) {
		defer close(yield)
		endKey := MakeFlatListKey(key, "end")
		endIdx := GetValue[uint64](store, endKey)

		for i := uint64(0); i < endIdx; i++ {
			idxKey := MakeFlatListKey(key, "idx", fmt.Sprintf("%d", i))
			if GetIsNil(store, idxKey) {
				continue
			}
			var item T
			if e := GetUnmarshal(store, idxKey, &item); e == nil {
				yield <- item
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
		if GetIsNil(store, idxKey) {
			continue
		}
		ok = GetUnmarshal(store, idxKey, &value) == nil
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
		if GetIsNil(store, idxKey) {
			continue
		}
		ok = GetUnmarshal(store, idxKey, &value) == nil
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
		dstIdx = freeIndexes[0]
		if err = SetMarshal(store, freeKey, freeIndexes[1:]); err != nil {
			err = fmt.Errorf("error recovering free index: %v - %v", freeKey, err)
			return
		}
	} else if err = SetMarshal(store, endKey, endIndex+1); err != nil {
		err = fmt.Errorf("error incremented end index: %v - %v", endKey, err)
		return
	} else {
		dstIdx = endIndex
	}

	dstKey := MakeFlatListKey(key, "idx", fmt.Sprintf("%d", dstIdx))
	if err = SetMarshal(store, dstKey, value); err != nil {
		err = fmt.Errorf("error storing value at key: %v - %v", dstKey, err)
		return
	}

	if err = SetMarshal(store, countKey, GetValue[uint64](store, countKey)+1); err != nil {
		err = fmt.Errorf("error storing flat-list count at key: %v - %v", countKey, err)
		return
	}

	return
}

func RemoveFromFlatList[T comparable](store feature.KeyValueStore, key string, value T) (err error) {
	// TODO: figure out how to shrink kvs flat lists

	endKey := MakeFlatListKey(key, "end")
	endIndex := GetValue[uint64](store, endKey)
	freeKey := MakeFlatListKey(key, "free")
	freeIndexes, _ := GetSlice[uint64](store, freeKey)
	countKey := MakeFlatListKey(key, "count")

	var found bool
	var rmIdx uint64
	for idx := uint64(0); idx <= endIndex; idx++ {
		idxKey := MakeFlatListKey(key, "idx", fmt.Sprintf("%d", idx))
		var item T
		if e := GetUnmarshal(store, idxKey, &item); e == nil {
			if found = item == value; found {
				rmIdx = idx
				break
			}
		}
	}

	if found {
		rmKey := MakeFlatListKey(key, "idx", fmt.Sprintf("%d", rmIdx))
		freeIndexes = append(freeIndexes, rmIdx)
		if err = SetSlice[uint64](store, freeKey, freeIndexes); err != nil {
			return
		}

		if err = SetMarshal(store, rmKey, gNilValue); err != nil {
			return
		}

		if err = SetMarshal(store, countKey, GetValue[uint64](store, countKey)-1); err != nil {
			err = fmt.Errorf("error storing flat-list count at key: %v - %v", countKey, err)
			return
		}
	}

	return
}
