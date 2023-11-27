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

package search

import (
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"

	"github.com/go-enjin/golang-org-x-text/language"
)

func NewIndexMapping(tag language.Tag) (indexMapping *mapping.IndexMappingImpl) {
	indexMapping = bleve.NewIndexMapping()
	_, indexMapping.DefaultMapping = NewDocumentMapping(tag)
	indexMapping.AddDocumentMapping("document", indexMapping.DefaultMapping)
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
