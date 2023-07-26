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

package be

import (
	"fmt"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

func (e *Enjin) DB(tag string) (db interface{}, err error) {
	for _, fdb := range e.eb.fDatabases {
		if beStrings.StringInSlices(tag, fdb.ListDB()) {
			db, err = fdb.DB(tag)
			return
		}
	}
	err = fmt.Errorf("db feature not found by tag: %v", tag)
	return
}

func (e *Enjin) MustDB(tag string) (db interface{}) {
	var err error
	if db, err = e.DB(tag); err != nil {
		log.FatalDF(1, err.Error())
	}
	return
}

func (e *Enjin) SpecificDB(fTag feature.Tag, tag string) (db interface{}, err error) {
	for _, fdb := range e.eb.fDatabases {
		if fTag == fdb.Tag() {
			db, err = fdb.DB(tag)
			return
		}
	}
	err = fmt.Errorf("db feature not found: %v", fTag)
	return
}

func (e *Enjin) MustSpecificDB(fTag feature.Tag, tag string) (db interface{}) {
	var err error
	if db, err = e.SpecificDB(fTag, tag); err != nil {
		log.FatalDF(1, err.Error())
	}
	return
}