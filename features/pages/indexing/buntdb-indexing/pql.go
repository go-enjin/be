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
	"fmt"
	"strconv"
	"strings"

	"github.com/go-enjin/golang-org-x-text/language"
	"github.com/tidwall/buntdb"

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
	log.WarnF("matcher processed: %v - %v", len(stubs), input)
	return
}

func (f *CFeature) PerformSelect(input string) (selected map[string]interface{}, err error) {
	selected, err = selector.NewProcess(input, f.enjin)
	log.WarnF("selector processed: %v - %v", len(selected), input)
	return
}

func (f *CFeature) pqlAddToNextIndex(stubIdx int, p *page.Page) (err error) {
	if f.cliBatch += 1; f.cliBatch >= 500 {
		if e := f.buntdb.Shrink(); e != nil {
			log.WarnF("error shrinking buntdb: %v", e)
		} else {
			log.DebugF("shrunk buntdb")
		}
		f.cliBatch = 1
	}

	stubIdxStr := strconv.Itoa(stubIdx)
	for k, v := range p.Context {
		if beStrings.StringInStrings(strings.ToLower(k), "content", "frontmatter") {
			continue
		}
		value := kvm.NewValue(v)
		var data []byte
		if data, err = value.MarshalBinary(); err != nil {
			return
		}
		thisDataShasum, _ := sha.DataHash10(data)

		var thisDataKey string
		var thisValueKeyIdx, existingValueIdx int

		if thisDataKey, existingValueIdx, err = f.pqlGetDataValueKeyIndex(k, thisDataShasum); err != nil {
			return
		} else if existingValueIdx >= 0 {
			thisValueKeyIdx = existingValueIdx
		} else {
			if thisValueKeyIdx, err = f.pqlGetNextValueKeyIndex(k); err != nil {
				return
			}
		}

		thisValueKeyPrefix := fmt.Sprintf("pql:%v:%d:", k, thisValueKeyIdx)
		thisStubsKey := thisValueKeyPrefix + "stubs"
		thisValueKey := thisValueKeyPrefix + "value"

		if err = f.buntdb.Update(func(tx *buntdb.Tx) (err error) {
			sidv, _ := tx.Get(thisStubsKey)
			var sids []string
			if sidv != "" {
				sids = strings.Split(sidv, ",")
				if !beStrings.StringInSlices(stubIdxStr, sids) {
					sids = append(sids, stubIdxStr)
				}
			} else {
				sids = append(sids, stubIdxStr)
			}
			if _, _, err = tx.Set(thisStubsKey, strings.Join(sids, ","), nil); err != nil {
				return
			}
			if _, _, err = tx.Set(thisValueKey, string(data), nil); err != nil {
				return
			}
			if _, _, err = tx.Set(thisDataKey, strconv.Itoa(thisValueKeyIdx), nil); err != nil {
				return
			}
			err = f.pqlConsumeNextValueKeyIndexTx(k, tx)
			return
		}); err != nil {
			err = fmt.Errorf("error updating pql index: %v", err)
			return
		}
	}
	return
}