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

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/search"
)

func NewFeaturesIndex(e feature.Internals) (index bleve.Index, err error) {
	if index, err = NewMemOnlyIndex(e); err != nil {
		return
	}
	for _, f := range e.Features() {
		if s, ok := f.(feature.Searchable); ok {
			log.DebugF("updating searchable feature: %v", f.Tag())
			if ee := s.UpdateSearch(index); ee != nil {
				err = ee
				return
			}
		}
	}
	return
}

func NewMemOnlyIndex(e feature.Internals) (index bleve.Index, err error) {
	indexMapping := bleve.NewIndexMapping()
	indexMapping.DefaultMapping = search.NewDocumentMapping()
	indexMapping.AddDocumentMapping("document", indexMapping.DefaultMapping)

	for _, f := range e.Features() {
		if m, ok := f.(feature.SearchDocumentMapper); ok {
			log.DebugF("adding %v search document mapping", f.Tag())
			m.AddSearchDocumentMapping(indexMapping)
		}
	}

	if index, err = bleve.NewMemOnly(indexMapping); err != nil {
		return
	}
	return
}