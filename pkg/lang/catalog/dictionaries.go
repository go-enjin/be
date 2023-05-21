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

package catalog

import (
	"github.com/go-enjin/golang-org-x-text/language"
	"github.com/go-enjin/golang-org-x-text/message/catalog"
)

var _ Dictionaries = (*dictionaries)(nil)

type Dictionaries interface {
	catalog.Dictionary

	Append(d *dictionary)
}

type dictionaries struct {
	tag  language.Tag
	list []*dictionary
}

func newDictionaries(tag language.Tag) (d *dictionaries) {
	d = &dictionaries{
		tag:  tag,
		list: make([]*dictionary, 0),
	}
	return
}

func (d *dictionaries) Append(given *dictionary) {
	d.list = append(d.list, given)
}

func (d *dictionaries) Lookup(key string) (data string, ok bool) {
	count := len(d.list)
	for i := count - 1; i >= 0; i-- {
		if data, ok = d.list[i].Lookup(key); ok {
			return
		}
	}
	return
}