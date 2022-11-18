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

package page

import (
	"fmt"

	"github.com/blevesearch/bleve/v2/mapping"

	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/search"
)

func (p *Page) SearchMapping() (doctype string, dm *mapping.DocumentMapping, err error) {
	if format := p.formats.GetFormat(p.Format); format != nil {
		doctype, _, dm = format.NewDocumentMapping(p.LanguageTag)
	} else {
		err = fmt.Errorf("unsupported page format: %v", p.Format)
	}
	return
}

func (p *Page) SearchDocument() (doc search.Document, err error) {
	if pgType := p.Context.String("type", "page"); pgType != "page" {
		log.TraceF("skipping search index for (not page type): %v", p.Url)
		return
	}
	if pgSearchable := p.Context.String("Searchable", "true"); pgSearchable != "true" {
		log.TraceF("skipping search index for (not searchable): %v", p.Url)
		return
	}
	if format := p.formats.GetFormat(p.Format); format != nil {
		doc, err = format.IndexDocument(p)
	} else {
		err = fmt.Errorf("unsupported page format: %v", p.Format)
	}
	return
}