//go:build !exclude_pages_formats && !exclude_pages_format_tmpl

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

package tmpl

import (
	"fmt"
	htmlTemplate "html/template"
	"net/http"

	"github.com/blevesearch/bleve/v2/mapping"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
)

const Tag feature.Tag = "pages-formats-tmpl"

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
	name = "tmpl"
	return
}

func (f *CFeature) Extensions() (extensions []string) {
	extensions = append(extensions, "tmpl")
	return
}

func (f *CFeature) Label() (label string) {
	label = "TMPL"
	return
}

func (f *CFeature) Prepare(ctx context.Context, content string) (out context.Context, err error) {
	return
}

func (f *CFeature) Process(ctx context.Context, content string) (html htmlTemplate.HTML, redirect string, err error) {
	renderer := f.Enjin.GetThemeRenderer(ctx)
	if rendered, e := renderer.RenderTextTemplateContent(f.Enjin.MustGetTheme(), ctx, content); e != nil {
		err = e
		return
	} else {
		html = htmlTemplate.HTML(rendered)
	}
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
	r = lang.SetTag(r, pg.LanguageTag())
	for _, ptp := range feature.FilterTyped[feature.PageTypeProcessor](f.Enjin.Features().List()) {
		if v, _, processed, e := ptp.ProcessRequestPageType(r, pg); e != nil {
			log.ErrorF("error processing page type for tmpl format indexing: %v - %v", pg.Url(), e)
		} else if processed {
			pg = v
		}
	}

	var rendered string
	renderer := f.Enjin.GetThemeRenderer(pg.Context())
	if rendered, err = renderer.RenderTextTemplateContent(f.Enjin.MustGetTheme(), pg.Context(), pg.Content()); err != nil {
		err = fmt.Errorf("error rendering .tmpl content: %v", err)
		return
	}

	doc := NewTmplDocument(pg.Language(), pg.Url(), pg.Title())
	doc.Contents = append(doc.Contents, rendered)

	out = doc
	return
}