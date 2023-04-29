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
	"context"
	"fmt"
	"net/http"

	"github.com/go-enjin/be/pkg/database"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

func (e *Enjin) DB(tag string) (db interface{}, err error) {
	for _, f := range e.Features() {
		if fdb, ok := f.(feature.Database); ok {
			if beStrings.StringInSlices(tag, fdb.ListDB()) {
				db, err = fdb.DB(tag)
				return
			}
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
	for _, f := range e.Features() {
		if fdb, ok := f.(feature.Database); ok {
			if fTag == fdb.Tag() {
				db, err = fdb.DB(tag)
				return
			}
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

func (e *Enjin) dbMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, f := range e.Features() {
			if fdb, ok := f.(feature.Database); ok {
				for _, tag := range fdb.ListDB() {
					db := fdb.MustDB(tag)
					tagKey, specificKey := database.MakeRequestDbSpecificKeys(fdb.Tag(), tag)
					if v := r.Context().Value(tagKey); v == nil {
						// first tag provided, no overwriting
						r = r.Clone(context.WithValue(r.Context(), tagKey, db))
						if tag == "default" {
							// log.DebugF("adding DefaultDb to request, from feature:  %v", fdb.Tag())
							r = r.Clone(context.WithValue(r.Context(), database.RequestDefaultDbKey, db))
						}
					}
					r = r.Clone(context.WithValue(r.Context(), specificKey, db))
					// log.DebugF("adding specific DB to request: %v", specificKey)
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}