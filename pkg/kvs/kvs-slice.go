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
	_ = GetUnmarshal(store, key, &list)
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