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
	"regexp"
	"strconv"
	"strings"

	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/maps/kvm"
)

const (
	gStubIndexKeyFormat  = "stub:%d:index"
	gStubShasumKeyFormat = "stub:%s:shasum"

	gPqlNextKeyFormat         = "pql:next:%s"
	gPqlDataKeyFormat         = "pql:data:%s:%s"
	gPqlContextKeyStubsFormat = "pql:%s:%d:stubs"
	gPqlContextKeyValueFormat = "pql:%s:%d:value"
)

var (
	gStubNextKey = []byte("stub:next")
)

func (f *CFeature) getIndexByShasum(shasum string) (idx int, err error) {
	idx = -1

	var data []byte
	stubKey := []byte(fmt.Sprintf(gStubShasumKeyFormat, shasum))
	data, _ = f.leveldb.Get(stubKey, nil)

	if sData := string(data); sData != "" {
		if idx, err = strconv.Atoi(sData); err != nil {
			log.Error(err)
		}
	}
	return
}

func (f *CFeature) getShasumByIndex(idx int) (shasum string, err error) {

	var data []byte
	stubKey := []byte(fmt.Sprintf(gStubIndexKeyFormat, idx))
	if data, err = f.leveldb.Get(stubKey, nil); err != nil {
		log.Error(err)
		return
	}

	shasum = string(data)
	return
}

func (f *CFeature) getShasumsFromIndexes(indexes []int) (shasums []string, err error) {

	cache := make(map[string]bool)
	for _, idx := range indexes {
		var data []byte
		stubIndexKey := []byte(fmt.Sprintf(gStubIndexKeyFormat, idx))
		if data, err = f.leveldb.Get(stubIndexKey, nil); err != nil {
			log.Error(err)
			return
		}
		cache[string(data)] = true
	}

	shasums = maps.Keys(cache)
	return
}

func (f *CFeature) getNextStubIndex() (idx int, err error) {
	if idx, err = f.leveldbGetIntVal(gStubNextKey); err == nil && idx == -1 {
		idx = 0
	}
	return
}

func (f *CFeature) consumeNextStubIndex() (err error) {
	return f.leveldbIncIntVal(gStubNextKey)
}

func (f *CFeature) pqlGetNextValueKeyIndex(contextKey string) (idx int, err error) {
	idx, err = f.leveldbGetIntVal([]byte(fmt.Sprintf(gPqlNextKeyFormat, contextKey)))
	return
}

func (f *CFeature) pqlConsumeNextValueKeyIndex(contextKey string) (err error) {
	return f.leveldbIncIntVal([]byte(fmt.Sprintf(gPqlNextKeyFormat, contextKey)))
}

func (f *CFeature) parseCsvInts(s string) (indexes []int) {
	for _, v := range strings.Split(s, ",") {
		if i, err := strconv.Atoi(v); err == nil {
			indexes = append(indexes, i)
		}
	}
	return
}

var rxPqlKey = regexp.MustCompile(`^pql:([^:]+?):(\d+?):([^:]+?):(.*)$`)

func (f *CFeature) parsePqlKey(input string) (contextKey string, idx int, vType string, val *kvm.Value, err error) {
	if rxPqlKey.MatchString(input) {
		m := rxPqlKey.FindAllStringSubmatch(input, 1)
		parts := m[0][1:]
		if idx, err = strconv.Atoi(parts[1]); err != nil {
			err = fmt.Errorf("error parsing pql key index: %v", parts[2])
			return
		}
		contextKey = parts[0]
		vType = parts[2]
		val, err = kvm.NewValueFromTypeData(vType, parts[3])
	} else {
		err = fmt.Errorf("not pql key: %v", input)
	}
	return
}

var rxPqlValueStubsKey = regexp.MustCompile(`^pql:([^:]+?):(\d+?):(value|stubs)$`)

func (f *CFeature) parsePqlValueStubsKey(input string) (contextKey string, idx int, vs string, err error) {
	if rxPqlValueStubsKey.MatchString(input) {
		m := rxPqlValueStubsKey.FindAllStringSubmatch(input, 1)
		parts := m[0][1:]
		if idx, err = strconv.Atoi(parts[1]); err != nil {
			err = fmt.Errorf("error parsing pql value/stubs key index: %v", parts[2])
			return
		}
		contextKey = parts[0]
		vs = parts[2]
	} else {
		err = fmt.Errorf("not pql value/stubs key: %v", input)
	}
	return
}

func (f *CFeature) pqlGetDataValueKeyIndex(contextKey, shasum string) (dataKey string, idx int, err error) {
	dataKey = fmt.Sprintf(gPqlDataKeyFormat, contextKey, shasum)
	idx, err = f.leveldbGetIntVal([]byte(dataKey))
	return
}