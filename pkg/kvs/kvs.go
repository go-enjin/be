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
	"bytes"
	"encoding/gob"
	"fmt"
	"strings"

	"github.com/go-enjin/be/pkg/feature"
	beGob "github.com/go-enjin/be/pkg/gob"
	"github.com/go-enjin/be/pkg/maths"
)

type privateKey string

const (
	gNilValue privateKey = "nil"
)

type Variables interface {
	maths.Number | byte | string
}

func SetMarshal(store feature.KeyValueStore, key string, value interface{}) (err error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err = enc.Encode(value); err != nil {
		return
	}
	err = store.Set(key, buf.Bytes())
	return
}

func GetUnmarshal[T interface{}](store feature.KeyValueStore, key string, value *T) (err error) {
	var data []byte
	if data, err = store.Get(key); err != nil {
		return
	}
	dec := gob.NewDecoder(bytes.NewBuffer(data))
	err = dec.Decode(value)
	return
}

func GetSlice[T Variables](store feature.KeyValueStore, key string) (values []T, err error) {
	err = GetUnmarshal(store, key, &values)
	return
}

func RemoveFromSlice[T Variables](store feature.KeyValueStore, key string, values ...T) (err error) {
	var list, items []T
	lookup := make(map[T]bool)
	for _, value := range values {
		lookup[value] = true
	}
	if err = GetUnmarshal(store, key, &items); err != nil {
		return
	}
	for _, item := range items {
		if _, remove := lookup[item]; !remove {
			list = append(list, item)
		}
	}
	err = SetMarshal(store, key, list)
	return
}

func SetSlice[T Variables](store feature.KeyValueStore, key string, values []T) (err error) {
	err = SetMarshal(store, key, values)
	return
}

func AppendToSlice[T Variables](store feature.KeyValueStore, key string, values ...T) (err error) {
	var list []T
	if err = GetUnmarshal(store, key, &list); err != nil {
		return
	}
	err = SetMarshal(store, key, append(list, values...))
	return
}

func StringSliceEmpty(store feature.KeyValueStore, key string) (empty bool) {
	var err error
	var values []string
	if err = GetUnmarshal(store, key, &values); err == nil {
		empty = len(values) == 0
	}
	return
}

func GetStringSlice(store feature.KeyValueStore, key string) (values []string, err error) {
	err = GetUnmarshal(store, key, &values)
	return
}

func AppendToStringSlice(store feature.KeyValueStore, key string, values ...string) (err error) {
	var list []string
	if err = GetUnmarshal(store, key, &list); err != nil {
		return
	}
	list = append(list, values...)
	err = SetMarshal(store, key, list)
	return
}

func GetValue[T interface{}](store feature.KeyValueStore, key string) (value T) {
	_ = GetUnmarshal(store, key, &value)
	return
}

func AddToNumber[T maths.Number](store feature.KeyValueStore, key string, increment T) (updated T, err error) {
	var current T
	_ = GetUnmarshal(store, key, &current)
	err = SetMarshal(store, key, current+increment)
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
		var value T
		if e := GetUnmarshal(store, idxKey, &value); e == nil {
			values = append(values, value)
		}
	}

	return
}

func EncodeKeyValue(value interface{}) (valueKey string, err error) {
	var v []byte
	if v, err = beGob.Encode(value); err != nil {
		return
	} else {
		valueKey = string(v)
	}
	return
}

func DecodeKeyValue(valueKey string) (value interface{}, err error) {
	value, err = beGob.Decode([]byte(valueKey))
	return
}

func DecodeValue[T interface{}](encoded string) (value T, err error) {
	var v interface{}
	if v, err = beGob.Decode([]byte(encoded)); err == nil {
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

func GetIsNil(store feature.KeyValueStore, key string) (isNil bool) {
	var pkv privateKey
	if e := GetUnmarshal(store, key, &pkv); e == nil {
		isNil = pkv == gNilValue
	}
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