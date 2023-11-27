// Copyright (c) 2023  The Go-Enjin Authors
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

package editor

import (
	"sort"

	"github.com/maruel/natural"

	"github.com/go-enjin/be/pkg/slices"
)

type Actions []*Action

func (list Actions) Len() (size int) {
	return len(list)
}

func (list Actions) Has(key string) (present bool) {
	for _, action := range list {
		if present = action.Key == key; present {
			return
		}
	}
	return
}

func (list Actions) Prune(keys ...string) (pruned Actions) {
	for _, action := range list {
		if !slices.Within(action.Key, keys) {
			pruned = append(pruned, action)
		}
	}
	return
}

func (list Actions) Sort() (sorted Actions) {
	sorted = append(sorted, list...)
	sort.Slice(sorted, func(i, j int) (less bool) {
		a, b := sorted[i], sorted[j]
		if less = a.Order < b.Order; less {
		} else if a.Order == b.Order {
			less = natural.Less(a.Name, b.Name)
		}
		return
	})
	return
}
