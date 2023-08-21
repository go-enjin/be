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

type Filter struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	Query   string `json:"query"`
	Present bool
	Url     string

	Group    int
	Position int
}

func makeFilter(group, position int, v map[string]interface{}) (filter Filter) {
	filter.Key, _ = v["key"].(string)
	filter.Label, _ = v["label"].(string)
	filter.Query, _ = v["query"].(string)
	filter.Group = group
	filter.Position = position
	return
}

func (f Filter) Copy() (duplicate Filter) {
	duplicate = Filter{
		Key:      f.Key,
		Label:    f.Label,
		Query:    f.Query,
		Present:  f.Present,
		Url:      f.Url,
		Group:    f.Group,
		Position: f.Position,
	}
	return
}