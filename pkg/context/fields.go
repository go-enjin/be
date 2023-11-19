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

package context

import (
	"sort"
	"strings"

	"github.com/maruel/natural"

	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/golang-org-x-text/message"
)

type Fields map[string]*Field

func (f Fields) Init(printer *message.Printer, parsers Parsers) {
	for _, field := range f {
		if field.Printer == nil {
			field.Printer = printer
		}
		if field.Format == "" {
			field.Format = "string"
		}
		if field.Parse == nil {
			if fn, ok := parsers[field.Format]; ok {
				field.Parse = fn
			} else {
				log.WarnDF(1, "field %q parser func not found: %q, falling back to StringParser", field.Key, field.Format)
				field.Parse = StringParser
			}
		}
	}
	return
}

func (f Fields) Lookup(key string) (field *Field, ok bool) {
	key = strings.TrimPrefix(key, ".matter")
	key = strings.TrimPrefix(key, ".")
	if field, ok = f[key]; ok {
		return
	} else if field, ok = f["."+key]; ok {
		return
	}
	return
}

func (f Fields) Len() (count int) {
	count = len(f)
	return
}

func (f Fields) Keys() (keys []string) {
	for k, _ := range f {
		keys = append(keys, k)
	}
	return
}

func (f Fields) SortedKeys() (keys []string) {
	keys = f.Keys()
	sort.SliceStable(keys, func(i, j int) (less bool) {
		a, b := f[keys[i]], f[keys[j]]
		if less = isFieldTabLess(a.Tab, b.Tab); less {
			// page tab levels first
		} else if less = isFieldCategoryLess(a.Category, b.Category); less {
			// important categories next
		} else if less = a.Weight < b.Weight; less {
			// weighted fields
		} else if a.Weight == b.Weight {
			// finally sorted natural by key
			less = natural.Less(a.Key, b.Key)
		}
		return
	})
	return
}

func (f Fields) TabFields(tab string) (fields Fields) {
	fields = Fields{}
	for _, v := range f {
		if v.Tab == tab {
			fields[v.Key] = v
		}
	}
	return
}

func (f Fields) TabCategoryFields(tab, category string, ignored ...string) (fields Fields) {
	omits := map[string]struct{}{}
	for _, ignore := range ignored {
		omits[ignore] = struct{}{}
	}
	fields = Fields{}
	for _, v := range f {
		if _, omit := omits[v.Key]; omit {
			continue
		}
		if v.Tab == tab && v.Category == category {
			fields[v.Key] = v
		}
	}
	if len(fields) == 0 {
		fields = nil
	}
	return
}

func (f Fields) TabNames() (names []string) {
	unique := map[string]struct{}{}
	for _, v := range f {
		unique[v.Tab] = struct{}{}
	}
	names = maps.Keys(unique)
	sort.SliceStable(names, func(i, j int) (less bool) {
		less = isFieldTabLess(names[i], names[j])
		return
	})
	return
}

func (f Fields) CategoryNames() (names []string) {
	unique := map[string]struct{}{}
	for _, v := range f {
		unique[v.Category] = struct{}{}
	}
	names = maps.Keys(unique)
	sort.SliceStable(names, func(i, j int) (less bool) {
		less = isFieldCategoryLess(names[i], names[j])
		return
	})
	return
}

func (f Fields) TabCategoryNames(tab string) (keys []string) {
	unique := map[string]struct{}{}
	for _, v := range f {
		if v.Tab == tab {
			unique[v.Category] = struct{}{}
		}
	}
	keys = maps.Keys(unique)
	sort.SliceStable(keys, func(i, j int) (less bool) {
		less = isFieldCategoryLess(keys[i], keys[j])
		return
	})
	return
}

func isFieldTabLess(a, b string) (less bool) {
	if a == "page" || b == "page" {
		// page tab first
		less = a == "page" && b != "page"
	} else if less = natural.Less(a, b); less {
		// other tabs sorted next
	}
	return
}

func isFieldCategoryLess(a, b string) (less bool) {
	if a == "important" || b == "important" {
		// important category first
		less = a == "important" && b != "important"
	} else if less = natural.Less(a, b); less {
		// other categories next
	}
	return
}