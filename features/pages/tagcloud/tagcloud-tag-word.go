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

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/maps"
)

type tagWord struct {
	tag     string
	count   int
	shasums map[string]struct{}
	sync.RWMutex
}

func newTagWord(tag string) (t *tagWord) {
	return &tagWord{
		tag:     tag,
		shasums: make(map[string]struct{}),
	}
}

func (t *tagWord) Empty() (empty bool) {
	t.RLock()
	defer t.RUnlock()
	empty = len(t.shasums) == 0
	return
}

func (t *tagWord) Shasums() (shasums []string) {
	t.RLock()
	defer t.RUnlock()
	shasums = maps.Keys(t.shasums)
	return
}

func (t *tagWord) Get() (tag *feature.CloudTag) {
	t.RLock()
	defer t.RUnlock()
	return &feature.CloudTag{
		Word:  t.tag,
		Count: t.count,
	}
}

func (t *tagWord) Add(shasum string) {
	t.Lock()
	defer t.Unlock()
	t.shasums[shasum] = struct{}{}
	t.count = len(t.shasums)
}

func (t *tagWord) Rem(shasum string) {
	t.Lock()
	defer t.Unlock()
	if _, present := t.shasums[shasum]; present {
		delete(t.shasums, shasum)
	}
	t.count = len(t.shasums)
}
