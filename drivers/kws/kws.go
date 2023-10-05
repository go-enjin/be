//go:build driver_kws || drivers || all

// Copyright (c) 2023  The Go-Enjin Authors
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

	"github.com/blevesearch/bleve/v2"
	bleveSearch "github.com/blevesearch/bleve/v2/search"
	"github.com/fvbommel/sortorder"
	"github.com/maruel/natural"
	"github.com/puzpuzpuz/xsync/v2"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/feature"
	indexingSearch "github.com/go-enjin/be/pkg/indexing/search"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/regexps"
	"github.com/go-enjin/be/pkg/search"
	"github.com/go-enjin/be/types/page"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "PagesIndexingKeyWordSearch"

type Feature interface {
	feature.Feature
	feature.KeywordProvider
	feature.SearchEnjinFeature
}

type CFeature struct {
	feature.CFeature

	keyword *xsync.MapOf[string, []string]
}

type MakeFeature interface {
	Make() Feature
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.keyword = xsync.NewMapOf[[]string]()
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Setup(enjin feature.Internals) {
	f.CFeature.Setup(enjin)
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	err = f.CFeature.Startup(ctx)
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
	var t feature.Theme
	if t, err = f.Enjin.GetTheme(); err != nil {
		return
	}
	langMode := f.Enjin.SiteLanguageMode()
	fallback := f.Enjin.SiteDefaultLanguage()

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
	matches := make(map[string]*feature.PageStub)

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
		if stubs, ok := f.keyword.Load(word); ok {
			for _, shasum := range stubs {
				notStubs[shasum] = true
			}
		}
	}

	numMustWords := len(mustWords)
	if numMustWords > 0 {
		mustMatch := make(map[string]feature.PageStubs)
		mustCache := make(map[string]map[string]int)
		for word, _ := range mustWords {
			if shasums, ok := f.keyword.Load(word); ok {
				// log.WarnF("found %d stubs for %v", len(stubs), word)
				for _, shasum := range shasums {
					if _, not := notStubs[shasum]; not {
						continue
					}
					if _, exists := mustCache[shasum]; !exists {
						mustCache[shasum] = make(map[string]int)
					}
					mustCache[shasum][word] += 1
					if len(mustCache[shasum]) == numMustWords {
						// log.WarnF("stub has all words: %v - %v", stub.Source, mustWords)
						if stub := f.Enjin.FindPageStub(shasum); stub != nil {
							mustMatch[word] = append(mustMatch[word], stub)
						} else {
							log.ErrorF("error finding page stub by shasum: %v", shasum)
						}
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
			if shasums, ok := f.keyword.Load(word); ok {
				for _, shasum := range shasums {
					if _, exists := scores[shasum]; exists {
						scores[shasum] += shouldScores[word]
					}
				}
			}
		}

		// log.WarnF("mustMatch: %v", mustMatch)

	} else {
		// no must words present
		for word, _ := range shouldWords {
			if shasums, ok := f.keyword.Load(word); ok {
				for _, shasum := range shasums {
					if _, not := notStubs[shasum]; not {
						continue
					}
					matches[shasum] = f.Enjin.FindPageStub(shasum)
					scores[shasum] += shouldScores[word]
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
				if p, ee := page.NewPageFromStub(stub, t); ee == nil {
					id := langMode.ToUrl(fallback, p.LanguageTag(), p.Url())
					hit := &bleveSearch.DocumentMatch{
						Index:     id,
						ID:        id,
						Score:     scores[shasum],
						HitNumber: count + 1,
						Fields: map[string]interface{}{
							"url":         id,
							"title":       p.Title(),
							"description": p.Description(),
						},
						Fragments: map[string][]string{
							"summary": {p.Description()},
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

func (f *CFeature) AddToSearchIndex(stub *feature.PageStub, p feature.Page) (err error) {
	f.Lock()
	defer f.Unlock()
	var doc search.Document
	if doc, err = indexingSearch.SearchDocument(p, f.Enjin.MustGetTheme()); err != nil {
		log.ErrorF("error creating page search.Document: %v", err)
		return
	} else if doc == nil {
		return
	}
	for _, content := range doc.GetContents() {
		words := regexps.RxKeywords.FindAllString(content, -1)
		for _, word := range words {
			lcw := strings.ToLower(word)
			//f.keyword[lcw] = append(f.keyword[lcw], stub.Shasum)
			shasums, _ := f.keyword.Load(lcw)
			shasums = append(shasums, stub.Shasum)
			f.keyword.Store(lcw, shasums)
		}
	}
	return
}

func (f *CFeature) RemoveFromSearchIndex(stub *feature.PageStub, p feature.Page) {
	//f.Lock()
	//defer f.Unlock()
	log.WarnF("%v feature does not support removing pages from the keywords index", f.Tag())
	return
}

func (f *CFeature) Size() (count int) {
	return f.keyword.Size()
}

func (f *CFeature) Range(fn func(keyword string, shasums []string) (proceed bool)) {
	f.keyword.Range(fn)
}

func (f *CFeature) KnownKeywords() (keywords []string) {
	//f.RLock()
	//defer f.RUnlock()
	f.keyword.Range(func(key string, value []string) bool {
		keywords = append(keywords, key)
		return true
	})
	sort.Sort(sortorder.Natural(keywords))
	return
}

func (f *CFeature) KeywordStubs(keyword string) (stubs feature.PageStubs) {
	//f.RLock()
	//defer f.RUnlock()
	shasums, _ := f.keyword.Load(keyword)
	for _, shasum := range shasums {
		if stub := f.Enjin.FindPageStub(shasum); stub != nil {
			stubs = append(stubs, stub)
		} else {
			log.ErrorF("error finding page stub by shasum: %v", shasum)
		}
	}
	return
}