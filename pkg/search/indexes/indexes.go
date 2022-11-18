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

package indexes

import (
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/search"
)

func NewFeaturesIndex(e feature.Internals) (index map[language.Tag]bleve.Index, all bleve.Index, err error) {
	locales := e.SiteLocales()
	index = make(map[language.Tag]bleve.Index)
	for _, tag := range locales {
		index[tag], err = NewMemOnlyIndexWithInternals(tag, e)
	}

	for _, f := range e.Features() {
		if s, ok := f.(feature.Searchable); ok {
			log.DebugF("updating searchable feature: %v", f.Tag())
			for _, tag := range locales {
				if ee := s.UpdateSearch(tag, index[tag]); ee != nil {
					err = ee
					index = nil
					return
				}
			}
		}
	}

	var list []bleve.Index
	for _, idx := range index {
		list = append(list, idx)
	}
	all = bleve.NewIndexAlias(list...)
	return
}

func NewIndexMapping(tag language.Tag) (indexMapping *mapping.IndexMappingImpl) {
	indexMapping = bleve.NewIndexMapping()
	_, indexMapping.DefaultMapping = search.NewDocumentMapping(tag)
	indexMapping.AddDocumentMapping("document", indexMapping.DefaultMapping)
	return
}

func NewMemOnlyIndexWithInternals(tag language.Tag, e feature.Internals) (index bleve.Index, err error) {
	indexMapping := NewIndexMapping(tag)

	for _, f := range e.Features() {
		if m, ok := f.(feature.SearchDocumentMapper); ok {
			log.DebugF("adding %v search document mapping", f.Tag())
			m.AddSearchDocumentMapping(tag, indexMapping)
		}
	}

	if index, err = bleve.NewMemOnly(indexMapping); err != nil {
		return
	}
	return
}

func NewMemOnlyIndexWithDocMaps(tag language.Tag, docMaps map[string]*mapping.DocumentMapping) (index bleve.Index, err error) {
	indexMapping := NewIndexMapping(tag)
	for doctype, dm := range docMaps {
		indexMapping.AddDocumentMapping(doctype, dm)
	}
	index, err = bleve.NewMemOnly(indexMapping)
	return
}