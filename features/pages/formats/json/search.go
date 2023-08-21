//go:build !exclude_pages_formats && !exclude_pages_format_json

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

package json

import (
	"github.com/blevesearch/bleve/v2/mapping"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/search"
)

var _ Document = (*CDocument)(nil)

type Document interface {
	search.Document
}

type CDocument struct {
	search.CDocument
}

func NewTextDocument(language, url, title string) (doc *CDocument) {
	doc = new(CDocument)
	doc.Type = "json"
	doc.Url = url
	doc.Title = title
	doc.Language = language
	return
}

func (d *CDocument) Self() interface{} {
	return d
}

func (f *CFeature) NewDocumentMapping(tag language.Tag) (doctype, analyzer string, dm *mapping.DocumentMapping) {
	doctype = f.Name()
	analyzer, dm = search.NewDocumentMapping(tag)
	return
}