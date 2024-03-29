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
	"fmt"
	htmlTemplate "html/template"
	"net/http"
	"strings"

	"github.com/blevesearch/bleve/v2/mapping"

	"github.com/go-corelibs/x-text/language"
	"github.com/go-corelibs/x-text/message"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
)

const Tag feature.Tag = "pages-formats-json"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type Feature interface {
	feature.Feature
	feature.PageFormat
}

type MakeFeature interface {
	Make() Feature
}

type CFeature struct {
	feature.CFeature
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.CFeature.Construct(f)
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Setup(enjin feature.Internals) {
	f.CFeature.Setup(enjin)
}

func (f *CFeature) Name() (name string) {
	name = "text"
	return
}

func (f *CFeature) Extensions() (extensions []string) {
	extensions = append(extensions, "json", "json.tmpl")
	return
}

func (f *CFeature) Label() (label string) {
	label = "JSON"
	return
}

func (f *CFeature) Prepare(ctx context.Context, content string) (out context.Context, err error) {
	return
}

func (f *CFeature) Process(ctx context.Context, content string) (html htmlTemplate.HTML, redirect string, err error) {
	html = htmlTemplate.HTML(content)
	return
}

func (f *CFeature) SearchDocumentMapping(tag language.Tag) (doctype string, dm *mapping.DocumentMapping) {
	doctype, _, dm = f.NewDocumentMapping(tag)
	return
}

func (f *CFeature) AddSearchDocumentMapping(tag language.Tag, indexMapping *mapping.IndexMappingImpl) {
	doctype, _, dm := f.NewDocumentMapping(tag)
	indexMapping.AddDocumentMapping(doctype, dm)
}

func (f *CFeature) IndexDocument(pg feature.Page) (out interface{}, err error) {

	r, _ := http.NewRequest("GET", pg.Url(), nil)
	r = message.SetTag(r, pg.LanguageTag())
	for _, ptp := range feature.FilterTyped[feature.PageTypeProcessor](f.Enjin.Features().List()) {
		if v, _, processed, e := ptp.ProcessRequestPageType(r, pg); e != nil {
			log.ErrorF("error processing page type for json format indexing: %v - %v", pg.Url(), e)
		} else if processed {
			pg = v
		}
	}

	var rendered string
	if strings.HasSuffix(pg.Format(), ".tmpl") {
		renderer := f.Enjin.GetThemeRenderer(pg.Context())
		if rendered, err = renderer.RenderTextTemplateContent(f.Enjin.MustGetTheme(), pg.Context(), pg.Content()); err != nil {
			err = fmt.Errorf("error rendering .json.tmpl content: %v", err)
			return
		}
	} else {
		rendered = pg.Content()
	}

	doc := NewTextDocument(pg.Language(), pg.Url(), pg.Title())
	doc.Contents = append(doc.Contents, rendered)

	out = doc
	return
}
