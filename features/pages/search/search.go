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
	"html/template"
	"math"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/urfave/cli/v2"
	"golang.org/x/net/html"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/page"
	"github.com/go-enjin/be/pkg/pagecache"
	"github.com/go-enjin/be/pkg/request/argv"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

var (
	_ Feature                   = (*CFeature)(nil)
	_ MakeFeature               = (*CFeature)(nil)
	_ feature.PageTypeProcessor = (*CFeature)(nil)
)

const Tag feature.Tag = "PagesSearch"

type Feature interface {
	feature.Feature
}

type CFeature struct {
	feature.CFeature

	cli   *cli.Context
	enjin feature.Internals

	path   string
	search pagecache.SearchEnjinFeature

	sync.RWMutex
}

type MakeFeature interface {
	SetPath(path string) MakeFeature

	Make() Feature
}

func New() MakeFeature {
	f := new(CFeature)
	f.Init(f)
	return f
}

func (f *CFeature) SetPath(path string) MakeFeature {
	f.path = path
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
}

func (f *CFeature) Tag() (tag feature.Tag) {
	tag = Tag
	return
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
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
	f.enjin = enjin
	if f.path == "" {
		f.path = "/search"
	}
	log.DebugF("using search path: %v", f.path)
	for _, feat := range f.enjin.Features() {
		if search, ok := feat.(pagecache.SearchEnjinFeature); ok {
			f.search = search
			break
		}
	}
	if f.search == nil {
		log.FatalF("searching pages requires a pagecache.SearchEnjinFeature")
	}
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	f.cli = ctx
	return
}

func (f *CFeature) FilterPageContext(themeCtx, pageCtx context.Context, r *http.Request) (out context.Context) {
	out = themeCtx
	out.SetSpecific("SiteSearchable", true)
	out.SetSpecific("SiteSearchPath", f.path)
	return
}

func (f *CFeature) ProcessRequestPageType(r *http.Request, p *page.Page) (pg *page.Page, redirect string, processed bool, err error) {
	if p.Type == "search" {
		if beStrings.StringInStrings(r.Method, "GET", "") {
			if len(r.URL.Query()) > 0 {
				if redirect, err = f.handleQueryRedirect(r); err != nil {
					p.Context.SetSpecific("SiteSearchError", err.Error())
					pg = p
					err = nil
					processed = true
					return
				} else if redirect != "" {
					processed = true
					// log.WarnF("redirecting from: %v, to: %v", r.URL.Path, redirect)
					return
				}
			}
		}

		// prepare arguments
		reqArgv := argv.GetRequestArgv(r)
		p.Context.SetSpecific(argv.RequestArgvConsumedKey, true)
		reqLangTag := lang.GetTag(r)
		numPerPage, pageNumber := 10, 0
		if reqArgv.NumPerPage > -1 {
			numPerPage = reqArgv.NumPerPage
		}
		if reqArgv.PageNumber > -1 {
			pageNumber = reqArgv.PageNumber
		}
		var input string
		if len(reqArgv.Argv) > 0 {
			for idx, argv := range reqArgv.Argv {
				if idx > 0 {
					input += " "
				}
				input += strings.Join(argv, ",")
			}
		}
		if cleaned, err := url.QueryUnescape(input); err != nil {
			log.ErrorF("error unescaping input query string: %v", err)
		} else {
			input = cleaned
		}
		input = html.UnescapeString(input)
		input = forms.StripTags(input)
		p.Context.SetSpecific("SiteSearchQuery", input)

		// log.WarnF("search info: numPerPage=%d, pageNumber=%d, input=%v", numPerPage, pageNumber, reqArgv.Argv)

		if input != "" {
			// perform search
			if results, err := f.search.PerformSearch(reqLangTag, input, numPerPage, pageNumber); err != nil {
				p.Context.SetSpecific("SiteSearchError", err.Error())
			} else {
				numPages := int(math.Ceil(float64(results.Total) / float64(numPerPage)))
				numHits := len(results.Hits)
				idStart := pageNumber*numPerPage + 1
				idEnd := idStart + numHits - 1

				printer := lang.GetPrinterFromRequest(r)
				var resultSummary, hitsSummary, pageSummary string
				switch results.Total {
				case 0:
					// Search result summary when no results are found
					resultSummary = printer.Sprintf("No results found")
				case 1:
					// Search result summary when exactly one result is found
					resultSummary = printer.Sprintf("1 result found")
				default:
					// Search result summary, <total-hits>
					resultSummary = printer.Sprintf("%d results found", results.Total)
				}
				switch numHits {
				case 1:
					// Search page summary, <number> of <pages>
					pageSummary = printer.Sprintf("Page %d of %d", pageNumber+1, numPages)
					// Search hits summary with only one hit, <hit-number> of <total-hits>
					hitsSummary = printer.Sprintf("Showing #%d of %d", idStart, results.Total)
				default:
					// Search page summary, <number> of <pages>
					pageSummary = printer.Sprintf("Page %d of %d", pageNumber+1, numPages)
					// Search hits summary with more than one hit, <first-hit-number>-<last-hit-number> of <total-hits>
					hitsSummary = printer.Sprintf("Showing %d-%d of %d", idStart, idEnd, results.Total)
				}

				p.Context.SetSpecific("SiteSearchSize", numPerPage)
				p.Context.SetSpecific("SiteSearchPage", pageNumber)
				p.Context.SetSpecific("SiteSearchPages", numPages)
				p.Context.SetSpecific("SiteSearchResults", results)
				p.Context.SetSpecific("SiteSearchPageSummary", template.HTML(pageSummary))
				p.Context.SetSpecific("SiteSearchHitsSummary", template.HTML(hitsSummary))
				p.Context.SetSpecific("SiteSearchResultSummary", template.HTML(resultSummary))
			}
		}

		// finalize
		pg = p
		redirect = ""
		processed = true
	}
	return
}