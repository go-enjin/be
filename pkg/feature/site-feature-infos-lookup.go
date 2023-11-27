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

package feature

import (
	"github.com/go-enjin/be/pkg/slices"
)

type SiteInfosLookup[T interface{}] struct {
	order []string
	infos map[string]T
}

func NewSiteInfosLookup[T interface{}]() (lookup *SiteInfosLookup[T]) {
	lookup = &SiteInfosLookup[T]{
		order: make([]string, 0),
		infos: make(map[string]T),
	}
	return
}

func (l *SiteInfosLookup[T]) Len() (count int) {
	count = len(l.order)
	return
}

func (l *SiteInfosLookup[T]) Keys() (keys []string) {
	keys = append(keys, l.order...)
	return
}

func (l *SiteInfosLookup[T]) Get(key string) (info T) {
	info, _ = l.infos[key]
	return
}

func (l *SiteInfosLookup[T]) Set(key string, info T) {
	l.infos[key] = info
	l.order = append(slices.Prune(l.order, key), key)
	return
}