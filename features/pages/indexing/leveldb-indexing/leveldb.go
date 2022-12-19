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
	"strconv"

	"github.com/go-enjin/be/pkg/log"
)

func (f *CFeature) leveldbGetIntVal(key []byte) (idx int, err error) {
	idx = -1
	var data []byte
	data, _ = f.leveldb.Get(key, nil)
	if sData := string(data); sData != "" {
		if idx, err = strconv.Atoi(sData); err != nil {
			log.Error(err)
		}
	}
	// log.WarnDF(1, "got: %v == %v", string(key), idx)
	return
}

func (f *CFeature) leveldbIncIntVal(key []byte) (err error) {
	idx := -1
	if d, e := f.leveldb.Get(key, nil); e == nil {
		if s := string(d); s != "" {
			if v, ee := strconv.Atoi(s); ee != nil {
				log.Error(ee)
			} else {
				idx = v
			}
		}
	}
	idx += 1

	if err = f.leveldb.Put(key, []byte(strconv.Itoa(idx)), nil); err != nil {
		log.Error(err)
	}

	// log.WarnDF(1, "set: %v == %v", string(key), idx)
	return
}