// Copyright (c) 2024  The Go-Enjin Authors
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

package gorm

import (
	"sync"

	"github.com/go-corelibs/maps"
	"github.com/go-corelibs/slices"
)

var (
	gDialects = &dbDialects{
		d: make(map[DialectName]*dbDialect),
		m: &sync.RWMutex{},
	}
)

type dbDialects struct {
	d map[DialectName]*dbDialect

	m *sync.RWMutex
}

func (d *dbDialects) supports() (list string) {
	d.m.RLock()
	defer d.m.RUnlock()
	for idx, name := range maps.SortedKeys(d.d) {
		if idx > 0 {
			list += ", "
		}
		list += string(name)
	}
	return
}

func (d *dbDialects) has(name DialectName) (present bool) {
	d.m.RLock()
	defer d.m.RUnlock()
	if _, present = d.d[name]; !present {
		for _, dialect := range d.d {
			if present = slices.Present(name, dialect.alias...); present {
				return
			}
		}
	}
	return
}

func (d *dbDialects) get(name DialectName) (dialect *dbDialect) {
	d.m.RLock()
	defer d.m.RUnlock()
	if found, ok := d.d[name]; ok {
		dialect = found
		return
	}
	for _, known := range d.d {
		if ok := slices.Present(name, known.alias...); ok {
			dialect = known
			return
		}
	}
	return
}

func (d *dbDialects) set(dialect *dbDialect) {
	d.m.Lock()
	defer d.m.Unlock()
	d.d[dialect.name] = dialect
	return
}
