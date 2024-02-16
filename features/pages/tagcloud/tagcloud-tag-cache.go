// Copyright (c) 2024  The Go-Enjin Authors
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

package tagcloud

import (
	"sync"

	"github.com/go-corelibs/slices"
	"github.com/go-enjin/be/pkg/feature"
)

type tagCache struct {
	l map[string]*tagWord
	m *sync.RWMutex
}

func newTagCache() (tc *tagCache) {
	return &tagCache{
		l: make(map[string]*tagWord),
		m: &sync.RWMutex{},
	}
}

func (tc *tagCache) TagCloud() (list feature.TagCloud) {
	tc.m.RLock()
	defer tc.m.RUnlock()
	for _, word := range tc.l {
		list = append(list, word.Get())
	}
	return
}

func (tc *tagCache) Find(shasum string) (list feature.TagCloud) {
	tc.m.RLock()
	defer tc.m.RUnlock()
	for _, word := range tc.l {
		if slices.Within(shasum, word.Shasums()) {
			list = append(list, word.Get())
		}
	}
	list.Sort()
	return
}

func (tc *tagCache) Get(word string) (tw *tagWord, ok bool) {
	tc.m.RLock()
	defer tc.m.RUnlock()
	tw, ok = tc.l[word]
	return
}

func (tc *tagCache) Add(shasum string, tags ...string) {
	tc.m.Lock()
	defer tc.m.Unlock()
	for _, tag := range tags {
		if _, ok := tc.l[tag]; !ok {
			tc.l[tag] = newTagWord(tag)
		}
		tc.l[tag].Add(shasum)
	}
}

func (tc *tagCache) Remove(shasum string, tags ...string) {
	tc.m.Lock()
	defer tc.m.Unlock()
	for _, tag := range tags {
		if tw, ok := tc.l[tag]; ok {
			if tw.Rem(shasum); tw.Empty() {
				delete(tc.l, tag)
			}
		}
	}
}
