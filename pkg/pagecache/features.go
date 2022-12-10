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

package pagecache

import (
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/page"
)

type QueryEnjinFeature interface {
	PageIndexFeature
	PerformQuery(input string) (stubs []*Stub, err error)
	PerformSelect(input string) (selected map[string]interface{}, err error)
}

type SearchEnjinFeature interface {
	PrepareSearch(tag language.Tag, input string) (query string)
	PerformSearch(tag language.Tag, input string, size, pg int) (results *bleve.SearchResult, err error)
	AddToSearchIndex(stub *Stub, p *page.Page) (err error)
	RemoveFromSearchIndex(tag language.Tag, file, shasum string)
}

type PageIndexFeature interface {
	AddToIndex(stub *Stub, p *page.Page) (err error)
	RemoveFromIndex(tag language.Tag, file, shasum string)
}

type SearchDocumentMapperFeature interface {
	SearchDocumentMapping(tag language.Tag) (doctype string, dm *mapping.DocumentMapping)
	AddSearchDocumentMapping(tag language.Tag, indexMapping *mapping.IndexMappingImpl)
}