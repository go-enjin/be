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

package eql

import (
	"fmt"
	"time"

	"github.com/iancoleman/strcase"

	"github.com/go-corelibs/enjinql"
	"github.com/go-enjin/be/pkg/feature"
)

func (f *CFeature) AddToIndex(stub *feature.PageStub, p feature.Page) (err error) {
	_, err = f.AddToIndexWithReport(stub, p)
	return
}

func (f *CFeature) AddToIndexWithReport(stub *feature.PageStub, p feature.Page) (report feature.PageIndexReport, err error) {

	// - check if this shasum has already been indexed
	// - add shasum,language,type,url,stub to primary source
	// - add redirections if any
	// - add included context keys
	// - add any included enjinql features

	report = make(feature.PageIndexReport)
	start := time.Now()
	defer func() {
		var delta time.Duration
		for key := range report {
			delta += report[key]
		}
		report["~"] = time.Now().Sub(start) - delta
	}()

	if _, ok := f.FindPageID(stub.Shasum); ok {
		// considered already indexed
		return
	}

	var stubData []byte
	if stubData, err = stub.Marshal(); err != nil {
		return
	}

	// acquire a transaction lock
	var tx enjinql.SqlTrunkTX
	if tx, err = f.eql.SqlBegin(); err != nil {
		return
	}
	// defer a rollback and release of the transaction lock
	defer tx.Rollback()

	var sid int64
	if sid, err = tx.Insert(
		enjinql.PageSource,
		stub.Shasum,
		stub.Language.String(),
		p.Type(),
		p.Archetype(),
		p.CreatedAt(),
		p.UpdatedAt(),
		p.Url(),
		string(stubData),
	); err != nil {
		return
	}

	// TODO: add redirections timing
	for _, redirection := range p.Redirections() {
		if redirection != "" && redirection[0] == '/' {
			if _, err = tx.Insert(enjinql.PageRedirectSource, sid, redirection); err != nil {
				err = fmt.Errorf("error adding to redirection source: %w", err)
				return
			}
		}
	}

	// TODO: add per?-context-key timing
	for _, key := range f.includeContextKeys {
		snake := strcase.ToSnake(key)
		value := p.Context().Get(strcase.ToCamel(snake))
		// TODO: figure out better nil value indexing
		if _, err = tx.Insert(snake, sid, value); err != nil {
			err = fmt.Errorf("error adding to %q source: %w", snake, err)
			return
		}
	}

	for _, other := range f.CSiteIncluding.Features {
		otherStart := time.Now()
		if err = other.AddToSource(tx.TX(), sid, stub, p); err != nil {
			err = fmt.Errorf("error adding to %q source: %w", other.Tag(), err)
			return
		}
		report[other.Tag()] += time.Now().Sub(otherStart)
	}

	// commit and release the transaction lock
	err = tx.Commit()
	return
}

func (f *CFeature) RemoveFromIndex(stub *feature.PageStub, p feature.Page) (err error) {
	_, err = f.RemoveFromIndexWithReport(stub, p)
	return
}

func (f *CFeature) RemoveFromIndexWithReport(stub *feature.PageStub, p feature.Page) (report feature.PageIndexReport, err error) {

	report = make(feature.PageIndexReport)
	start := time.Now()
	defer func() {
		var delta time.Duration
		for key := range report {
			delta += report[key]
		}
		report["+"] = time.Now().Sub(start) - delta
	}()

	// - check if this shasum has not been indexed
	// - get id for shasum
	// - remove from any included enjinql features
	// - remove from included context keys
	// - remove from redirections
	// - remove from primary source

	if sid, ok := f.FindPageID(stub.Shasum); ok {

		// acquire a transaction lock
		var tx enjinql.SqlTrunkTX
		if tx, err = f.eql.SqlBegin(); err != nil {
			return
		}
		// defer a rollback and release of the transaction lock
		defer tx.Rollback()

		for _, other := range f.CSiteIncluding.Features {
			otherStart := time.Now()
			if err = other.RemoveFromSource(tx.TX(), sid, stub, p); err != nil {
				err = fmt.Errorf("error removing from %q source: %w", other.Tag(), err)
				return
			}
			report[other.Tag()] += time.Now().Sub(otherStart)
		}

		for _, key := range f.includeContextKeys {
			snake := strcase.ToSnake(key)
			if _, err = tx.DeleteWhereEQ(snake, enjinql.PageSourceIdKey, sid); err != nil {
				err = fmt.Errorf("error removing from %q source: %w", snake, err)
				return
			}
		}

		if _, err = tx.DeleteWhereEQ(enjinql.PageRedirectSource, enjinql.PageSourceIdKey, sid); err != nil {
			return
		}

		if _, err = tx.Delete(enjinql.PageSource, sid); err != nil {
			return
		}

		// commit and release the transaction lock
		err = tx.Commit()
	}

	return
}
