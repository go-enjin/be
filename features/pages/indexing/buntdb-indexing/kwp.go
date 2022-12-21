//go:build buntdb_indexing || buntdb || all

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

package indexing

import (
	"strconv"
	"strings"

	"github.com/tidwall/buntdb"

	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/pagecache"
)

func (f *CFeature) KnownKeywords() (keywords []string) {
	// f.RLock()
	// defer f.RUnlock()
	tmp := make(map[string]bool)
	if err := f.kvs.DB("contents:").View(func(tx *buntdb.Tx) (err error) {
		_ = tx.AscendKeys("contents:*", func(k, v string) (ok bool) {
			word := strings.TrimPrefix(k, "contents:")
			tmp[word] = true
			ok = true
			return
		})
		return
	}); err != nil {
		log.ErrorF("error getting known keywords: %v", err)
	}
	keywords = maps.SortedKeys(tmp)
	return
}

func (f *CFeature) KeywordStubs(keyword string) (stubs pagecache.Stubs) {
	f.RLock()
	defer f.RUnlock()
	tmp := make(map[int]bool)
	if err := f.kvs.DB("contents:").View(func(tx *buntdb.Tx) (err error) {
		err = tx.AscendKeys("contents:"+keyword, func(k, v string) (ok bool) {
			for _, s := range strings.Split(v, ",") {
				var idx int
				if idx, err = strconv.Atoi(s); err != nil {
					log.ErrorF("error parsing stub index: \"%v\" - %v", s, err)
				} else {
					tmp[idx] = true
				}
			}
			ok = true
			return
		})
		return
	}); err != nil {
		log.ErrorF("error getting known keywords: %v", err)
	}
	log.WarnF("found %d stub indexes for: %v", len(tmp), keyword)
	for idx, _ := range tmp {
		if shasum, err := f.getShasumByIndex(idx); err != nil {
			log.ErrorF("error getting shasum by index: %v - %v", idx, err)
		} else {
			if stub, ok := f.stubs[shasum]; ok {
				stubs = append(stubs, stub)
				log.WarnF("found stub index %d: %v", idx, shasum)
			} else {
				log.ErrorF("stub not found: %v", shasum)
			}
		}
	}
	return
}