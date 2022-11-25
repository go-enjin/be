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

package kws

import (
	"sort"
	"strings"
	"sync"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	bleveSearch "github.com/blevesearch/bleve/v2/search"
	"github.com/go-enjin/golang-org-x-text/language"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
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

const Tag feature.Tag = "PagesSearchKeyword"

type Feature interface {
	feature.Feature
}

type CFeature struct {
	feature.CFeature

	cli   *cli.Context
	enjin feature.Internals

	keyword map[string][]*pagecache.Stub
	docMaps map[language.Tag]map[string]*mapping.DocumentMapping

	pql pagecache.QueryEnjinFeature

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
	f.keyword = make(map[string][]*pagecache.Stub)
}

func (f *CFeature) Tag() (tag feature.Tag) {
	tag = Tag
	return
}

func (f *CFeature) Setup(enjin feature.Internals) {
	f.enjin = enjin
	locales := f.enjin.SiteLocales()
	for _, feat := range f.enjin.Features() {
		if v, ok := feat.Self().(pagecache.QueryEnjinFeature); ok {
			if f.pql != nil {
				log.FatalF("only one pagecache.QueryEnjinFeature per enjin allowed")
			}
			f.pql = v
		} else if v, ok := feat.Self().(pagecache.SearchDocumentMapperFeature); ok {
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

func (f *CFeature) PerformSearch(tag language.Tag, input string, size, pg int) (results *bleve.SearchResult, err error) {
	var t *theme.Theme
	if t, err = f.enjin.GetTheme(); err != nil {
		return
	}
	langMode := f.enjin.SiteLanguageMode()
	fallback := f.enjin.SiteDefaultLanguage()

	cleaned := regexps.RxKeywordTrims.ReplaceAllString(input, "")
	keywords := regexps.RxEmptySpace.Split(cleaned, -1)
	matches := make(map[string]*pagecache.Stub)
	scores := make(map[string]float64)
	numKeywords := len(keywords)

	kwValues := map[string]float64{}
	for idx, word := range keywords {
		baseValue := 1.0 / float64(numKeywords)
		multiplier := float64(numKeywords - idx)
		kwValues[word] = multiplier * (baseValue)
	}

	for _, word := range keywords {
		if stubs, ok := f.keyword[word]; ok {
			for _, stub := range stubs {
				matches[stub.Shasum] = stub
				if _, ok := scores[stub.Shasum]; !ok {
					scores[stub.Shasum] = 0
				}
				scores[stub.Shasum] += kwValues[word]
			}
		}
	}

	var sorted []string
	for shasum, _ := range scores {
		sorted = append(sorted, shasum)
	}
	sort.Slice(sorted, func(i, j int) (less bool) {
		a, b := sorted[i], sorted[j]
		if scores[a] == scores[b] {
			less = a < b
		} else {
			less = scores[a] > scores[b]
		}
		return
	})

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
	var doc search.Document
	if doc, err = pagecache.SearchDocument(p); err != nil {
		log.ErrorF("error creating page search.Document: %v", err)
		return
	} else if doc == nil {
		return
	}
	if f.keyword == nil {
		f.keyword = make(map[string][]*pagecache.Stub)
	}
	for _, content := range doc.GetContents() {
		cleaned := regexps.RxKeywordTrims.ReplaceAllString(content, "")
		words := regexps.RxEmptySpace.Split(cleaned, -1)
		for _, word := range words {
			lcw := strings.ToLower(word)
			f.keyword[lcw] = append(f.keyword[lcw], stub)
		}
	}
	if f.pql != nil {
		err = f.pql.AddToQueryIndex(stub, p)
	}
	return
}

func (f *CFeature) RemoveFromSearchIndex(tag language.Tag, file, shasum string) {
	// panic("implement me")
	return
}