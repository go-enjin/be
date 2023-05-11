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

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/page"
)

var (
	_ MakeFeature               = (*CFeature)(nil)
	_ feature.PageTypeProcessor = (*CFeature)(nil)
)

const Tag feature.Tag = "pages-query"

type Feature interface {
	feature.Feature
}

type CFeature struct {
	feature.CFeature

	sync.RWMutex
}

type MakeFeature interface {
	Make() Feature
}

func New() MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.FeatureTag = Tag
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	return
}

func (f *CFeature) ProcessRequestPageType(r *http.Request, p *page.Page) (pg *page.Page, redirect string, processed bool, err error) {
	if p.Type != "query" {
		return
	}

	if ctxQueries, ok := p.Context.Get("Query").(map[string]interface{}); ok {
		qErrors := make(map[string]error)
		qInputs := make(map[string]string)
		qResults := make(map[string][]*page.Page)
		for _, queryKey := range maps.SortedKeys(ctxQueries) {
			camelKey := strcase.ToCamel(queryKey)
			queryInput := ctxQueries[queryKey]
			if q, ok := queryInput.(string); ok {
				qInputs[camelKey] = q
				if matches, e := f.Enjin.CheckMatchQL(q); e != nil {
					qErrors[camelKey] = e
				} else {
					qResults[camelKey] = matches
				}
			} else {
				qErrors[camelKey] = fmt.Errorf("unexpected query context value structure: %T", queryInput)
			}
		}

		if len(qErrors) > 0 {
			p.Context.SetSpecific("QueryErrors", qErrors)
		}
		p.Context.SetSpecific("Query", qInputs)
		p.Context.SetSpecific("QueryResults", qResults)
		processed = true
	}

	if ctxSelects, ok := p.Context.Get("Select").(map[string]interface{}); ok {
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
			p.Context.SetSpecific("SelectedErrors", sErrors)
		}
		p.Context.SetSpecific("Select", sInputs)
		p.Context.SetSpecific("Selected", sResults)
		processed = true
	}

	if processed {
		pg = p
	}
	return
}