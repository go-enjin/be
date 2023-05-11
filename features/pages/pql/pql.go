//go:build page_pql || pages || all

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

package pql

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-enjin/golang-org-x-text/language"
	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/kvs"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/page"
	"github.com/go-enjin/be/pkg/page/matter"
	"github.com/go-enjin/be/pkg/pagecache"
	"github.com/go-enjin/be/pkg/pageql/matcher"
	"github.com/go-enjin/be/pkg/pageql/selector"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

var (
	AlwaysIgnorePageIndexKeys = []string{"content", "frontmatter"}
	DefaultPageIndexKeys      = []string{"type", "language", "url", "title", "description"}
)

const (
	gAllUrlsBucketName           = "page_urls_all"
	gPageUrlsBucketName          = "page_urls"
	gPageStubsBucketName         = "page_stubs"
	gPageRedirectionsBucketName  = "page_redirections"
	gPageTranslationsBucketName  = "page_translations"
	gPageTranslatedByBucketName  = "page_translated_by"
	gPageContextValuesBucketName = "page_context_values"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "pages-pql"

type Feature interface {
	feature.Feature
	feature.PageProvider
	pagecache.PageIndexFeature
	pagecache.QueryIndexFeature
	pagecache.PageContextProvider
}

type MakeFeature interface {
	Make() Feature

	SetKVC(tag feature.Tag, name string) MakeFeature
}

type CFeature struct {
	feature.CFeature

	kvcTag  feature.Tag
	kvcName string

	cache kvs.KeyValueCache

	indexPageKeys  []string
	ignorePageKeys []string

	allUrlsBucket           kvs.KeyValueStore
	pageUrlsBucket          kvs.KeyValueStore
	pageStubsBucket         kvs.KeyValueStore
	redirectionsBucket      kvs.KeyValueStore
	translatedByBucket      kvs.KeyValueStore
	translationsBucket      map[language.Tag]kvs.KeyValueStore
	contextValueKeyedBucket kvs.KeyValueStore

	// index map[string]map[interface{}]matter.PageStubs

	sync.RWMutex
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.FeatureTag = tag
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.indexPageKeys = DefaultPageIndexKeys
	f.translationsBucket = make(map[language.Tag]kvs.KeyValueStore)
}

func (f *CFeature) SetKVC(tag feature.Tag, name string) MakeFeature {
	f.kvcTag = tag
	f.kvcName = name
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	if f.kvcTag == "" || f.kvcName == "" {
		err = fmt.Errorf("%v feature requires a KeyValueCache, use f.SetKVC(tag,name) during enjin build process", f.Tag())
		return
	}
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}

	for _, ef := range f.Enjin.Features() {
		if kvcf, ok := ef.(kvs.KeyValueCaches); ok && ef.Tag() == f.kvcTag {
			var kvc kvs.KeyValueCache
			if kvc, err = kvcf.Get(f.kvcName); err != nil {
				err = fmt.Errorf("%v cache not found: %v", f.kvcTag, f.kvcName)
				return
			}
			f.cache = kvc
			break
		}
	}

	if f.cache == nil {
		err = fmt.Errorf("%v feature not found", f.kvcTag)
		return
	}

	if f.allUrlsBucket, err = f.cache.Bucket(gAllUrlsBucketName); err != nil {
		return
	}
	if f.pageUrlsBucket, err = f.cache.Bucket(gPageUrlsBucketName); err != nil {
		return
	}
	if f.pageStubsBucket, err = f.cache.Bucket(gPageStubsBucketName); err != nil {
		return
	}
	if f.translatedByBucket, err = f.cache.Bucket(gPageTranslatedByBucketName); err != nil {
		return
	}
	if f.redirectionsBucket, err = f.cache.Bucket(gPageRedirectionsBucketName); err != nil {
		return
	}
	if f.contextValueKeyedBucket, err = f.cache.Bucket(gPageContextValuesBucketName); err != nil {
		return
	}
	return
}

func (f *CFeature) PerformQuery(input string) (stubs []*matter.PageStub, err error) {
	f.RLock()
	defer f.RUnlock()
	stubs, err = matcher.NewProcess(input, f.Enjin)
	return
}

func (f *CFeature) PerformSelect(input string) (selected map[string]interface{}, err error) {
	f.RLock()
	defer f.RUnlock()
	t, _ := f.Enjin.GetTheme()
	selected, err = selector.NewProcessWith(input, t, f)
	return
}

func (f *CFeature) AddToIndex(stub *matter.PageStub, p *page.Page) (err error) {
	f.Lock()
	defer f.Unlock()

	// log.DebugF("adding to index: %v - %v", p.Url, p.Shasum)

	if err = f.processPageStub(p.Shasum, stub); err != nil {
		return
	}

	if err = f.processPageUrl(p.Url, p.Shasum); err != nil {
		return
	}

	if redirects := p.Redirections(); len(redirects) > 0 {
		if err = f.processRedirections(p.Shasum, redirects); err != nil {
			return
		}
	}

	if err = f.processTranslatedBy(p.Url, p.Shasum); err != nil {
		return
	}

	if err = f.processTranslations(p.LanguageTag, p.Shasum, p.Url); err != nil {
		return
	}
	if p.Translates != "" {
		if err = f.processTranslations(p.LanguageTag, p.Shasum, p.Translates); err != nil {
			return
		}
	}

	for pCtxKey, pCtxValue := range p.Context {

		kebab := strcase.ToKebab(pCtxKey)
		if beStrings.StringInSlices(kebab, AlwaysIgnorePageIndexKeys, f.ignorePageKeys) {
			continue
		} else if !beStrings.StringInSlices(kebab, f.indexPageKeys) {
			continue
		}

		switch t := pCtxValue.(type) {
		case string,
			float32, float64,
			int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64,
			time.Time, time.Duration:
			if err = f.processPageContextValue(pCtxKey, p.Shasum, t); err != nil {
				return
			}
		case []string:
			for _, tv := range t {
				if err = f.processPageContextValue(pCtxKey, p.Shasum, tv); err != nil {
					return
				}
			}
		case []interface{}:
			for _, tv := range t {
				if err = f.processPageContextValue(pCtxKey, p.Shasum, tv); err != nil {
					return
				}
			}
		}

	}
	return
}

func (f *CFeature) RemoveFromIndex(tag language.Tag, file, shasum string) {
	// TODO: remove page from pql index
	log.FatalDF(1, "method not implemented")
	return
}

func (f *CFeature) YieldPageContextValues(key string) (values chan interface{}) {
	values = kvs.YieldFlatList[interface{}](f.contextValueKeyedBucket, key)
	return
}

func (f *CFeature) YieldPageContextValueStubs(key string) (pairs chan *pagecache.ValueStubPair) {
	pairs = make(chan *pagecache.ValueStubPair)
	go func() {
		defer close(pairs)
		ctxKeyBucketName := f.makeCtxValBucketName(key)
		ctxKeyBucket := f.cache.MustBucket(ctxKeyBucketName)

		for value := range f.YieldPageContextValues(key) {
			if shasums, err := kvs.GetSlice[string](ctxKeyBucket, value); err == nil && len(shasums) > 0 {
				for _, shasum := range shasums {
					if vStub, eeee := f.pageStubsBucket.Get(shasum); eeee == nil {
						if stub, ok := vStub.(*matter.PageStub); ok {
							pairs <- &pagecache.ValueStubPair{
								Value: value,
								Stub:  stub,
							}
						}
					}
				}
			}
		}

	}()
	return
}

func (f *CFeature) FindRedirection(url string) (p *page.Page) {
	f.RLock()
	defer f.RUnlock()

	theme, _ := f.Enjin.GetTheme()

	if shasum := kvs.GetValue[string](f.redirectionsBucket, url); shasum != "" {
		if vStub, e := f.pageStubsBucket.Get(shasum); e == nil {
			if stub, ok := vStub.(*matter.PageStub); ok {
				if p, e = page.NewFromPageStub(stub, theme); e != nil {
					log.ErrorF("error making redirected page from stub: %v - %v", url, e)
				}
			}
		}
	}

	return
}

func (f *CFeature) FindTranslations(url string) (pages []*page.Page) {
	f.RLock()
	defer f.RUnlock()

	if shasums, ee := kvs.GetSlice[string](f.translatedByBucket, url); ee == nil {
		for _, shasum := range shasums {
			if pg := f.findStubPage(shasum); pg != nil {
				pages = append(pages, pg)
			}
		}
	}

	return
}

func (f *CFeature) FindPage(tag language.Tag, url string) (pg *page.Page) {
	f.RLock()
	defer f.RUnlock()
	if tag == language.Und {
		if p, e := f.Lookup(f.Enjin.SiteDefaultLanguage(), url); e == nil {
			pg = p
			return
		}
	}
	if p, e := f.Lookup(tag, url); e == nil {
		pg = p
		return
	}
	return
}

func (f *CFeature) Lookup(tag language.Tag, path string) (pg *page.Page, err error) {
	f.RLock()
	defer f.RUnlock()
	if txBucket, ok := f.translationsBucket[tag]; ok {
		if shasums, e := kvs.GetSlice[string](txBucket, path); e == nil {
			for _, shasum := range shasums {
				if p := f.findStubPage(shasum); p != nil && p.LanguageTag == tag {
					pg = p
					return
				}
			}
		}
	}
	err = fmt.Errorf("page not found")
	return
}

func (f *CFeature) LookupPrefixed(prefix string) (pages []*page.Page) {
	f.RLock()
	defer f.RUnlock()
	if allUrls, e := kvs.GetSlice[string](f.allUrlsBucket, "all"); e == nil {
		for _, url := range allUrls {
			if strings.HasPrefix(url, prefix) {
				if shasums, ee := kvs.GetSlice[string](f.pageUrlsBucket, url); ee == nil {
					for _, shasum := range shasums {
						if pg := f.findStubPage(shasum); pg != nil {
							pages = append(pages, pg)
						}
					}
				}
			}
		}
	}
	return
}