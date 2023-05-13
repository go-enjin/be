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

package indexing

import (
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/page"
	"github.com/go-enjin/be/pkg/page/matter"
)

type CacheEnjinFeature interface {
	NewCache(bucket string) (err error)
	Mounted(bucket, path string) (ok bool)
	Mount(bucket, mount, path string, mfs fs.FileSystem)
	Rebuild(bucket string) (ok bool)
	Lookup(bucket string, tag language.Tag, url string) (mount, path string, p *page.Page, err error)
	LookupTranslations(bucket, url string) (pgs []*page.Page)
	LookupRedirect(bucket, url string) (p *page.Page, ok bool)
	LookupPrefix(bucket, prefix string) (found []*page.Page)
	TotalCached(bucket string) (count uint64)
}

type SearchEnjinFeature interface {
	PrepareSearch(tag language.Tag, input string) (query string)
	PerformSearch(tag language.Tag, input string, size, pg int) (results *bleve.SearchResult, err error)
	AddToSearchIndex(stub *matter.PageStub, p *page.Page) (err error)
	RemoveFromSearchIndex(tag language.Tag, file, shasum string)
}

type PageIndexFeature interface {
	AddToIndex(stub *matter.PageStub, p *page.Page) (err error)
	RemoveFromIndex(tag language.Tag, file, shasum string)
}

type QueryIndexFeature interface {
	PerformQuery(input string) (stubs []*matter.PageStub, err error)
	PerformSelect(input string) (selected map[string]interface{}, err error)
}

type SearchDocumentMapperFeature interface {
	SearchDocumentMapping(tag language.Tag) (doctype string, dm *mapping.DocumentMapping)
	AddSearchDocumentMapping(tag language.Tag, indexMapping *mapping.IndexMappingImpl)
}

type KeywordProvider interface {
	KnownKeywords() (keywords []string)
	KeywordStubs(keyword string) (stubs matter.PageStubs)
}

type PageContextProvider interface {
	YieldPageContextValues(key string) (values chan interface{})
	YieldPageContextValueStubs(key string) (pairs chan *matter.ValueStubPair)
}