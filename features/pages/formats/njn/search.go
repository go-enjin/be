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

package njn

import (
	"github.com/blevesearch/bleve/v2/analysis/analyzer/simple"
	"github.com/blevesearch/bleve/v2/mapping"

	"github.com/go-enjin/be/pkg/search"
)

var _ Document = (*CDocument)(nil)

type Document interface {
	search.Document
	AddHeading(text string)
}

type CDocument struct {
	search.CDocument

	Links     []string `json:"links"`
	Headings  []string `json:"headings"`
	Footnotes []string `json:"footnotes"`
}

func NewEnjinDocument(url, title string) (doc *CDocument) {
	doc = new(CDocument)
	doc.Type = "njn"
	doc.Title = title
	doc.Url = url
	return
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

func NewEnjinDocumentMapping() (dm *mapping.DocumentMapping) {
	dm = search.NewDocumentMapping()
	dm.AddFieldMappingsAt("links", search.NewDefaultTextFieldMapping(simple.Name))
	dm.AddFieldMappingsAt("headings", search.NewDefaultTextFieldMapping(simple.Name))
	dm.AddFieldMappingsAt("footnotes", search.NewDefaultTextFieldMapping(simple.Name))
	return
}