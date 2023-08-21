//go:build !exclude_pages_formats && !exclude_pages_format_njn

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
	"strings"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
)

type Filters [][]Filter

func makeFilters(data map[string]interface{}) (filters Filters) {
	if indexFilters, check := data["index-filters"].([]interface{}); check {
		for idx, item := range indexFilters {
			switch t := item.(type) {
			case map[string]interface{}:
				filters = append(filters, []Filter{makeFilter(idx, 0, t)})
			case []interface{}:
				var subFilters []Filter
				for jdx, subItem := range t {
					switch tt := subItem.(type) {
					case map[string]interface{}:
						subFilters = append(subFilters, makeFilter(idx, jdx, tt))
					default:
						log.ErrorF("invalid filter data structure: %T", tt)
					}
				}
				filters = append(filters, subFilters)
			}
		}
	}
	return
}

func (f Filters) Copy() (duplicate Filters) {
	for _, group := range f {
		var copied []Filter
		for _, filter := range group {
			copied = append(copied, filter.Copy())
		}
		duplicate = append(duplicate, copied)
	}
	return
}

func (f Filters) SetPresent(key string) (updated bool) {
	var filter Filter
	if filter, updated = f.Find(key); updated {
		f[filter.Group][filter.Position].Present = true
		// log.WarnF("set present: %d/%d - %v", filter.Group, filter.Position, key)
	}
	return
}

func (f Filters) Update(filter Filter) (updated bool) {
	for idx, group := range f {
		for jdx, fltr := range group {
			if updated = fltr.Key == filter.Key; updated {
				f[idx][jdx] = filter
				break
			}
		}
	}
	return
}

func (f Filters) Find(key string) (filter Filter, ok bool) {
	for idx, group := range f {
		for jdx, fltr := range group {
			fltr.Group = idx
			fltr.Position = jdx
			f[idx][jdx] = fltr
			if ok = fltr.Key == key; ok {
				filter = fltr
				return
			}
		}
	}
	return
}

func (f Filters) UpdateUrls(tag, reqArgvPath string, argv ...string) {
	var filterLinkGroup []int
	var filterLinkChain []string
	for idx, group := range f {
		for _, filter := range group {
			if filter.Present {
				filterLinkGroup = append(filterLinkGroup, idx)
				filterLinkChain = append(filterLinkChain, filter.Key)
				break // one per group
			}
		}
	}
	for idx, group := range f {
		for jdx, filter := range group {
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
			f[idx][jdx].Url = reqArgvPath

			prefix := tag
			if len(argv) > 0 {
				prefix += "," + strings.Join(argv, ",")
			}
			if len(chain) > 0 {
				prefix += "," + strings.Join(chain, ",")
			}

			if removed {
				f[idx][jdx].Url += "/:" + prefix
			} else {
				f[idx][jdx].Url += "/:" + prefix + "," + filter.Key
			}

		}
	}
	return
}

func (f Filters) FilterPages(pages []feature.Page) (filtered []feature.Page) {
	var present []Filter
	for _, group := range f {
		for _, filter := range group {
			if filter.Present {
				present = append(present, filter)
				break // move to next group, only one filter present per group
			}
		}
	}
	needed := len(present)
	for _, pg := range pages {
		count := 0
		for _, filter := range present {
			if matched, e := pg.MatchQL(filter.Query); e != nil {
				log.ErrorF("error matching filter: %v - %v", filter.Query, e)
			} else if matched {
				count += 1
			}
		}
		if count == needed {
			filtered = append(filtered, pg)
		}
	}
	return
}