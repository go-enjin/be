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

package feature

import (
	"time"

	"github.com/go-corelibs/context"
	"github.com/go-corelibs/enjinql"
)

// PageIndexReport is the mapping of features to time durations, returned by
// PageIndexingFeature types to report how long it took to perform the
// indexing process
type PageIndexReport map[Tag]time.Duration

type PageIndexFeature interface {
	Feature
	AddToIndex(stub *PageStub, p Page) (err error)
	RemoveFromIndex(stub *PageStub, p Page) (err error)
}

type PageIndexingFeature interface {
	PageIndexFeature

	AddToIndexWithReport(stub *PageStub, p Page) (report PageIndexReport, err error)
	RemoveFromIndexWithReport(stub *PageStub, p Page) (report PageIndexReport, err error)
}

type QueryIndexFeature interface {
	Feature

	// FindPageID returns in internal EnjinQL primary source identifier
	//
	// Note: this id is only relevant to this particular QueryIndexFeature
	FindPageID(shasum string) (id int64, ok bool)

	PerformQuery(format string, argv ...interface{}) (stubs PageStubs, err error)
	PerformLookup(format string, argv ...interface{}) (columns []string, results context.Contexts, err error)

	EQL() enjinql.EnjinQL
}

// QueryIndexSourceFeature is a feature providing one or more enjinql source
// configs and manages the indexing of these additional sources
type QueryIndexSourceFeature interface {
	Feature

	// AddSources returns a list of enjinql.SourceConfig instances
	AddSources() (sources enjinql.ConfigSources)

	// AddToSource is called when the enjin is indexing a page
	AddToSource(tx enjinql.SqlTX, sid int64, stub *PageStub, p Page) (err error)

	// RemoveFromSource is called when the enjin is de-indexing a page
	//
	// Note: id is the primary source id being removed
	RemoveFromSource(tx enjinql.SqlTX, sid int64, stub *PageStub, p Page) (err error)
}
