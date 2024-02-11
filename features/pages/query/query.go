//go:build page_query || pages || all

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

package query

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/maps"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "pages-query"

type Feature interface {
	feature.Feature
	feature.PageTypeProcessor
	feature.PageContextFieldsProvider
}

type MakeFeature interface {
	Make() Feature
}

type CFeature struct {
	feature.CFeature

	sync.RWMutex
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

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	return
}

func (f *CFeature) MakePageContextFields(r *http.Request) (fields context.Fields) {
	//printer := message.GetPrinter(r)
	fields = context.Fields{
		"query": {
			Key:    "query",
			Tab:    "query",
			Format: "string-map",
		},
		"select": {
			Key:    "select",
			Tab:    "query",
			Format: "string-map",
		},
	}
	return
}

func (f *CFeature) PageTypeNames() (names []string) {
	names = []string{"query"}
	return
}

func (f *CFeature) ProcessRequestPageType(r *http.Request, p feature.Page) (pg feature.Page, redirect string, processed bool, err error) {
	if p.Type() != "query" {
		return
	}

	if ctxQueries, ok := p.Context().Get("Query").(map[string]interface{}); ok {
		qErrors := make(map[string]error)
		qInputs := make(map[string]string)
		qResults := make(map[string][]feature.Page)
		for _, queryKey := range maps.SortedKeys(ctxQueries) {
			camelKey := strcase.ToCamel(queryKey)
			queryInput := ctxQueries[queryKey]
			if q, ok := queryInput.(string); ok {
				qInputs[camelKey] = q
				if matches, e := f.Enjin.CheckMatchQL(q); e != nil {
					qErrors[camelKey] = e
				} else {
					f.Enjin.ApplyPageContextUpdaters(r, matches...)
					qResults[camelKey] = matches
				}
			} else {
				qErrors[camelKey] = fmt.Errorf("unexpected query context value structure: %T", queryInput)
			}
		}

		if len(qErrors) > 0 {
			p.Context().SetSpecific("QueryErrors", qErrors)
		}
		p.Context().SetSpecific("Query", qInputs)
		p.Context().SetSpecific("QueryResults", qResults)
		processed = true
	}

	if ctxSelects, ok := p.Context().Get("Select").(map[string]interface{}); ok {
		sErrors := make(map[string]error)
		sInputs := make(map[string]string)
		sResults := make(map[string]interface{})
		for selectKey, selectInput := range ctxSelects {
			camelKey := strcase.ToCamel(selectKey)
			if q, ok := selectInput.(string); ok {
				sInputs[camelKey] = q
				if selected, ee := f.Enjin.CheckSelectQL(q); ee != nil {
					sErrors[camelKey] = ee
				} else if len(selected) == 1 {
					for _, only := range selected {
						sResults[camelKey] = only
						break
					}
				} else {
					sResults[camelKey] = selected
				}
			} else {
				sErrors[camelKey] = fmt.Errorf("unexpected select context value structure: %T", selectInput)
			}
		}

		if len(sErrors) > 0 {
			p.Context().SetSpecific("SelectedErrors", sErrors)
		}
		p.Context().SetSpecific("Select", sInputs)
		p.Context().SetSpecific("Selected", sResults)
		processed = true
	}

	if processed {
		pg = p
	}
	return
}
