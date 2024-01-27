//go:build page_search || pages || all

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
	"fmt"
	"html/template"
	"math"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/urfave/cli/v2"
	"golang.org/x/net/html"

	"github.com/go-corelibs/slices"
	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/request/argv"
)

var (
	DefaultSearchPath = "/search"
)

const Tag feature.Tag = "pages-search"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type ResultsPostProcessor interface {
	SearchResultsPostProcess(r *http.Request, p feature.Page)
}

type Feature interface {
	feature.Feature
	feature.PageTypeProcessor
	feature.PageContextModifier
}

type MakeFeature interface {
	SetSearchPath(path string) MakeFeature
	SetSearchEnjin(tag feature.Tag) MakeFeature

	Make() Feature
}

type CFeature struct {
	feature.CFeature

	path   string
	sefTag feature.Tag
	search feature.SearchEnjinFeature

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
	f.path = DefaultSearchPath
	f.sefTag = feature.NilTag
}

func (f *CFeature) SetSearchPath(path string) MakeFeature {
	f.path = path
	return f
}

func (f *CFeature) SetSearchEnjin(tag feature.Tag) MakeFeature {
	f.sefTag = tag
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	if f.sefTag == feature.NilTag {
		if f.search = feature.FirstTyped[feature.SearchEnjinFeature](b.Features().List()); f.search == nil {
			err = fmt.Errorf("%v feature requires at least one feature.SearchEnjinFeature present", f.Tag())
			return
		}
	} else if v, ok := b.Features().Get(f.sefTag); ok {
		if sef, ok := v.(feature.SearchEnjinFeature); ok {
			f.search = sef
		} else {
			err = fmt.Errorf("%v is not a feature.SearchEnjinFeature", v.Tag())
			return
		}
	} else {
		err = fmt.Errorf("%v feature.SearchEnjinFeature not found", f.sefTag)
		return
	}

	b.AddCommands(&cli.Command{
		Name:        "search",
		Usage:       "search through content",
		Action:      f.SearchAction,
		UsageText:   globals.BinName + " search -- query string",
		Description: "Search for content within an enjin environment",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "size",
				Usage: "results per page",
				Value: 10,
			},
			&cli.IntFlag{
				Name:  "pg",
				Usage: "page to return",
				Value: 0,
			},
		},
	})
	return
}

func (f *CFeature) Setup(enjin feature.Internals) {
	f.CFeature.Setup(enjin)

	if f.path == "" {
		f.path = "/search"
	}

	log.DebugF("using search path: %v", f.path)
	log.DebugF("using search enjin: %v", f.search.Tag())
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	return
}

func (f *CFeature) FilterPageContext(themeCtx, pageCtx context.Context, r *http.Request) (out context.Context) {
	out = themeCtx
	out.SetSpecific("SiteSearchable", true)
	out.SetSpecific("SiteSearchPath", f.path)
	return
}

func (f *CFeature) PageTypeNames() (names []string) {
	names = []string{"search"}
	return
}

func (f *CFeature) ProcessRequestPageType(r *http.Request, p feature.Page) (pg feature.Page, redirect string, processed bool, err error) {
	if p.Type() == "search" {
		if slices.Present(r.Method, "GET", "") {
			if len(r.URL.Query()) > 0 {
				if redirect, err = f.handleQueryRedirect(r); err != nil {
					p.Context().SetSpecific("SiteSearchError", err.Error())
					pg = p
					err = nil
					processed = true
					return
				} else if redirect != "" {
					processed = true
					// log.WarnRF(r, "redirecting from: %v, to: %v", r.URL.Path, redirect)
					return
				}
			}
		}

		// prepare arguments
		reqArgv := argv.Get(r)
		p.Context().SetSpecific(argv.RequestConsumedKey, true)
		reqLangTag := lang.GetTag(r)
		numPerPage, pageIndex, pageNumber := p.Context().ValueAsInt("DefaultNumPerPage", 10), 0, 1
		if reqArgv.NumPerPage > -1 {
			numPerPage = reqArgv.NumPerPage
		}
		if reqArgv.PageNumber > -1 {
			pageIndex = reqArgv.PageNumber
			pageNumber = pageIndex + 1
		}
		var input string
		if len(reqArgv.Argv) > 0 {
			for idx, args := range reqArgv.Argv {
				if idx > 0 {
					input += " "
				}
				input += strings.Join(args, " ")
			}
		}
		if cleaned, err := url.PathUnescape(input); err != nil {
			log.ErrorRF(r, "error unescaping input query string: %v", err)
		} else {
			input = cleaned
		}
		input = html.UnescapeString(input)
		// input = forms.StrictPolicy(input)

		query := f.search.PrepareSearch(reqLangTag, input)
		queryPath := f.path + "/:" + url.PathEscape(query)
		if reqArgv.NumPerPage > 0 && reqArgv.PageNumber >= 0 {
			queryPath += fmt.Sprintf("/%d/%d/", reqArgv.NumPerPage, reqArgv.PageNumber)
		}
		if len(reqArgv.Argv) > 0 && reqArgv.String() != queryPath {
			redirect = queryPath
			return
		}
		p.Context().SetSpecific("SiteSearchQuery", query)

		if input != "" {
			// perform search
			if results, err := f.search.PerformSearch(reqLangTag, query, numPerPage, pageIndex); err != nil {
				p.Context().SetSpecific("SiteSearchError", err.Error())
			} else {
				numPages := int(math.Ceil(float64(results.Total) / float64(numPerPage)))
				numHits := len(results.Hits)
				idStart := pageIndex*numPerPage + 1
				idEnd := idStart + numHits - 1

				printer := lang.GetPrinterFromRequest(r)
				var resultSummary, hitsSummary, pageSummary string
				// Search result summary, <total-hits>
				resultSummary = printer.Sprintf("%[1]d results found", results.Total)
				// Search page summary, <number> of <pages>
				pageSummary = printer.Sprintf("Page %[1]d of %[2]d", pageNumber, numPages)
				switch numHits {
				case 1:
					// Search hits summary with only one hit, <hit-number> of <total-hits>
					hitsSummary = printer.Sprintf("Showing #%[1]d of %[2]d", idStart, results.Total)
				default:
					// Search hits summary with more than one hit, <first-hit-number>-<last-hit-number> of <total-hits>
					hitsSummary = printer.Sprintf("Showing %[1]d-%[2]d of %[3]d", idStart, idEnd, results.Total)
				}

				p.Context().SetSpecific("SiteSearchSize", numPerPage)
				p.Context().SetSpecific("SiteSearchPage", pageIndex)
				p.Context().SetSpecific("SiteSearchPages", numPages)
				p.Context().SetSpecific("SiteSearchResults", results)
				p.Context().SetSpecific("SiteSearchPageSummary", template.HTML(pageSummary))
				p.Context().SetSpecific("SiteSearchHitsSummary", template.HTML(hitsSummary))
				p.Context().SetSpecific("SiteSearchResultSummary", template.HTML(resultSummary))
			}
		}

		// finalize
		for _, rp := range feature.FilterTyped[ResultsPostProcessor](f.Enjin.Features().List()) {
			rp.SearchResultsPostProcess(r, p)
		}
		pg = p
		redirect = ""
		processed = true
	}
	return
}
