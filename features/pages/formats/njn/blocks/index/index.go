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
	"github.com/go-enjin/be/pkg/page"
	"github.com/go-enjin/be/pkg/pageql"
	"github.com/go-enjin/be/pkg/types/site"
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

	reqArgv := re.RequestArgv()
	// log.WarnF("reqArgv=%#v", reqArgv.String())

	var blockDataContent map[string]interface{}
	if blockDataContent, err = re.PrepareGenericBlockData(data["content"]); err != nil {
		return
	}
	delete(blockDataContent, "section")

	block = re.PrepareGenericBlock("index", data)

	var ok bool
	var tag string
	if tag, ok = block["Tag"].(string); !ok {
		err = fmt.Errorf("index block missing data-block-tag")
		return
	}

	var pgntn string
	if pgntn, ok = data["index-pagination"].(string); ok {
		pgntn = strings.ToLower(pgntn)
		switch pgntn {
		case "none", "more", "page":
		default:
			err = fmt.Errorf("invalid index block pagination value: %v", pgntn)
			return
		}
	}
	block["Pagination"] = pgntn

	numPerPage, pageNumber := 10, 0

	if v, check := data["index-num-per-page"].(float64); check {
		numPerPage = int(v)
	}

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

	var filters Filters
	if indexFilters, check := data["index-filters"].([]interface{}); check {
		for _, item := range indexFilters {
			switch t := item.(type) {
			case map[string]interface{}:
				filters = append(filters, []Filter{MakeFilterFrom(t)})
			case []interface{}:
				var subFilters []Filter
				for _, subItem := range t {
					switch tt := subItem.(type) {
					case map[string]interface{}:
						subFilters = append(subFilters, MakeFilterFrom(tt))
					}
				}
				filters = append(filters, subFilters)
			}
		}
	}

	for idx, pieces := range reqArgv.Argv {
		if pieces[0] != "" && pieces[0] == tag {
			re.RequestContext().SetSpecific(site.RequestArgvConsumedKey, true)
			if reqArgv.NumPerPage > -1 {
				numPerPage = reqArgv.NumPerPage
			}
			reqArgv.NumPerPage = numPerPage
			if reqArgv.PageNumber > -1 {
				pageNumber = reqArgv.PageNumber
			}
			reqArgv.PageNumber = pageNumber
			var fixArgs []string
			for _, piece := range pieces[1:] {
				foundPiece := false
				for jdx, filterSet := range filters {
					for kdx, filter := range filterSet {
						if filter.Key == piece {
							foundPiece = true
							filters[jdx][kdx].Present = true
							break
						}
					}
					if foundPiece {
						break
					}
				}
				if foundPiece {
					fixArgs = append(fixArgs, piece)
				} else {
					break
				}
			}
			if len(fixArgs) != len(pieces[1:]) {
				reqArgv.Argv[idx] = append([]string{pieces[0]}, fixArgs...)
				re.RequestContext().SetSpecific(site.RequestRedirectKey, reqArgv.String())
			}
		}
	}

	var filterLinkGroup []int
	var filterLinkChain []string
	for idx, group := range filters {
		for _, filter := range group {
			if filter.Present {
				filterLinkGroup = append(filterLinkGroup, idx)
				filterLinkChain = append(filterLinkChain, filter.Key)
				break // one per group
			}
		}
	}

	for idx, group := range filters {
		for jdx, filter := range group {
			filters[idx][jdx].Url = reqArgv.Path
			var chain []string
			var removed bool
			for cdx, chained := range filterLinkChain {
				gdx := filterLinkGroup[cdx]
				if chained == filter.Key {
					removed = true
				} else if gdx != idx {
					chain = append(chain, chained)
				}
			}
			if len(chain) == 0 {
				if removed {
					filters[idx][jdx].Url += "/:" + tag
				} else {
					filters[idx][jdx].Url += "/:" + tag + "," + filter.Key
				}
			} else {
				if removed {
					filters[idx][jdx].Url += "/:" + tag + "," + strings.Join(chain, ",")
				} else {
					filters[idx][jdx].Url += "/:" + tag + "," + strings.Join(chain, ",") + "," + filter.Key
				}
			}
		}
	}

	if len(filters) > 0 {
		block["Filters"] = filters
		// log.DebugF("index-filters: %#v", filters)
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

		totalFound := len(found)

		runFilter := func(filter Filter) {
			var modified []*page.Page
			for _, pg := range found {
				if matched, e := pg.MatchQL(filter.Query); e != nil {
					log.ErrorF("error parsing filter query: %v - %v", filter.Query, e)
				} else if matched {
					// log.DebugF("%v - matches - %v", pg.Url, filter.Query)
					modified = append(modified, pg)
				}
			}
			found = modified
		}
		for _, filterSet := range filters {
			for _, filter := range filterSet {
				if filter.Present {
					runFilter(filter)
				}
			}
		}

		if maxResults > 0 && totalFound > maxResults {
			found = found[:maxResults]
		}

		totalFiltered := len(found)

		totalPages := totalFiltered / numPerPage
		if pageNumber > totalPages {
			reqArgv.PageNumber = totalPages - 1
			pageNumber = reqArgv.PageNumber
			re.RequestContext().SetSpecific(site.RequestRedirectKey, reqArgv.String())
		}

		if numPerPage > 0 {
			start := pageNumber * numPerPage
			end := start + numPerPage
			if start < len(found) {
				if end < totalFiltered-1 {
					found = found[start:end]
				} else {
					found = found[start:]
				}
			}
		}

		block["Results"] = found

		// log.DebugF("index block found %v (total=%v, max=%v) pages with query: %v", totalFiltered, totalFound, maxResults, query)
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