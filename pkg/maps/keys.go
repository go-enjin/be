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

package maps

import (
	"cmp"
	"sort"
	"strconv"

	"github.com/maruel/natural"

	"github.com/go-enjin/be/pkg/regexps"
	"github.com/go-enjin/be/pkg/strings"
)

func ValuesSortedByKeys[K ~string, V interface{}](data map[K]V) (values []V) {
	for _, k := range SortedKeys(data) {
		values = append(values, data[k])
	}
	return
}

// SortedKeyLengths returns the list of keys natural sorted and from longest to
// shortest
func SortedKeyLengths[K ~string, V interface{}](data map[K]V) (keys []K) {
	for key := range data {
		keys = append(keys, key)
	}
	// longest -> shortest, natsort same lengths
	sort.Slice(keys, func(i, j int) (less bool) {
		if il, jl := len(keys[i]), len(keys[j]); il == jl {
			less = natural.Less(string(keys[i]), string(keys[j]))
		} else {
			less = il > jl
		}
		return
	})
	return
}

func SortedKeys[K ~string, V interface{}](data map[K]V) (keys []K) {
	for key := range data {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) (less bool) {
		less = natural.Less(string(keys[i]), string(keys[j]))
		return
	})
	return
}

func ReverseSortedKeys[K ~string, V interface{}](data map[K]V) (keys []K) {
	for key := range data {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) (less bool) {
		less = natural.Less(string(keys[j]), string(keys[i]))
		return
	})
	return
}

func OrderedKeys[K cmp.Ordered, V interface{}](data map[K]V) (keys []K) {
	for key := range data {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) (less bool) {
		less = cmp.Less(keys[i], keys[j]) // i vs j
		return
	})
	return
}

func ReverseOrderedKeys[K cmp.Ordered, V interface{}](data map[K]V) (keys []K) {
	for key := range data {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) (less bool) {
		less = cmp.Less(keys[j], keys[i]) // j vs i
		return
	})
	return
}

func SortedKeysByLastKeyword[K ~string, V interface{}](data map[K]V) (keys []K) {
	lookup := make(map[K]string)
	for key := range data {
		keywords := regexps.RxKeywords.FindAllString(string(key), -1)
		lookup[key] = keywords[len(keywords)-1]
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) (less bool) {
		a, b := keys[i], keys[j]
		la, lb := lookup[a], lookup[b]
		if la == lb {
			less = natural.Less(string(a), string(b))
			return
		}
		less = natural.Less(la, lb)
		return less
	})
	return
}

func SortedKeysByLastName[K ~string, V interface{}](data map[K]V) (keys []K) {
	lookup := make(map[K]string)
	for key := range data {
		lookup[key] = strings.LastName(string(key))
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) (less bool) {
		a, b := keys[i], keys[j]
		la, lb := lookup[a], lookup[b]
		if la == lb {
			less = natural.Less(string(a), string(b))
			return
		}
		less = natural.Less(la, lb)
		return less
	})
	return
}

func Keys[K ~string, V interface{}](data map[K]V) (keys []K) {
	for key := range data {
		keys = append(keys, key)
	}
	return
}

func AnyKeys[V interface{}](data map[interface{}]V) (keys []interface{}) {
	for key := range data {
		keys = append(keys, key)
	}
	return
}

func TypedKeys[T comparable, V interface{}](data map[T]V) (keys []T) {
	for key := range data {
		keys = append(keys, key)
	}
	return
}

func ParseKeySlice(input string) (key string, idx int, ok bool) {
	var err error
	if ok = regexps.RxKeySlice.MatchString(input); ok {
		km := regexps.RxKeySlice.FindStringSubmatch(input)
		if km[2] == "" {
			idx = -1
		} else {
			if idx, err = strconv.Atoi(km[2]); err != nil {
				ok = false
				idx = -1
				return
			}
		}
		key = km[1]
	}
	return
}
