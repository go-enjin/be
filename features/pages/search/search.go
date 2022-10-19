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
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/blevesearch/bleve/v2"
	bleveHtml "github.com/blevesearch/bleve/v2/search/highlight/format/html"
	"github.com/urfave/cli/v2"
	"golang.org/x/net/html"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/forms/nonce"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net"
	"github.com/go-enjin/be/pkg/page"
	beSearch "github.com/go-enjin/be/pkg/search/indexes"
)

var _ MakeFeature = (*CFeature)(nil)
var _ feature.Middleware = (*CFeature)(nil)

const Tag feature.Tag = "PagesSearch"

type Feature interface {
	feature.Middleware
}

type CFeature struct {
	feature.CMiddleware

	cli   *cli.Context
	enjin feature.Internals

	path string
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

	if results, e := f.PerformSearch(input, size, pg); e != nil {
		err = e
	} else {
		fmt.Printf("%v\n", results)
	}
	return
}

func (f *CFeature) PerformSearch(input string, size, pg int) (results *bleve.SearchResult, err error) {
	if index, e := beSearch.NewFeaturesIndex(f.enjin); e != nil {
		err = e
		return
	} else {
		input = forms.StripTags(input)
		log.DebugF("performing site search: %v", input)
		query := bleve.NewQueryStringQuery(input)
		if err = query.Validate(); err != nil {
			return
		}
		req := bleve.NewSearchRequest(query)
		if size == 0 {
			size = 10
		}
		req.Size = size
		req.From = pg * size
		req.Fields = []string{"*"}
		req.Highlight = bleve.NewHighlightWithStyle(bleveHtml.Name)
		if results, e = index.Search(req); e != nil {
			err = e
			return
		}
	}
	return
}

func (f *CFeature) FilterPageContext(themeCtx, pageCtx context.Context, r *http.Request) (out context.Context) {
	out = themeCtx
	out.SetSpecific("SiteSearchPath", f.path)
	return
}

var rxQueryPaginationPage = regexp.MustCompile(`(.+)/(\d+)/??$`)
var rxQueryPaginationSizePage = regexp.MustCompile(`(.+)/(\d+)/(\d+)/??$`)

func (f *CFeature) Use(s feature.System) feature.MiddlewareFn {
	log.DebugF("including page search middleware")

	handleQueryRedirect := func(w http.ResponseWriter, r *http.Request) (ok bool, err error) {
		var query string
		var foundNonce, foundQuery bool
		for k, v := range r.URL.Query() {
			switch k {
			case "nonce":
				value := forms.StripTags(v[0])
				if vv, e := url.PathUnescape(value); e != nil {
					log.ErrorF("error un-escaping url path: %v", e)
				} else {
					value = vv
				}
				value = html.UnescapeString(value)
				if !nonce.Validate("site-search-form", value) {
					err = fmt.Errorf("search form expired")
					return
				}
				foundNonce = true
				if foundQuery {
					break
				}
			case "query":
				// trap random page ?query= requests?
				query = forms.StripTags(v[0])
				if vv, e := url.PathUnescape(query); e != nil {
					log.ErrorF("error un-escaping url path: %v", e)
				} else {
					query = vv
				}
				query = html.EscapeString(query)
				foundQuery = true
				if foundNonce {
					break
				}
			}
		}
		if foundQuery && !foundNonce {
			err = fmt.Errorf("search form error")
		} else {
			if ok = foundQuery && foundNonce; ok {
				http.Redirect(w, r, f.path+"/"+query, http.StatusSeeOther)
			}
		}
		return
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := net.TrimQueryParams(r.URL.Path)

			if strings.HasPrefix(path, f.path) {
				// this is the search page being requested
				// POST requests need to redirect to f.path with query appended instead of search

				var searchPage *page.Page
				if searchPage = s.FindPage(f.path); searchPage == nil {
					log.ErrorF("search path not found: \"%v\"", f.path)
					s.Serve500(w, r)
					return
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
					if ok, err := handleQueryRedirect(w, r); ok && err == nil {
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
					input = html.UnescapeString(input)
					input = forms.StripTags(input)
					if results, err := f.PerformSearch(input, size, pg); err != nil {
						searchPage.Context.SetSpecific("SiteSearchError", err.Error())
					} else {
						searchPage.Context.SetSpecific("SiteSearchQuery", input)
						searchPage.Context.SetSpecific("SiteSearchResults", results)
						searchPage.Context.SetSpecific("SiteSearchSize", size)
						searchPage.Context.SetSpecific("SiteSearchPage", pg)
						pages := results.Total / uint64(size)
						searchPage.Context.SetSpecific("SiteSearchPages", pages)
					}
				}

				if err := s.ServePage(searchPage, w, r); err != nil {
					log.ErrorF("error serving search page: %v", err)
					s.Serve500(w, r)
				}

				return
			}

			if ok, _ := handleQueryRedirect(w, r); ok {
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}