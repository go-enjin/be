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

package page

import (
	"sort"
	"time"

	"github.com/go-enjin/be/pkg/log"
)

func SortPages(filtered []*Page, orderBy, sortDir string) (sorted []*Page) {
	if orderBy == "" {
		orderBy = "Title"
	}
	switch sortDir {
	case "ASC", "DSC":
	default:
		sortDir = "ASC"
	}
	sort.Slice(filtered, func(i, j int) (less bool) {
		var a, b interface{}
		a = filtered[i].Context.Get(orderBy)
		b = filtered[j].Context.Get(orderBy)
		if (a == nil && b == nil) || (a != nil && b == nil) {
			less = false
		} else if a == nil && b != nil {
			less = true
		} else {

			switch ta := a.(type) {
			case string:
				if tb, ok := b.(string); ok {
					less = ta < tb
				}
			case int:
				if tb, ok := b.(int); ok {
					less = ta < tb
				}
			case time.Time:
				if tb, ok := b.(time.Time); ok {
					less = ta.Before(tb)
				}
			default:
				log.ErrorF("unsupported sort key type: %T", a)
			}

		}
		if sortDir != "ASC" {
			less = !less
		}
		return
	})
	sorted = filtered
	return
}