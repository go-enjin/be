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
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/blevesearch/bleve/v2"
	bleveHtml "github.com/blevesearch/bleve/v2/search/highlight/format/html"
	"github.com/urfave/cli/v2"
	"golang.org/x/net/html"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/forms/nonce"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/page"
	beSearch "github.com/go-enjin/be/pkg/search/indexes"
)

var (
	_ MakeFeature          = (*CFeature)(nil)
	_ feature.Middleware   = (*CFeature)(nil)
	_ feature.PageProvider = (*CFeature)(nil)
)

const Tag feature.Tag = "PagesSearch"

type Feature interface {
	feature.Middleware
}

type CFeature struct {
	feature.CMiddleware

	cli   *cli.Context
	enjin feature.Internals

	path string

	findingSelf bool
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
	f.CMiddleware.Init(this)
}

func (f *CFeature) Tag() (tag feature.Tag) {
	tag = Tag
	return
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	b.AddCommands(&cli.Command{
		Name:      "search",
		Usage:     "search through searchable content",
		Action:    f.SearchAction,
		UsageText: globals.BinName + " search -- query string",
		Description: "All features that are feature.Searchable are indexed" +
			" and queried using the Bleve text indexing package." +
			" See: http://blevesearch.com/docs/Query-String-Query/ for more" +
			"details on how to use the query string.",
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
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	f.cli = ctx
	return
}

func (f *CFeature) SearchAction(ctx *cli.Context) (err error) {
	var input string
	var size, pg int
	argv := ctx.Args().Slice()
	argc := len(argv)
	if argc == 0 {
		err = fmt.Errorf("search query is required")
		return
	}
	input = argv[0]
	if v := ctx.Int("size"); v > 0 {
		size = v
	} else {
		size = 10
	}
	if v := ctx.Int("pg"); v > 1 {
		pg = v - 1
	}

	if results, e := f.PerformSearch(language.Und, input, size, pg); e != nil {
		err = e
	} else {
		fmt.Printf("%v\n", results)
	}
	return
}

var RxLanguageKey = regexp.MustCompile(`language:(\*|[a-z][-a-zA-Z]+)\s*`)

func (f *CFeature) PerformSearch(tag language.Tag, input string, size, pg int) (results *bleve.SearchResult, err error) {
	if indexes, all, e := beSearch.NewFeaturesIndex(f.enjin); e != nil {
		err = e
		return
	} else {
		searchAll := false
		inputWantsTag := language.Und
		input = forms.StripTags(input)
		if i, ee := url.PathUnescape(input); ee != nil {
			log.ErrorF("error unescaping input: %v - %v", input, ee)
		} else {
			input = i
		}
		input = html.UnescapeString(input)

		log.DebugF("performing site search: %v", input)

		// handle user input `language:%v`
		if RxLanguageKey.MatchString(input) {
			m := RxLanguageKey.FindAllStringSubmatch(input, 1)
			if m[0][1] == "*" {
				searchAll = true
				input = RxLanguageKey.ReplaceAllString(input, "")
			} else if queryLangTag, eee := language.Parse(m[0][1]); eee != nil {
				err = fmt.Errorf("invalid language")
				return
			} else {
				var found bool
				for _, siteLocale := range f.enjin.SiteLocales() {
					if found = language.Compare(siteLocale, queryLangTag); found {
						break
					}
				}
				if !found {
					err = fmt.Errorf("unsupported language")
					return
				}
				inputWantsTag = queryLangTag
				input = RxLanguageKey.ReplaceAllString(input, "")
			}
		}

		// construct a new query from the input
		query := bleve.NewQueryStringQuery(input)
		if err = query.Validate(); err != nil {
			return
		}

		// construct a new search request from the query
		req := bleve.NewSearchRequest(query)
		if size == 0 {
			size = 10
		}
		req.Size = size
		req.From = pg * size
		req.Fields = []string{"*"}
		req.Highlight = bleve.NewHighlightWithStyle(bleveHtml.Name)

		// determine which index to search
		index := all
		if !searchAll {
			if !language.Compare(inputWantsTag, language.Und) {
				if idx, ok := indexes[inputWantsTag]; ok {
					index = idx
				}
			}
			if index == all && !language.Compare(tag, language.Und) {
				if idx, ok := indexes[tag]; ok {
					index = idx
				}
			}
		}

		if results, e = index.Search(req); e != nil {
			err = e
			return
		}
	}
	return
}

func (f *CFeature) FilterPageContext(themeCtx, pageCtx context.Context, r *http.Request) (out context.Context) {
	out = themeCtx
	out.SetSpecific("SiteSearchable", true)
	out.SetSpecific("SiteSearchPath", f.path)
	return
}

var rxQueryPaginationPage = regexp.MustCompile(`(.+)/(\d+)/??$`)
var rxQueryPaginationSizePage = regexp.MustCompile(`(.+)/(\d+)/(\d+)/??$`)

func (f *CFeature) handleQueryRedirect(w http.ResponseWriter, r *http.Request) (ok bool, err error) {
	tag := lang.GetTag(r)
	printer := lang.GetPrinterFromRequest(r)
	var query string
	var foundNonce, foundQuery bool
	for k, v := range r.URL.Query() {
		switch k {
		case "nonce":
			value := forms.StripTags(v[0])
			if vv, e := url.QueryUnescape(value); e != nil {
				log.ErrorF("error un-escaping url path: %v", e)
			} else {
				value = vv
			}
			value = forms.Sanitize(value)
			if !nonce.Validate("site-search-form", value) {
				// Let the visitor know that their search form expired
				err = fmt.Errorf(printer.Sprintf("search form expired"))
				return
			}
			foundNonce = true
			if foundQuery {
				break
			}
		case "query":
			// trap random page ?query= requests?
			query = forms.StripTags(v[0])
			query = forms.Sanitize(query)
			if vv, e := url.QueryUnescape(query); e != nil {
				log.ErrorF("error un-escaping url path: %v", e)
			} else {
				query = vv
			}
			query = forms.Sanitize(query)
			query = html.UnescapeString(query)
			// query = html.EscapeString(query)
			// query = url.PathEscape(query)
			foundQuery = true
			if foundNonce {
				break
			}
		}
	}
	if foundQuery && !foundNonce {
		// Let the user know that the search form submission resulted in a (generic) error
		err = fmt.Errorf(printer.Sprintf("search form error"))
		return
	}
	if ok = foundQuery && foundNonce; ok {
		var dst string
		query = url.PathEscape(query)
		if !language.Compare(tag, f.enjin.SiteDefaultLanguage()) {
			dst = "/" + tag.String() + f.path + "/" + query
		} else {
			dst = f.path + "/" + query
		}
		log.DebugF("search redirecting: %v", dst)
		http.Redirect(w, r, dst, http.StatusSeeOther)
	}
	return
}

func (f *CFeature) Use(s feature.System) feature.MiddlewareFn {
	log.DebugF("including page search middleware")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := forms.SanitizeRequestPath(r.URL.Path)
			if _, p, ok := lang.ParseLangPath(path); ok {
				path = p
			}
			if strings.HasPrefix(path, f.path) {
				// this is the search page being requested
				// POST requests need to redirect to f.path with query appended instead of search

				var searchPage *page.Page
				reqLangTag := lang.GetTag(r)
				if searchPage = s.FindPage(reqLangTag, f.path); searchPage == nil {
					if searchPage = s.FindPage(f.enjin.SiteDefaultLanguage(), f.path); searchPage == nil {
						if searchPage = s.FindPage(language.Und, f.path); searchPage == nil {
							log.ErrorF("search path not found: [%v] \"%v\"", reqLangTag, f.path)
							s.ServeInternalServerError(w, r)
							return
						}
					}
				}

				size := 10
				pg := 0
				fPathLen := len(f.path)

				var input string
				if len(path) > fPathLen+1 {
					if input = path[fPathLen+1:]; input != "" {
						if rxQueryPaginationSizePage.MatchString(input) {
							m := rxQueryPaginationSizePage.FindAllStringSubmatch(input, 1)
							input = m[0][1]
							if v, e := strconv.Atoi(m[0][2]); e == nil {
								size = v
							}
							if v, e := strconv.Atoi(m[0][3]); e == nil {
								pg = v
							}
						} else if rxQueryPaginationPage.MatchString(input) {
							m := rxQueryPaginationPage.FindAllStringSubmatch(input, 1)
							input = m[0][1]
							if v, e := strconv.Atoi(m[0][2]); e == nil {
								pg = v
							}
						}
					}
				}

				var queryError bool
				if input == "" {
					if ok, err := f.handleQueryRedirect(w, r); ok && err == nil {
						return
					} else if err != nil {
						searchPage.Context.SetSpecific("SiteSearchError", err.Error())
						queryError = true
					}
				}

				if !queryError && input != "" {
					if v, e := url.PathUnescape(input); e == nil {
						input = v
					} else {
						log.ErrorF("error un-escaping path url: %v - %v", input, e)
					}
					// input = html.UnescapeString(input)
					if i, err := url.QueryUnescape(input); err != nil {
						log.ErrorF("error un-escaping input query string: %v", err)
					} else {
						input = i
					}
					input = forms.StripTags(input)
					searchPage.Context.SetSpecific("SiteSearchQuery", input)
					if results, err := f.PerformSearch(reqLangTag, input, size, pg); err != nil {
						searchPage.Context.SetSpecific("SiteSearchError", err.Error())
					} else {
						searchPage.Context.SetSpecific("SiteSearchResults", results)
						searchPage.Context.SetSpecific("SiteSearchSize", size)
						searchPage.Context.SetSpecific("SiteSearchPage", pg)
						pages := results.Total / uint64(size)
						searchPage.Context.SetSpecific("SiteSearchPages", pages)
						numHits := len(results.Hits)
						idStart := (pg * size) + 1
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
							pageSummary = printer.Sprintf("Page %d of %d", pg+1, pages+1)
							// Search hits summary with only one hit, <hit-number> of <total-hits>
							hitsSummary = printer.Sprintf("Showing #%d of %d", idStart, results.Total)
						default:
							// Search page summary, <number> of <pages>
							pageSummary = printer.Sprintf("Page %d of %d", pg+1, pages+1)
							// Search hits summary with more than one hit, <first-hit-number>-<last-hit-number> of <total-hits>
							hitsSummary = printer.Sprintf("Showing %d-%d of %d", idStart, idEnd, results.Total)
						}
						searchPage.Context.SetSpecific("SiteSearchPageSummary", template.HTML(pageSummary))
						searchPage.Context.SetSpecific("SiteSearchHitsSummary", template.HTML(hitsSummary))
						searchPage.Context.SetSpecific("SiteSearchResultSummary", template.HTML(resultSummary))
					}
				}

				if err := s.ServePage(searchPage, w, r); err != nil {
					log.ErrorF("error serving search page: %v", err)
					s.Serve500(w, r)
				}

				return
			}

			if ok, _ := f.handleQueryRedirect(w, r); ok {
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (f *CFeature) FindRedirection(_ string) (p *page.Page) {
	// nop for search page redirects
	return
}

func (f *CFeature) FindTranslations(path string) (found []*page.Page) {
	if strings.HasPrefix(path, f.path) {
		for _, tag := range f.enjin.SiteLocales() {
			if pg := f.FindPage(tag, path); pg != nil {
				found = append(found, pg)
			}
		}
	}
	return
}

func (f *CFeature) FindPages(prefix string) (pages []*page.Page) {
	if strings.HasPrefix(f.path, prefix) {
		if pg := f.FindPage(f.enjin.SiteDefaultLanguage(), f.path); pg != nil {
			pages = append(pages, pg)
		}
	}
	return
}

func (f *CFeature) FindPage(tag language.Tag, path string) (searchPage *page.Page) {
	if f.findingSelf == true {
		return
	}
	f.Lock()
	defer f.Unlock()
	if strings.HasPrefix(path, f.path) {
		// using f.path for actual page-finding due to the variability of the search endpoint
		// actual page begins with the configured search path
		f.findingSelf = true
		if searchPage = f.enjin.FindPage(tag, f.path); searchPage == nil {
			if searchPage = f.enjin.FindPage(language.Und, f.path); searchPage == nil {
				log.ErrorF("search page not found: [%v] \"%v\"", tag, f.path)
				return
			}
		}
		f.findingSelf = false
	}
	return
}

func (f *CFeature) MatchQL(query string) (pages []*page.Page) {
	// search cannot be found with MatchQL?
	return
}