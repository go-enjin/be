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
	"github.com/go-enjin/be/pkg/log"
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
		queryErrors := make(map[string]error)
		queryStrings := make(map[string]string)
		queryResults := make(map[string][]*page.Page)
		for k, v := range ctxQueries {
			if q, ok := v.(string); ok {
				key := strcase.ToCamel(k)
				queryStrings[key] = q
				if matches, e := f.Enjin.CheckMatchQL(q); e != nil {
					queryErrors[key] = e
				} else {
					queryResults[key] = matches
				}
			} else {
				err = fmt.Errorf("unexpected query context value structure: %T", v)
				log.ErrorRF(r, "%v", err)
				return
			}
		}

		if len(queryErrors) > 0 {
			p.Context.SetSpecific("QueryErrors", queryErrors)
		}
		p.Context.SetSpecific("Query", queryStrings)
		p.Context.SetSpecific("QueryResults", queryResults)
		processed = true
	}

	if ctxSelects, ok := p.Context.Get("Select").(map[string]interface{}); ok {
		selectedErrors := make(map[string]error)
		selectedStrings := make(map[string]string)
		selectedValues := make(map[string]interface{})
		for k, v := range ctxSelects {
			if q, ok := v.(string); ok {
				key := strcase.ToCamel(k)
				selectedStrings[key] = q
				if selected, ee := f.Enjin.CheckSelectQL(q); ee != nil {
					selectedErrors[key] = ee
				} else if len(selected) == 1 {
					for _, only := range selected {
						selectedValues[key] = only
						break
					}
				} else {
					selectedValues[key] = selected
				}
			} else {
				err = fmt.Errorf("unexpected select context value structure: %T", v)
				log.ErrorRF(r, "%v", err)
				return
			}
		}

		if len(selectedErrors) > 0 {
			p.Context.SetSpecific("SelectedErrors", selectedErrors)
		}
		p.Context.SetSpecific("Select", selectedStrings)
		p.Context.SetSpecific("Selected", selectedValues)
		processed = true
	}

	if processed {
		pg = p
	}
	return
}