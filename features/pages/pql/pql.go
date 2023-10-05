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
	"os"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/features/pages/pql/matcher"
	"github.com/go-enjin/be/features/pages/pql/selector"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/kvs"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/pages"
	bePath "github.com/go-enjin/be/pkg/path"
	"github.com/go-enjin/be/pkg/slices"
	"github.com/go-enjin/be/types/page"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "pages-pql"

type Feature interface {
	feature.Feature
	feature.PageProvider
	feature.PageIndexFeature
	feature.QueryIndexFeature
	feature.PageContextProvider
}

type MakeFeature interface {
	Make() Feature

	// SetKeyValueCache Specifies the feature.KeyValueCaches tag and
	// feature.KeyValueCache name to use for storing runtime data
	SetKeyValueCache(tag feature.Tag, name string) MakeFeature

	// IncludeContextKeys appends to the list of page context keys to index,
	// by default the following keys are included:
	//   "Type", "Language", "Url", "Title", "Description"
	IncludeContextKeys(keys ...string) MakeFeature

	// SetIncludedContextKeys overwrites the list of context keys to index, use
	// this instead of IncludeContextKeys when needing to remove one or more of
	// the default list of inclusions
	SetIncludedContextKeys(keys ...string) MakeFeature
}

type CFeature struct {
	feature.CFeature

	kvcTag  feature.Tag
	kvcName string

	cache kvs.KeyValueCache

	excludeContextKeys []string
	includeContextKeys []string

	allUrlsBucket           kvs.KeyValueStore
	pageUrlsBucket          kvs.KeyValueStore
	pageStubsBucket         kvs.KeyValueStore
	permalinksBucket        kvs.KeyValueStore
	redirectionsBucket      kvs.KeyValueStore
	translatedByBucket      kvs.KeyValueStore
	translationsBucket      map[language.Tag]kvs.KeyValueStore
	contextValueKeyedBucket kvs.KeyValueStore
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
	f.includeContextKeys = BaseIncludeContextKeys()
	f.translationsBucket = make(map[language.Tag]kvs.KeyValueStore)
}

func (f *CFeature) SetKeyValueCache(tag feature.Tag, name string) MakeFeature {
	f.kvcTag = tag
	f.kvcName = name
	return f
}

func (f *CFeature) SetIncludedContextKeys(keys ...string) MakeFeature {
	f.includeContextKeys = nil
	f.IncludeContextKeys(keys...)
	return f
}

func (f *CFeature) IncludeContextKeys(keys ...string) MakeFeature {
	for _, key := range keys {
		kebab := strcase.ToKebab(key)
		if !slices.Within(kebab, f.includeContextKeys) {
			f.includeContextKeys = append(f.includeContextKeys, kebab)
		}
	}
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	if f.kvcTag == "" || f.kvcName == "" {
		err = fmt.Errorf("%v feature requires a KeyValueCache, use f.SetKeyValueCache(tag,name) during enjin build process", f.Tag())
		return
	}
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}

	if kvcf, ok := f.Enjin.Features().Get(f.kvcTag); !ok {
		err = fmt.Errorf("%v feature not found", f.kvcTag)
		return
	} else if kvcs, ok := kvcf.(kvs.KeyValueCaches); !ok {
		err = fmt.Errorf("%v feature is not a kvs.KeyValueCaches", f.kvcTag)
		return
	} else if kvc, ee := kvcs.Get(f.kvcName); ee != nil {
		err = fmt.Errorf("%v cache not found: %v", f.kvcTag, f.kvcName)
		return
	} else {
		f.cache = kvc
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
	if f.permalinksBucket, err = f.cache.Bucket(gPermalinksBucketName); err != nil {
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

func (f *CFeature) FindPageStub(shasum string) (stub *feature.PageStub) {
	//f.RLock()
	//defer f.RUnlock()
	if v, err := f.pageStubsBucket.Get(shasum); err == nil && v != nil {
		stub, _ = v.(*feature.PageStub)
	}
	return
}

func (f *CFeature) PerformQuery(input string) (stubs []*feature.PageStub, err error) {
	//f.RLock()
	//defer f.RUnlock()
	stubs, err = matcher.NewProcessWith(input, f.Enjin.MustGetTheme(), f)
	return
}

func (f *CFeature) PerformSelect(input string) (selected map[string]interface{}, err error) {
	//f.RLock()
	//defer f.RUnlock()
	selected, err = selector.NewProcessWith(input, f.Enjin.MustGetTheme(), f)
	return
}

func (f *CFeature) PageContextValuesCount(key string) (count uint64) {
	count = kvs.CountDistinctFlatListValues[string](f.contextValueKeyedBucket, key)
	return
}

func (f *CFeature) PageContextValueCounts(key string) (counts map[interface{}]uint64) {

	counts = make(map[interface{}]uint64)
	ctxKeyedValueBucketName := f.makeCtxValBucketName(key)
	ctxKeyedValueBucket := f.cache.MustBucket(ctxKeyedValueBucketName)

	for value := range kvs.YieldFlatList[string](f.contextValueKeyedBucket, key) {

		valueKey, _ := kvs.EncodeKeyValue(value)
		//list, _ := kvs.GetStringSlice(ctxKeyedValueBucket, valueKey)
		list := kvs.GetFlatList[string](ctxKeyedValueBucket, valueKey)
		counts[value] += uint64(len(list))

	}

	return
}

func (f *CFeature) YieldPageContextValues(key string) (values chan interface{}) {
	values = kvs.YieldFlatList[interface{}](f.contextValueKeyedBucket, key)
	return
}

func (f *CFeature) YieldPageContextValueStubs(key string) (pairs chan *feature.ValueStubPair) {
	pairs = make(chan *feature.ValueStubPair)
	go func() {
		defer close(pairs)

		ctxKeyedValueBucketName := f.makeCtxValBucketName(key)
		ctxKeyedValueBucket := f.cache.MustBucket(ctxKeyedValueBucketName)

		found := make(map[string]struct{})

		for value := range f.YieldPageContextValues(key) {

			var err error
			var valueKey string
			if valueKey, err = kvs.EncodeKeyValue(value); err != nil {
				log.ErrorF("error encoding %v key value: %T - %v", key, value, err)
				continue
			}

			//if shasums, ee := kvs.GetStringSlice(ctxKeyedValueBucket, valueKey); ee == nil && len(shasums) > 0 {
			if shasums := kvs.GetFlatList[string](ctxKeyedValueBucket, valueKey); len(shasums) > 0 {
				for _, shasum := range shasums {
					if _, seen := found[shasum]; !seen {
						if vStub, eeee := f.pageStubsBucket.Get(shasum); eeee == nil {
							if stub, ok := vStub.(*feature.PageStub); ok {
								pairs <- &feature.ValueStubPair{
									Value: value,
									Stub:  stub,
								}
								found[shasum] = struct{}{}
							}
						}
					}
				}
			}

		}

	}()
	return
}

func (f *CFeature) YieldFilterPageContextValueStubs(include bool, key string, value interface{}) (pairs chan *feature.ValueStubPair) {

	ctxKeyedValueBucketName := f.makeCtxValBucketName(key)
	ctxKeyedValueBucket := f.cache.MustBucket(ctxKeyedValueBucketName)
	var searchKey string
	if valueKey, err := kvs.EncodeKeyValue(value); err != nil {
		log.ErrorF("error encoding %v key value: %T - %v", key, value, err)
	} else {
		searchKey = valueKey
	}
	pairs = make(chan *feature.ValueStubPair)
	go func() {
		defer close(pairs)

		found := make(map[string]struct{})

		for yv := range kvs.YieldFlatList[interface{}](f.contextValueKeyedBucket, key) {

			var err error
			var valueKey string
			if valueKey, err = kvs.EncodeKeyValue(yv); err != nil {
				log.ErrorF("error encoding %v key value: %T - %v", key, yv, err)
				continue
			} else if include && searchKey != valueKey {
				continue
			} else if !include && searchKey == valueKey {
				continue
			}

			for shasum := range kvs.YieldFlatList[string](ctxKeyedValueBucket, valueKey) {
				if _, seen := found[shasum]; !seen {
					if vStub, eeee := f.pageStubsBucket.Get(shasum); eeee == nil {
						if stub, ok := vStub.(*feature.PageStub); ok {
							pairs <- &feature.ValueStubPair{
								Value: yv,
								Stub:  stub,
							}
							found[stub.Shasum] = struct{}{}
						}
					}
				}
			}

		}

	}()
	return
}

func (f *CFeature) FilterPageContextValueStubs(include bool, key string, value interface{}) (stubs feature.PageStubs) {

	ctxKeyedValueBucketName := f.makeCtxValBucketName(key)
	ctxKeyedValueBucket := f.cache.MustBucket(ctxKeyedValueBucketName)
	var searchKey string
	if valueKey, err := kvs.EncodeKeyValue(value); err != nil {
		log.ErrorF("error encoding %v key value: %T - %v", key, value, err)
		return
	} else {
		searchKey = valueKey
	}

	found := make(map[string]struct{})

	for yv := range kvs.YieldFlatList[interface{}](f.contextValueKeyedBucket, key) {

		var err error
		var valueKey string
		if valueKey, err = kvs.EncodeKeyValue(yv); err != nil {
			log.ErrorF("error encoding %v key value: %T - %v", key, yv, err)
			continue
		} else if include && searchKey != valueKey {
			continue
		} else if !include && searchKey == valueKey {
			continue
		}

		for shasum := range kvs.YieldFlatList[string](ctxKeyedValueBucket, valueKey) {
			if _, seen := found[shasum]; !seen {
				if vStub, eeee := f.pageStubsBucket.Get(shasum); eeee == nil {
					if stub, ok := vStub.(*feature.PageStub); ok {
						stubs = append(stubs, stub)
						found[stub.Shasum] = struct{}{}
					}
				}
			}
		}

	}

	return
}

func (f *CFeature) FindRedirection(url string) (p feature.Page) {
	//f.RLock()
	//defer f.RUnlock()

	theme, _ := f.Enjin.GetTheme()
	url = bePath.CleanWithSlash(url)

	if shasum := kvs.GetValue[string](f.redirectionsBucket, url); shasum != "" {
		if vStub, e := f.pageStubsBucket.Get(shasum); e == nil {
			if stub, ok := vStub.(*feature.PageStub); ok {
				if p, e = page.NewPageFromStub(stub, theme); e != nil {
					log.ErrorF("error making redirected page from stub: %v - %v", url, e)
				}
			}
		}
	}

	return
}

func (f *CFeature) FindTranslations(url string) (pages []feature.Page) {
	//f.RLock()
	//defer f.RUnlock()

	url = bePath.CleanWithSlash(url)

	if shasums := kvs.GetFlatList[string](f.translatedByBucket, url); len(shasums) > 0 {
		for _, shasum := range shasums {
			if pg := f.findStubPage(shasum); pg != nil {
				pages = append(pages, pg)
			}
		}
	}

	return
}

func (f *CFeature) FindPage(tag language.Tag, url string) (pg feature.Page) {
	//f.RLock()
	//defer f.RUnlock()

	url = bePath.CleanWithSlash(url)

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

func (f *CFeature) Lookup(tag language.Tag, path string) (pg feature.Page, err error) {
	//f.RLock()
	//defer f.RUnlock()

	path = bePath.CleanWithSlash(path)

	if id, ok := pages.ParsePermalink(path); ok {
		if v, ee := f.permalinksBucket.Get(id); ee == nil {
			shasum, _ := v.(string)
			if p := f.findStubPage(shasum); p != nil {
				pg = p
				return
			}
		}
	}

	process := func(tag language.Tag, path string) (pg feature.Page, err error) {
		if txBucket, ok := f.translationsBucket[tag]; ok {
			if v, e := txBucket.Get(path); e == nil {
				shasum, _ := v.(string)
				if p := f.findStubPage(shasum); p != nil && p.LanguageTag() == tag {
					pg = p
					return
				}
			}
		}
		err = os.ErrNotExist
		return
	}

	// check for the given tag
	if pg, err = process(tag, path); err == nil || err != os.ErrNotExist {
		return
	}

	if tag != language.Und {
		// path not found, check for (Und) fallback
		if pg, err = process(language.Und, path); err == nil || err != os.ErrNotExist {
			return
		}
	}

	err = os.ErrNotExist
	return
}

func (f *CFeature) LookupPrefixed(prefix string) (pages []feature.Page) {
	//f.RLock()
	//defer f.RUnlock()

	prefix = bePath.CleanWithSlash(prefix)

	allUrls := kvs.GetFlatList[string](f.allUrlsBucket, "all")
	for _, url := range allUrls {
		if strings.HasPrefix(url, prefix) {
			//if shasums, ee := kvs.GetSlice[string](f.pageUrlsBucket, url); ee == nil {
			if shasums := kvs.GetFlatList[string](f.pageUrlsBucket, url); len(shasums) > 0 {
				for _, shasum := range shasums {
					if pg := f.findStubPage(shasum); pg != nil {
						pages = append(pages, pg)
					}
				}
			}
		}
	}
	return
}