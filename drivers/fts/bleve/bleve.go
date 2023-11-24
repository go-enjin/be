//go:build driver_fts_bleve || drivers_fts || bleve || all

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

package bleve

import (
	"fmt"
	"net/url"
	"sync"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	bleveFormatHtml "github.com/blevesearch/bleve/v2/search/highlight/format/html"
	"github.com/urfave/cli/v2"
	"golang.org/x/net/html"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	beIndexSearch "github.com/go-enjin/be/pkg/indexing/search"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/regexps"
	beSearch "github.com/go-enjin/be/pkg/search"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "drivers-fts-bleve"

type Feature interface {
	feature.Feature
	feature.SearchEnjinFeature
}

type CFeature struct {
	feature.CFeature

	indexes map[language.Tag]bleve.Index
	docMaps map[language.Tag]map[string]*mapping.DocumentMapping

	sync.RWMutex
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
	f.indexes = make(map[language.Tag]bleve.Index)
	f.docMaps = make(map[language.Tag]map[string]*mapping.DocumentMapping)
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Setup(enjin feature.Internals) {
	f.CFeature.Setup(enjin)
	locales := f.Enjin.SiteLocales()
	for _, sdm := range feature.FilterTyped[feature.SearchDocumentMapperFeature](f.Enjin.Features().List()) {
		for _, tag := range locales {
			if _, exists := f.docMaps[tag]; !exists {
				f.docMaps[tag] = make(map[string]*mapping.DocumentMapping)
			}
			doctype, dm := sdm.SearchDocumentMapping(tag)
			f.docMaps[tag][doctype] = dm
		}
	}
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	return
}

func (f *CFeature) PrepareSearch(tag language.Tag, input string) (query string) {
	query = input
	return
}

func (f *CFeature) PerformSearch(tag language.Tag, input string, size, pg int) (results *bleve.SearchResult, err error) {
	f.RLock()
	defer f.RUnlock()

	var list []bleve.Index
	for _, index := range f.indexes {
		list = append(list, index)
	}
	all := bleve.NewIndexAlias(list...)

	searchAll := false
	inputWantsTag := language.Und
	input = forms.StrictSanitize(input)
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
			for _, siteLocale := range f.Enjin.SiteLocales() {
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
	var index bleve.Index = all
	if !searchAll {
		if !language.Compare(inputWantsTag, language.Und) {
			if idx, ok := f.indexes[inputWantsTag]; ok {
				index = idx
			}
		}
		if index == all && !language.Compare(tag, language.Und) {
			if idx, ok := f.indexes[tag]; ok {
				index = idx
			}
		}
	}

	results, err = index.Search(req)
	return
}

func (f *CFeature) AddToSearchIndex(stub *feature.PageStub, p feature.Page) (err error) {

	if p.Context().Bool("NoPageIndexing", false) || p.Context().Bool("NoSearchIndexing", false) {
		return
	}

	f.Lock()
	defer f.Unlock()
	if f.indexes == nil {
		f.indexes = make(map[language.Tag]bleve.Index)
	}
	var ok bool
	var index bleve.Index
	if index, ok = f.indexes[p.LanguageTag()]; !ok {
		if index, err = beSearch.NewMemOnlyIndexWithDocMaps(p.LanguageTag(), f.docMaps[p.LanguageTag()]); err != nil {
			return
		}
		f.indexes[p.LanguageTag()] = index
	}
	var doc beSearch.Document
	if doc, err = beIndexSearch.SearchDocument(p, f.Enjin.MustGetTheme()); err != nil {
		return
	} else if doc == nil {
		return
	}
	pgUrl := p.Url()
	langMode := f.Enjin.SiteLanguageMode()
	fallback := f.Enjin.SiteDefaultLanguage()
	if !language.Compare(p.LanguageTag(), fallback, language.Und) {
		pgUrl = langMode.ToUrl(fallback, p.LanguageTag(), p.Url())
	}
	if err = index.Index(pgUrl, doc.Self()); err != nil {
		return
	}
	return
}

func (f *CFeature) RemoveFromSearchIndex(stub *feature.PageStub, p feature.Page) {
	f.Lock()
	defer f.Unlock()
	if index, ok := f.indexes[p.LanguageTag()]; ok {
		pgUrl := p.Url()
		langMode := f.Enjin.SiteLanguageMode()
		fallback := f.Enjin.SiteDefaultLanguage()
		if !language.Compare(p.LanguageTag(), fallback, language.Und) {
			pgUrl = langMode.ToUrl(fallback, p.LanguageTag(), p.Url())
		}
		if err := index.Delete(pgUrl); err != nil {
			log.ErrorF("error removing bleve index: %v - %v", p.Url(), err)
		}
	}
	return
}
