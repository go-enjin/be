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

package index

import (
	"fmt"
	"html/template"
	"sort"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/pageql"
)

const (
	Tag feature.Tag = "NjnIndexBlock"
)

var (
	_ Block     = (*CBlock)(nil)
	_ MakeBlock = (*CBlock)(nil)
)

type Block interface {
	feature.EnjinBlock
}

type MakeBlock interface {
	Make() Block
}

type CBlock struct {
	feature.CEnjinBlock

	enjin feature.Internals
}

func New() (field MakeBlock) {
	f := new(CBlock)
	f.Init(f)
	return f
}

func (f *CBlock) Tag() feature.Tag {
	return Tag
}

func (f *CBlock) Init(this interface{}) {
	f.CEnjinBlock.Init(this)
}

func (f *CBlock) Make() Block {
	return f
}

func (f *CBlock) Setup(enjin feature.Internals) {
	f.enjin = enjin
}

func (f *CBlock) NjnClass() (tagClass feature.NjnClass) {
	tagClass = feature.InlineNjnClass
	return
}

func (f *CBlock) NjnBlockType() (name string) {
	name = "index"
	return
}

func (f *CBlock) PrepareBlock(re feature.EnjinRenderer, blockType string, data map[string]interface{}) (block map[string]interface{}, err error) {
	if blockType != "index" {
		err = fmt.Errorf("%v does not implement %v block type", f.Tag(), blockType)
		return
	}

	var blockDataContent map[string]interface{}
	if blockDataContent, err = re.PrepareGenericBlockData(data["content"]); err != nil {
		return
	}
	delete(blockDataContent, "section")

	block = re.PrepareGenericBlock("content", data)
	block["Type"] = "index"

	var indexViews []string
	if views, ok := data["index-views"].(string); ok {
		for _, view := range strings.Split(views, ",") {
			indexViews = append(indexViews, strcase.ToKebab(strings.TrimSpace(view)))
		}
	} else {
		err = fmt.Errorf("index blocks require an index-view property set")
		return
	}
	block["Views"] = indexViews

	var ok bool
	var orderBy, sortDir string
	if orderBy, ok = data["index-order-by"].(string); ok {
		orderBy = strcase.ToCamel(orderBy)
	} else {
		orderBy = "Title"
	}
	if sortDir, ok = data["index-sort-dir"].(string); ok {
		switch strings.ToUpper(sortDir) {
		case "ASC", "DSC":
		default:
			log.ErrorF("unknown sort direction: %v", sortDir)
			sortDir = "ASC"
		}
	} else {
		sortDir = "ASC"
	}

	var maxResults int
	if v, found := data["index-max-results"]; found {
		switch t := v.(type) {
		case string:
			if maxResults, err = strconv.Atoi(t); err != nil {
				err = fmt.Errorf("error parsing index-max-results integer: %v", err)
				return
			}
		case float64:
			maxResults = int(t)
		default:
			err = fmt.Errorf("unsupported index-max-results value type: %T %v", v, v)
			return
		}
	}

	var query string
	if query, ok = data["index-query"].(string); ok {
		if err = pageql.Validate(query); err != nil {
			err = fmt.Errorf("query error: %v - %v", query, err)
			return
		}

		found := f.enjin.MatchQL(query)

		sortFn := func(i, j int) (less bool) {
			var a, b interface{}
			a = found[i].Context.Get(orderBy)
			b = found[j].Context.Get(orderBy)
			if (a == nil && b == nil) || (a != nil && b == nil) {
				less = false
			} else if a == nil && b != nil {
				less = true
			} else {
				var ta, tb string
				if ta, ok = a.(string); ok {
					if tb, ok = b.(string); ok {
						less = ta < tb

					}
				}
			}
			if sortDir != "ASC" {
				less = !less
			}
			return
		}

		sort.Slice(found, sortFn)

		log.DebugF("index block found %v (max=%v) pages with query: %v", len(found), maxResults, query)
		if maxResults > 0 && len(found) > maxResults {
			found = found[:maxResults]
		}

		block["Results"] = found

	} else {
		err = fmt.Errorf("index blocks require an index-query property set")
		return
	}

	if heading, ok := re.PrepareBlockHeader(blockDataContent); ok {
		block["Heading"] = heading
	}

	if footer, ok := re.PrepareBlockFooter(blockDataContent); ok {
		block["Footer"] = footer
	}

	return
}

func (f *CBlock) RenderPreparedBlock(re feature.EnjinRenderer, block map[string]interface{}) (html template.HTML, err error) {
	html, err = re.RenderNjnTemplate("block/index", block)
	return
}

func (f *CBlock) ProcessBlock(re feature.EnjinRenderer, blockType string, data map[string]interface{}) (html template.HTML, err error) {
	if block, e := f.PrepareBlock(re, blockType, data); e != nil {
		err = e
		return
	} else {
		html, err = f.RenderPreparedBlock(re, block)
	}
	return
}