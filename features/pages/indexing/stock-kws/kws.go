//go:build stock_kws || pages || all

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

package kws

import (
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	bleveSearch "github.com/blevesearch/bleve/v2/search"
	"github.com/go-enjin/golang-org-x-text/language"
	"github.com/maruel/natural"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/page"
	"github.com/go-enjin/be/pkg/pagecache"
	"github.com/go-enjin/be/pkg/regexps"
	"github.com/go-enjin/be/pkg/search"
	"github.com/go-enjin/be/pkg/theme"
)

var (
	_ Feature                      = (*CFeature)(nil)
	_ MakeFeature                  = (*CFeature)(nil)
	_ pagecache.SearchEnjinFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "PagesIndexingKeyWordSearch"

type Feature interface {
	feature.Feature
}

type CFeature struct {
	feature.CFeature

	cli   *cli.Context
	enjin feature.Internals

	keyword map[string]pagecache.Stubs
	docMaps map[language.Tag]map[string]*mapping.DocumentMapping

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
	f.CFeature.Init(this)
	f.docMaps = make(map[language.Tag]map[string]*mapping.DocumentMapping)
	f.keyword = make(map[string]pagecache.Stubs)
}

func (f *CFeature) Tag() (tag feature.Tag) {
	tag = Tag
	return
}

func (f *CFeature) Setup(enjin feature.Internals) {
	f.enjin = enjin
	locales := f.enjin.SiteLocales()
	for _, feat := range f.enjin.Features() {
		if v, ok := feat.Self().(pagecache.SearchDocumentMapperFeature); ok {
			for _, tag := range locales {
				if _, exists := f.docMaps[tag]; !exists {
					f.docMaps[tag] = make(map[string]*mapping.DocumentMapping)
				}
				doctype, dm := v.SearchDocumentMapping(tag)
				f.docMaps[tag][doctype] = dm
			}
		}
	}
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	f.cli = ctx
	return
}

var RxKeywords = regexp.MustCompile(`([-+]?(?:` + regexps.KeywordPattern + `))`)

func (f *CFeature) PrepareSearch(tag language.Tag, input string) (query string) {
	keywords := RxKeywords.FindAllString(input, -1)
	for idx, keyword := range keywords {
		keyword = strings.ToLower(keyword)
		if idx > 0 {
			query += " "
		}
		query += keyword
	}
	return
}

func (f *CFeature) PerformSearch(tag language.Tag, input string, size, pg int) (results *bleve.SearchResult, err error) {
	f.RLock()
	defer f.RUnlock()
	var t *theme.Theme
	if t, err = f.enjin.GetTheme(); err != nil {
		return
	}
	langMode := f.enjin.SiteLanguageMode()
	fallback := f.enjin.SiteDefaultLanguage()

	keywords := RxKeywords.FindAllString(input, -1)
	mustWords, shouldWords, notWords := make(map[string]int), make(map[string]int), make(map[string]int)
	for idx, keyword := range keywords {
		if keyword = strings.ToLower(keyword); keyword != "" {
			keywords[idx] = keyword
			switch keyword[0] {
			case '+':
				word := keyword[1:]
				mustWords[word] = idx
			case '-':
				word := keyword[1:]
				notWords[word] = idx
			default:
				shouldWords[keyword] = idx
			}
		}
	}

	scores := make(map[string]float64)
	matches := make(map[string]*pagecache.Stub)

	// pre-calculate word presence scores

	numKeywords := len(keywords)
	baseValue := 1.0 / float64(numKeywords)

	shouldScores := make(map[string]float64)
	for word, idx := range shouldWords {
		multiplier := float64(numKeywords - idx)
		shouldScores[word] = multiplier * (baseValue)
	}
	mustScores := make(map[string]float64)
	for word, idx := range mustWords {
		multiplier := float64(numKeywords - idx)
		mustScores[word] = multiplier * (baseValue + 100.0)
	}
	notScores := make(map[string]float64)
	for word, _ := range notWords {
		notScores[word] = -1000.0
	}
	notStubs := make(map[string]bool)
	for word, _ := range notWords {
		if stubs, ok := f.keyword[word]; ok {
			for _, stub := range stubs {
				notStubs[stub.Shasum] = true
			}
		}
	}

	numMustWords := len(mustWords)
	if numMustWords > 0 {
		mustMatch := make(map[string]pagecache.Stubs)
		mustCache := make(map[string]map[string]int)
		for word, _ := range mustWords {
			if stubs, ok := f.keyword[word]; ok {
				// log.WarnF("found %d stubs for %v", len(stubs), word)
				for _, stub := range stubs {
					if _, not := notStubs[stub.Shasum]; not {
						continue
					}
					if _, exists := mustCache[stub.Shasum]; !exists {
						mustCache[stub.Shasum] = make(map[string]int)
					}
					mustCache[stub.Shasum][word] += 1
					if len(mustCache[stub.Shasum]) == numMustWords {
						// log.WarnF("stub has all words: %v - %v", stub.Source, mustWords)
						mustMatch[word] = append(mustMatch[word], stub)
					}
				}
			}
		}

		for word, stubs := range mustMatch {
			for _, stub := range stubs {
				matches[stub.Shasum] = stub
				scores[stub.Shasum] += mustScores[word] * float64(numMustWords)
			}
		}

		for word, _ := range shouldWords {
			if stubs, ok := f.keyword[word]; ok {
				for _, stub := range stubs {
					if _, exists := scores[stub.Shasum]; exists {
						scores[stub.Shasum] += shouldScores[word]
					}
				}
			}
		}

		// log.WarnF("mustMatch: %v", mustMatch)

	} else {
		// no must words present
		for word, _ := range shouldWords {
			if stubs, ok := f.keyword[word]; ok {
				for _, stub := range stubs {
					if _, not := notStubs[stub.Shasum]; not {
						continue
					}
					matches[stub.Shasum] = stub
					scores[stub.Shasum] += shouldScores[word]
				}
			}
		}
	}

	// sort results based on score

	var sorted []string
	for shasum, _ := range scores {
		sorted = append(sorted, shasum)
	}
	sort.Slice(sorted, func(i, j int) (less bool) {
		a, b := sorted[i], sorted[j]
		if scores[a] == scores[b] {
			less = natural.Less(a, b)
		} else {
			less = scores[a] > scores[b]
		}
		return
	})

	// prepare return values

	var maxScore float64
	var hits []*bleveSearch.DocumentMatch
	if size > 0 && pg >= 0 {
		var count uint64 = 0
		var start = uint64(pg * size)
		var end = start + uint64(size)
		for idx, shasum := range sorted {
			count = uint64(idx)
			stub := matches[shasum]
			if maxScore < scores[shasum] {
				maxScore = scores[shasum]
			}
			if count >= start && count < end {
				if p, ee := stub.Make(t); ee == nil {
					id := langMode.ToUrl(fallback, p.LanguageTag, p.Url)
					hit := &bleveSearch.DocumentMatch{
						Index:     id,
						ID:        id,
						Score:     scores[shasum],
						HitNumber: count + 1,
						Fields: map[string]interface{}{
							"url":         id,
							"title":       p.Title,
							"description": p.Description,
						},
						Fragments: map[string][]string{
							"summary": {p.Description},
						},
					}
					hits = append(hits, hit)
				}
			}
		}
	}

	total := len(sorted)

	// TODO: populate bleve.SearchResult as much as possible
	results = &bleve.SearchResult{
		Status: &bleve.SearchStatus{
			Total:      total,
			Failed:     0,
			Successful: total,
		},
		Hits:     hits,
		Total:    uint64(total),
		Request:  nil,
		MaxScore: maxScore,
	}
	return
}

func (f *CFeature) AddToSearchIndex(stub *pagecache.Stub, p *page.Page) (err error) {
	f.Lock()
	defer f.Unlock()
	var doc search.Document
	if doc, err = pagecache.SearchDocument(p); err != nil {
		log.ErrorF("error creating page search.Document: %v", err)
		return
	} else if doc == nil {
		return
	}
	if f.keyword == nil {
		f.keyword = make(map[string]pagecache.Stubs)
	}
	for _, content := range doc.GetContents() {
		words := regexps.RxKeywords.FindAllString(content, -1)
		for _, word := range words {
			lcw := strings.ToLower(word)
			f.keyword[lcw] = append(f.keyword[lcw], stub)
		}
	}
	return
}

func (f *CFeature) RemoveFromSearchIndex(tag language.Tag, file, shasum string) {
	f.Lock()
	defer f.Unlock()
	return
}

func (f *CFeature) KnownKeywords() (keywords []string) {
	f.RLock()
	defer f.RUnlock()
	keywords = maps.SortedKeys(f.keyword)
	return
}

func (f *CFeature) KeywordStubs(keyword string) (stubs pagecache.Stubs) {
	f.RLock()
	defer f.RUnlock()
	stubs = f.keyword[keyword]
	return
}