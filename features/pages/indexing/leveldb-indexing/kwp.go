//go:build leveldb_indexing || all

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
	"fmt"
	"strconv"
	"strings"

	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/pagecache"
)

func (f *CFeature) KnownKeywords() (keywords []string) {
	var err error

	iter := f.leveldb.NewIterator(util.BytesPrefix([]byte("contents:")), nil)
	for iter.Next() {
		dataKey := iter.Key()
		thisKey := string(dataKey)
		if !strings.HasPrefix(thisKey, "contents:") {
			break
		}
		thisWord := strings.TrimPrefix(thisKey, "contents:")
		keywords = append(keywords, thisWord)
	}
	iter.Release()
	if err != nil {
		return
	}
	if err = iter.Error(); err != nil {
		log.Error(err)
	}

	// var iterate leveldb.LookupIterator
	// if iterate, err = snapshot.Lookup([]byte("contents:"), nil); err != nil {
	// 	log.Error(err)
	// } else {
	// 	var dataKey []byte
	// 	var dataErr error
	// 	for {
	// 		if dataKey, _, dataErr = iterate.Next(); dataErr == leveldb.EndOfIterator {
	// 			break
	// 		}
	// 		thisKey := string(dataKey)
	// 		if !strings.HasPrefix(thisKey, "contents:") {
	// 			break
	// 		}
	// 		thisWord := strings.TrimPrefix(thisKey, "contents:")
	// 		keywords = append(keywords, thisWord)
	// 	}
	// }

	return
}

func (f *CFeature) KeywordStubs(keyword string) (stubs pagecache.Stubs) {
	var err error

	var data []byte
	if data, err = f.leveldb.Get([]byte(fmt.Sprintf("contents:%s", keyword)), nil); err != nil {
		log.Error(err)
		return
	}

	var idxs []int
	for _, s := range strings.Split(string(data), ",") {
		if v, e := strconv.Atoi(s); e == nil {
			idxs = append(idxs, v)
		}
	}

	var shasums []string
	if shasums, err = f.getShasumsFromIndexes(idxs); err != nil {
		log.Error(err)
		return
	}

	for _, shasum := range shasums {
		if stub, ok := f.stubs[shasum]; ok {
			stubs = append(stubs, stub)
		} else {
			log.ErrorF("stub not found: %v", shasum)
		}
	}

	return
}