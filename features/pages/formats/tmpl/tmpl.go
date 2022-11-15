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
	"bytes"
	"html/template"

	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/search"
	"github.com/go-enjin/be/pkg/types/theme-types"
)

var (
	_ Feature      = (*CFeature)(nil)
	_ MakeFeature  = (*CFeature)(nil)
	_ types.Format = (*CFeature)(nil)
)

type Feature interface {
	feature.Feature
	types.Format
}

type MakeFeature interface {
	Make() Feature
}

type CFeature struct {
	feature.CFeature

	enjin feature.Internals
}

func New() MakeFeature {
	f := new(CFeature)
	f.Init(f)
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
}

func (f *CFeature) Setup(enjin feature.Internals) {
	f.enjin = enjin
}

func (f *CFeature) Tag() (tag feature.Tag) {
	tag = "PageFormatTMPL"
	return
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

func (f *CFeature) Process(ctx context.Context, t types.Theme, content string) (html template.HTML, err *types.EnjinError) {
	if tmpl, e := t.NewTextTemplateWithContext("content.tmpl", ctx); e != nil {
		log.ErrorF("error preparing text template: %v", e)
		return
	} else {
		if tt, ee := tmpl.Parse(content); ee == nil {
			var w bytes.Buffer
			if ee = tt.Execute(&w, ctx); ee == nil {
				html = template.HTML(w.Bytes())
			} else {
				log.ErrorF("error rendering text template: %v", ee)
			}
		}
	}
	return
}

func (f *CFeature) AddSearchDocumentMapping(tag language.Tag, indexMapping *mapping.IndexMappingImpl) {
}

func (f *CFeature) IndexDocument(thing interface{}) (doc search.Document, err error) {
	return
}