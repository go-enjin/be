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
	"net/url"

	"github.com/blevesearch/bleve/v2"
	bleveFormatHtml "github.com/blevesearch/bleve/v2/search/highlight/format/html"
	"github.com/go-enjin/golang-org-x-text/language"
	"github.com/urfave/cli/v2"
	"golang.org/x/net/html"

	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/regexps"
	beSearchIndexes "github.com/go-enjin/be/pkg/search/indexes"
)

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

func (f *CFeature) PerformSearch(tag language.Tag, input string, size, pg int) (results *bleve.SearchResult, err error) {
	if indexes, all, e := beSearchIndexes.NewFeaturesIndex(f.enjin); e != nil {
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
		if regexps.RxLanguageKey.MatchString(input) {
			m := regexps.RxLanguageKey.FindAllStringSubmatch(input, 1)
			if m[0][1] == "*" {
				searchAll = true
				input = regexps.RxLanguageKey.ReplaceAllString(input, "")
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
				input = regexps.RxLanguageKey.ReplaceAllString(input, "")
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
		req.Highlight = bleve.NewHighlightWithStyle(bleveFormatHtml.Name)

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