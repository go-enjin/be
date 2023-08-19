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

type Views []*View

type View struct {
	Index int
	Key   string
	Url   string
	Label string

	Present bool

	Filters    Filters
	HasFilters bool

	Paginate string
	NextMore string

	FirstPage string
	PrevPage  string
	NextPage  string
	LastPage  string

	PageIndex  int
	PageNumber int
	TotalPages int
	NumPerPage int

	SearchAction string
}

func makeView(idx int, key, label string, filters Filters) (view *View) {
	view = &View{
		Index:      idx,
		Key:        key,
		Label:      label,
		Filters:    filters,
		HasFilters: len(filters) > 0,
		Present:    false,
	}
	return
}

func (v Views) HasFilters() (has bool) {
	for _, view := range v {
		if has = view.HasFilters; has {
			return
		}
	}
	return
}

func (v Views) HasFilterPresent() (has bool) {
	for _, view := range v {
		for _, group := range view.Filters {
			for _, filter := range group {
				if has = filter.Present; has {
					return
				}
			}
		}
	}
	return
}

func (v *View) HasFilterPresent() (has bool) {
	for _, group := range v.Filters {
		for _, filter := range group {
			if has = filter.Present; has {
				return
			}
		}
	}
	return
}

func (v *View) GroupHasFilterPresent(idx int) (has bool) {
	if idx >= 0 && len(v.Filters) > idx {
		for _, filter := range v.Filters[idx] {
			if has = filter.Present; has {
				return
			}
		}
	}
	return
}