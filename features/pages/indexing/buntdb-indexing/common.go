//go:build buntdb_indexing || all

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

	"github.com/tidwall/buntdb"

	"github.com/go-enjin/be/pkg/maps/kvm"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

func (f *CFeature) getIndexByShasum(shasum string) (idx int, err error) {
	idx = -1
	if err = f.buntdb.View(func(tx *buntdb.Tx) (err error) {
		var value string
		if value, err = tx.Get(fmt.Sprintf("stub:%v:index", shasum)); err != nil {
			err = fmt.Errorf("error getting index by shasum: %v - %v", shasum, err)
			// panic("check")
			return
		} else if value != "" {
			if idx, err = strconv.Atoi(value); err != nil {
				err = fmt.Errorf("error converting index to int: %v - %v", value, err)
				return
			}
		}
		return
	}); err != nil {
		err = fmt.Errorf("error viewing index by shasum: %v", err)
		return
	}
	return
}

func (f *CFeature) getShasumByIndex(idx int) (shasum string, err error) {
	if err = f.buntdb.View(func(tx *buntdb.Tx) (err error) {
		if shasum, err = tx.Get(fmt.Sprintf("stub:%v:shasum", idx)); err == nil && shasum == "" {
			err = fmt.Errorf("stub shasum not found at index: %v", idx)
		}
		return
	}); err != nil {
		err = fmt.Errorf("error getting shasum by index: %d - %v", idx, err)
		return
	}
	return
}

func (f *CFeature) getShasumsFromIndexes(indexes []int) (shasums []string, err error) {
	if err = f.buntdb.View(func(tx *buntdb.Tx) (err error) {
		for _, idx := range indexes {
			var shasum string
			if shasum, err = tx.Get(fmt.Sprintf("stub:%v:shasum", idx)); err == nil && shasum == "" {
				err = fmt.Errorf("stub shasum not found at index: %v", idx)
				return
			} else if !beStrings.StringInSlices(shasum, shasums) {
				shasums = append(shasums, shasum)
			}
		}
		return
	}); err != nil {
		err = fmt.Errorf("error getting shasums frmo indexes: %v - %v", err, indexes)
		return
	}
	return
}

func (f *CFeature) getShasumsFromIndexesTx(indexes []int, tx *buntdb.Tx) (shasums []string, err error) {
	for _, idx := range indexes {
		var shasum string
		if shasum, err = tx.Get(fmt.Sprintf("stub:%v:shasum", idx)); err == nil && shasum == "" {
			err = fmt.Errorf("stub shasum not found at index: %v", idx)
			return
		} else if !beStrings.StringInSlices(shasum, shasums) {
			shasums = append(shasums, shasum)
		}
	}
	return
}

func (f *CFeature) getNextStubIndex() (idx int, err error) {
	stubIndexKey := fmt.Sprintf("stub:next")
	err = f.buntdb.View(func(tx *buntdb.Tx) (err error) {
		var v string
		if v, _ = tx.Get(stubIndexKey); v != "" {
			if idx, err = strconv.Atoi(v); err != nil {
				idx = -1
				return
			}
		}
		return
	})
	return
}

func (f *CFeature) consumeNextStubIndex() (err error) {
	stubIndexKey := fmt.Sprintf("stub:next")
	err = f.buntdb.Update(func(tx *buntdb.Tx) (err error) {
		var idx int
		var v string
		if v, _ = tx.Get(stubIndexKey); v != "" {
			if idx, err = strconv.Atoi(v); err != nil {
				idx = -1
				return
			}
		}
		_, _, err = tx.Set(stubIndexKey, strconv.Itoa(idx+1), nil)
		return
	})
	return
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

func (f *CFeature) pqlGetNextValueKeyIndex(contextKey string) (idx int, err error) {
	pqlNextIndexKey := fmt.Sprintf("pql:next:%v", contextKey)
	if err = f.buntdb.View(func(tx *buntdb.Tx) (err error) {
		var v string
		if v, err = tx.Get(pqlNextIndexKey); err == nil {
			idx, err = strconv.Atoi(v)
		}
		return
	}); err != nil && strings.Contains(err.Error(), "not found") {
		idx = 0
		err = nil
	}
	return
}

func (f *CFeature) pqlConsumeNextValueKeyIndex(contextKey string) (err error) {
	pqlNextIndexKey := fmt.Sprintf("pql:next:%v", contextKey)
	err = f.buntdb.View(func(tx *buntdb.Tx) (err error) {
		var next int
		sidx, _ := tx.Get(pqlNextIndexKey)
		if next, err = strconv.Atoi(sidx); err != nil {
			return
		}
		_, _, err = tx.Set(pqlNextIndexKey, strconv.Itoa(next+1), nil)
		return
	})
	return
}

func (f *CFeature) pqlConsumeNextValueKeyIndexTx(contextKey string, tx *buntdb.Tx) (err error) {
	pqlNextIndexKey := fmt.Sprintf("pql:next:%v", contextKey)
	var next int
	if sidx, _ := tx.Get(pqlNextIndexKey); sidx != "" {
		if next, err = strconv.Atoi(sidx); err != nil {
			return
		}
	}
	_, _, err = tx.Set(pqlNextIndexKey, strconv.Itoa(next+1), nil)
	return
}

func (f *CFeature) pqlGetDataValueKeyIndex(contextKey, shasum string) (dataKey string, idx int, err error) {
	idx = -1
	dataKey = fmt.Sprintf("pql:data:%v:%v", contextKey, shasum)
	err = f.buntdb.View(func(tx *buntdb.Tx) (err error) {
		if sidx, _ := tx.Get(dataKey); sidx != "" {
			if idx, err = strconv.Atoi(sidx); err != nil {
				idx = -1
				return
			}
		}
		return
	})
	return
}

func (f *CFeature) pqlSetDataKeyIndex(contextKey, shasum string, idx int) (err error) {
	pqlDataKey := fmt.Sprintf("pql:data:%v:%v", contextKey, shasum)
	err = f.buntdb.Update(func(tx *buntdb.Tx) (err error) {
		_, _, err = tx.Set(pqlDataKey, strconv.Itoa(idx), nil)
		return
	})
	return
}

func (f *CFeature) getContextKeyValues(contextKey string) (values []interface{}, err error) {
	key := fmt.Sprintf("pql:%v:*:values", contextKey)
	if err = f.buntdb.View(func(tx *buntdb.Tx) (err error) {
		if err = tx.AscendKeys(key, func(k, v string) (ok bool) {
			var thisContextKey string
			if thisContextKey, _, _, err = f.parsePqlValueStubsKey(k); err != nil {
				return
			}
			if thisContextKey != contextKey {
				ok = true
				return
			}

			var vi kvm.Value
			if err = vi.UnmarshalBinary([]byte(v)); err != nil {
				return
			}

			switch tv := vi.Get().(type) {
			case []string:
				for _, vtv := range tv {
					values = append(values, vtv)
				}
			case []interface{}:
				for _, vtv := range tv {
					values = append(values, vtv)
				}
			default:
				values = append(values, tv)
			}
			return
		}); err != nil {
			return
		}
		// log.DebugF("values:\n%v", values)
		return
	}); err != nil {
		err = fmt.Errorf("error getting context key values: %v", err)
		return
	}
	return
}