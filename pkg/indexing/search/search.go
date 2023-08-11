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

package search

import (
	"fmt"

	"github.com/blevesearch/bleve/v2/mapping"

	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/page"
	"github.com/go-enjin/be/pkg/search"
	"github.com/go-enjin/be/pkg/slices"
)

var (
	knownSearchPageTypes = []string{"page"}
)

func RegisterSearchPageType(pgtype string) {
	if !slices.Within(pgtype, knownSearchPageTypes) {
		knownSearchPageTypes = append(knownSearchPageTypes, pgtype)
	}
}

func SearchMapping(p *page.Page) (doctype string, dm *mapping.DocumentMapping, err error) {
	if format := p.Formats.GetFormat(p.Format); format != nil {
		doctype, _, dm = format.NewDocumentMapping(p.LanguageTag)
	} else {
		err = fmt.Errorf("unsupported page format: %v", p.Format)
	}
	return
}

func SearchDocument(p *page.Page) (doc search.Document, err error) {
	if pgType := p.Context.String("type", "page"); !slices.Within(pgType, knownSearchPageTypes) {
		log.TraceF("skipping search index for (not known search page type): %v", p.Url)
		return
	}
	if pgSearchable := p.Context.String("Searchable", "true"); pgSearchable != "true" {
		log.TraceF("skipping search index for (not searchable): %v", p.Url)
		return
	}
	if format := p.Formats.GetFormat(p.Format); format != nil {
		if v, e := format.IndexDocument(p); e != nil {
			err = e
		} else if v != nil {
			var ok bool
			if doc, ok = v.(search.Document); !ok {
				log.ErrorF("format.IndexDocument returned invalid structure: %T", v)
			}
		} else {
			log.ErrorF("format indexing had nil result: %v - %v", p.Format, p.Url)
		}
	} else {
		err = fmt.Errorf("unsupported page format: %v (valid: %v)", p.Format, p.Formats.ListFormats())
	}
	return
}