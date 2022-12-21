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

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/hash/sha"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps/kvm"
	"github.com/go-enjin/be/pkg/page"
	"github.com/go-enjin/be/pkg/pagecache"
	"github.com/go-enjin/be/pkg/pageql/matcher"
	"github.com/go-enjin/be/pkg/pageql/selector"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

func (f *CFeature) AddToIndex(stub *pagecache.Stub, p *page.Page) (err error) {
	return
}

func (f *CFeature) RemoveFromIndex(tag language.Tag, file, shasum string) {
	return
}

func (f *CFeature) PerformQuery(input string) (stubs []*pagecache.Stub, err error) {
	stubs, err = matcher.NewProcess(input, f.enjin)
	// log.WarnF("matcher processed: %v - %v", len(stubs), input)
	return
}

func (f *CFeature) PerformSelect(input string) (selected map[string]interface{}, err error) {
	selected, err = selector.NewProcess(input, f.enjin)
	// log.WarnF("selector processed: %v - %v", len(selected), input)
	return
}

func (f *CFeature) pqlAddToNextIndex(stubIdx int, p *page.Page) (err error) {
	stubIdxStr := strconv.Itoa(stubIdx)
	for k, v := range p.Context {
		if beStrings.StringInStrings(strings.ToLower(k), "content", "frontmatter") {
			continue
		}
		value := kvm.NewValue(v)
		var data []byte
		if data, err = value.MarshalBinary(); err != nil {
			log.Error(err)
			return
		}
		thisDataShasum, _ := sha.DataHash10(data)

		var thisDataKey string
		var thisValueKeyIdx, existingValueIdx int

		if thisDataKey, existingValueIdx, err = f.pqlGetDataValueKeyIndex(k, thisDataShasum); err != nil {
			log.Error(err)
			return
		} else if existingValueIdx >= 0 {
			thisValueKeyIdx = existingValueIdx
		} else {
			if thisValueKeyIdx, err = f.pqlGetNextValueKeyIndex(k); err != nil {
				log.Error(err)
				return
			}
		}

		thisStubsKey := fmt.Sprintf(gPqlContextKeyStubsFormat, k, thisValueKeyIdx)
		thisValueKey := fmt.Sprintf(gPqlContextKeyValueFormat, k, thisValueKeyIdx)

		var sids []string
		sidb, _ := f.leveldb.Get([]byte(thisStubsKey), nil)
		if sidv := string(sidb); sidv != "" {
			sids = strings.Split(sidv, ",")
			if !beStrings.StringInSlices(stubIdxStr, sids) {
				sids = append(sids, stubIdxStr)
			}
		} else {
			sids = append(sids, stubIdxStr)
		}

		if err = f.leveldb.Put([]byte(thisStubsKey), []byte(strings.Join(sids, ",")), nil); err != nil {
			log.Error(err)
		}
		if err = f.leveldb.Put([]byte(thisValueKey), data, nil); err != nil {
			log.Error(err)
		}
		if err = f.leveldb.Put([]byte(thisDataKey), []byte(strconv.Itoa(thisValueKeyIdx)), nil); err != nil {
			log.Error(err)
		}
		if err = f.pqlConsumeNextValueKeyIndex(k); err != nil {
			log.Error(err)
		}
		// log.WarnF("put: %v - %v", k, thisStubsKey)
	}
	return
}