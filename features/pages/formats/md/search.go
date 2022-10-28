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

package md

import (
	"github.com/blevesearch/bleve/v2/analysis/analyzer/standard"
	"github.com/blevesearch/bleve/v2/mapping"

	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/search"
	"github.com/go-enjin/golang-org-x-text/language"
)

var _ Document = (*CDocument)(nil)

type Document interface {
	search.Document
	AddLink(text string)
	AddHeading(text string)
	AddFootnote(text string)
}

type CDocument struct {
	search.CDocument

	Links     []string `json:"links"`
	Headings  []string `json:"headings"`
	Footnotes []string `json:"footnotes"`
}

func NewMarkdownDocument(language, url, title string) (doc *CDocument) {
	doc = new(CDocument)
	doc.Type = "md"
	doc.Url = url
	doc.Title = title
	doc.Language = language
	return
}

func (d *CDocument) Self() interface{} {
	return d
}

func (d *CDocument) AddLink(text string) {
	d.Links = append(d.Links, text)
}

func (d *CDocument) AddHeading(text string) {
	d.Headings = append(d.Headings, text)
}

func (d *CDocument) AddFootnote(text string) {
	d.Footnotes = append(d.Footnotes, text)
}

func NewMarkdownDocumentMapping(tag language.Tag) (dm *mapping.DocumentMapping) {
	dm = search.NewDocumentMapping()
	analyzer := standard.Name
	if lang.BleveSupportedAnalyzer(tag) {
		analyzer = tag.String()
	}
	dm.AddFieldMappingsAt("links", search.NewDefaultTextFieldMapping(analyzer))
	dm.AddFieldMappingsAt("headings", search.NewDefaultTextFieldMapping(analyzer))
	dm.AddFieldMappingsAt("footnotes", search.NewDefaultTextFieldMapping(analyzer))
	return
}