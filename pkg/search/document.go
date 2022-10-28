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
	"github.com/blevesearch/bleve/v2/analysis/analyzer/simple"
	"github.com/blevesearch/bleve/v2/mapping"
)

var _ Document = (*CDocument)(nil)

type Document interface {
	Self() interface{}
	GetUrl() (url string)
	GetTitle() (title string)
	GetLanguage() (language string)
	GetContents() (contents []string)
	BleveType() string
	AddContent(text string)
}

type CDocument struct {
	Type     string   `json:"type"`
	Url      string   `json:"url"`
	Title    string   `json:"title"`
	Language string   `json:"language"`
	Contents []string `json:"contents"`
}

func NewDocument(language, url, title string) (doc *CDocument) {
	doc = new(CDocument)
	doc.Type = "document"
	doc.Url = url
	doc.Title = title
	doc.Language = language
	return
}

func (d *CDocument) Self() interface{} {
	return d
}

func (d *CDocument) GetUrl() (url string) {
	return d.Url
}

func (d *CDocument) GetTitle() (title string) {
	return d.Title
}

func (d *CDocument) GetLanguage() (language string) {
	return d.Language
}

func (d *CDocument) GetContents() (contents []string) {
	return d.Contents
}

func (d *CDocument) BleveType() string {
	return d.Type
}

func (d *CDocument) AddContent(text string) {
	d.Contents = append(d.Contents, text)
}

func NewDocumentMapping() (dm *mapping.DocumentMapping) {
	dm = bleve.NewDocumentMapping()

	newField := func(analyzer string) (f *mapping.FieldMapping) {
		f = bleve.NewTextFieldMapping()
		f.Store = true
		f.Analyzer = analyzer
		f.IncludeInAll = true
		f.IncludeTermVectors = true
		return
	}

	dm.AddFieldMappingsAt("Url", newField(simple.Name))
	dm.AddFieldMappingsAt("Title", newField(simple.Name))
	dm.AddFieldMappingsAt("Content", newField(simple.Name))
	dm.AddFieldMappingsAt("Headings", newField(simple.Name))
	dm.AddFieldMappingsAt("Language", newField(simple.Name))
	return
}