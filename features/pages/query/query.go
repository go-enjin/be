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
	"github.com/go-enjin/be/pkg/page"
)

var (
	_ MakeFeature        = (*CFeature)(nil)
	_ feature.Middleware = (*CFeature)(nil)
)

const Tag feature.Tag = "PagesQuery"

type Feature interface {
	feature.Middleware
}

type CFeature struct {
	feature.CMiddleware

	cli   *cli.Context
	enjin feature.Internals

	sync.RWMutex
}

type MakeFeature interface {
	Make() Feature
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
	f.CMiddleware.Init(this)
}

func (f *CFeature) Tag() (tag feature.Tag) {
	tag = Tag
	return
}

func (f *CFeature) Setup(enjin feature.Internals) {
	f.enjin = enjin
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	f.cli = ctx
	return
}

func (f *CFeature) ProcessRequestPageType(r *http.Request, p *page.Page) (pg *page.Page, redirect string, processed bool, err error) {
	// reqArgv := site.GetRequestArgv(r)
	if p.Type == "query" {
		if ctxQueries, ok := p.Context.Get("Query").(context.Context); ok {
			queryStrings := make(map[string]string)
			queryResults := make(map[string][]*page.Page)
			for k, v := range ctxQueries {
				if q, ok := v.(string); ok {
					key := strcase.ToCamel(k)
					queryStrings[key] = q
					queryResults[key] = f.enjin.MatchQL(q)
				} else {
					err = fmt.Errorf("unexpected query context value structure: %T", v)
					return
				}
			}
			p.Context.SetSpecific("Query", queryStrings)
			p.Context.SetSpecific("QueryResults", queryResults)
			pg = p
			processed = true
		}
	}
	return
}