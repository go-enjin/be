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
	Index      int
	Key        string
	Url        string
	Label      string
	Filters    Filters
	HasFilters bool
	Present    bool

	Paginate string
	NextMore string

	FirstPage string
	PrevPage  string
	NextPage  string
	LastPage  string

	PageNumber int
	TotalPages int
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