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

package pages

import (
	"fmt"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/blevesearch/bleve/v2/search/highlight/format/html"

	"github.com/go-enjin/golang-org-x-text/language"

	indexing "github.com/go-enjin/be/pkg/indexing/search"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/page"
	"github.com/go-enjin/be/pkg/regexps"
	"github.com/go-enjin/be/pkg/search"
)

// TODO: update SearchWithin to use pagecache.SearchEnjinFeature

func SearchWithin(input string, numPerPage, pageNumber int, pages []*page.Page, defaultLang, tag language.Tag, langMode lang.Mode) (matches map[string]*page.Page, results *bleve.SearchResult, err error) {
	var locales []language.Tag
	lookupPage := make(map[string]*page.Page)

	docMaps := make(map[string]*mapping.DocumentMapping)
	for _, pg := range pages {
		if !lang.TagInTagSlices(pg.LanguageTag, locales) {
			locales = append(locales, pg.LanguageTag)
		}
		if _, ok := docMaps[pg.Format]; !ok {
			if doctype, dm, e := indexing.SearchMapping(pg); e != nil {
				err = fmt.Errorf("error getting page search mapping: %v - %v", pg.Url, e)
				return
			} else if dm != nil {
				docMaps[doctype] = dm
			}
		}
	}

	localeIndexed := make(map[language.Tag]bleve.Index)
	for _, locale := range locales {
		if localeIndexed[locale], err = search.NewMemOnlyIndexWithDocMaps(locale, docMaps); err != nil {
			return
		}
	}

	for _, pg := range pages {
		if doc, e := indexing.SearchDocument(pg); e != nil {
			err = fmt.Errorf("error preparing search document: %v - %v", pg.Url, e)
			return
		} else if doc != nil {
			key := langMode.ToUrl(defaultLang, pg.LanguageTag, pg.Url)
			if ee := localeIndexed[pg.LanguageTag].Index(key, doc.Self()); ee != nil {
				err = fmt.Errorf("error indexing search document: %v - %v", pg.Url, ee)
				return
			}
			lookupPage[key] = pg
		}
	}

	var list []bleve.Index
	for _, idx := range localeIndexed {
		list = append(list, idx)
	}
	allIndexed := bleve.NewIndexAlias(list...)

	var searchAll bool
	inputWantsTag := defaultLang

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
			if found := lang.TagInTags(queryLangTag, locales...); !found {
				err = fmt.Errorf("unsupported language: %v", queryLangTag)
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
	if numPerPage == 0 {
		numPerPage = 10
	}
	req.Size = numPerPage
	req.From = pageNumber * numPerPage
	req.Fields = []string{"*"}
	req.Highlight = bleve.NewHighlightWithStyle(html.Name)

	// determine which index to search
	var index bleve.Index = allIndexed
	if !searchAll {
		if !language.Compare(inputWantsTag, language.Und) {
			if idx, ok := localeIndexed[inputWantsTag]; ok {
				index = idx
			}
		}
		if index == allIndexed && !language.Compare(tag, language.Und) {
			if idx, ok := localeIndexed[tag]; ok {
				index = idx
			}
		}
	}

	if results, err = index.Search(req); err != nil {
		return
	}

	matches = make(map[string]*page.Page)
	for _, hit := range results.Hits {
		if pg, ok := lookupPage[hit.ID]; ok {
			matches[hit.ID] = pg
		}
	}
	return
}