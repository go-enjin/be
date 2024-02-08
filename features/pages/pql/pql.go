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
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"

	"github.com/go-corelibs/x-text/language"

	clPath "github.com/go-corelibs/path"
	"github.com/go-corelibs/slices"
	"github.com/go-enjin/be/features/pages/pql/matcher"
	"github.com/go-enjin/be/features/pages/pql/selector"
	"github.com/go-enjin/be/pkg/feature"
	uses_kvc "github.com/go-enjin/be/pkg/feature/uses-kvc"
	"github.com/go-enjin/be/pkg/kvs"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/pages"
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

	UnsafeAllUrls() (store feature.KeyValueStore)
}

type MakeFeature interface {
	uses_kvc.MakeFeature[MakeFeature]

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

	Make() Feature
}

type CFeature struct {
	feature.CFeature
	uses_kvc.CUsesKVC[MakeFeature]

	excludeContextKeys []string
	includeContextKeys []string

	allUrlsBucket           feature.ExtendedKeyValueStore
	pageUrlsBucket          feature.KeyValueStore
	pageStubsBucket         feature.KeyValueStore
	permalinksBucket        feature.KeyValueStore
	redirectionsBucket      feature.KeyValueStore
	translatedByBucket      feature.KeyValueStore
	translationsBucket      map[language.Tag]feature.KeyValueStore
	contextValueKeyedBucket feature.KeyValueStore
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.CFeature.Construct(f)
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.CUsesKVC.InitUsesKVC(f)
	f.includeContextKeys = BaseIncludeContextKeys()
	f.translationsBucket = make(map[language.Tag]feature.KeyValueStore)
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
	if err = f.CFeature.Build(b); err != nil {
		return
	} else if err = f.CUsesKVC.BuildUsesKVC(); err != nil {
		return
	}
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	} else if err = f.CUsesKVC.StartupUsesKVC(f.Enjin.Features()); err != nil {
		return
	}

	if f.allUrlsBucket, err = kvs.ExtendedBucket(f.KVC(), gAllUrlsBucketName); err != nil {
		return
	} else if f.pageUrlsBucket, err = f.KVC().Bucket(gPageUrlsBucketName); err != nil {
		return
	} else if f.pageStubsBucket, err = f.KVC().Bucket(gPageStubsBucketName); err != nil {
		return
	} else if f.permalinksBucket, err = f.KVC().Bucket(gPermalinksBucketName); err != nil {
		return
	} else if f.translatedByBucket, err = f.KVC().Bucket(gPageTranslatedByBucketName); err != nil {
		return
	} else if f.redirectionsBucket, err = f.KVC().Bucket(gPageRedirectionsBucketName); err != nil {
		return
	} else if f.contextValueKeyedBucket, err = f.KVC().Bucket(gPageContextValuesBucketName); err != nil {
		return
	}

	// ensure the site locale buckets exist so that existing kvs data can be found
	for _, tag := range append(f.Enjin.SiteLocales(), language.Und) {
		if err = f.prepareTranslationsBucket(tag); err != nil {
			err = fmt.Errorf("error preparing translations bucket: %v - %w", tag, err)
			return
		}
	}
	return
}

func (f *CFeature) UnsafeAllUrls() (store feature.KeyValueStore) {
	store = f.allUrlsBucket
	return
}

func (f *CFeature) FindPageStub(shasum string) (stub *feature.PageStub) {
	//f.RLock()
	//defer f.RUnlock()
	s := &feature.PageStub{}
	if err := kvs.GetUnmarshal(f.pageStubsBucket, shasum, s); err == nil {
		stub = s
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
	ctxKeyedValueBucket := f.KVC().MustBucket(ctxKeyedValueBucketName)

	for value := range kvs.YieldFlatList[string](f.contextValueKeyedBucket, key) {

		valueKey, _ := kvs.MarshalConcrete(value)
		list := kvs.GetFlatList[string](ctxKeyedValueBucket, valueKey)
		counts[value] += uint64(len(list))

	}

	return
}

func (f *CFeature) YieldPageContextValueStubs(key string) (pairs chan *feature.ValueStubPair) {
	pairs = make(chan *feature.ValueStubPair)
	go func() {
		defer close(pairs)

		ctxKeyedValueBucketName := f.makeCtxValBucketName(key)
		ctxKeyedValueBucket := f.KVC().MustBucket(ctxKeyedValueBucketName)

		found := make(map[string]struct{})

		for valueKey := range kvs.YieldFlatList[string](f.contextValueKeyedBucket, key) {

			//var err error
			//var valueKey string
			//if valueKey, err = kvs.EncodeKeyValue(value); err != nil {
			//	log.ErrorF("error encoding %v key value: %T - %v", key, value, err)
			//	continue
			//}

			var err error
			var value interface{}
			if value, err = kvs.UnmarshalConcrete(valueKey); err != nil {
				log.ErrorF("error unmarshalling concrete value: %q - %v", valueKey, err)
				continue
			}

			if shasums := kvs.GetFlatList[string](ctxKeyedValueBucket, valueKey); len(shasums) > 0 {
				for _, shasum := range shasums {
					if _, seen := found[shasum]; !seen {
						if stub := f.getPageStub(f.pageStubsBucket, shasum); stub != nil {
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

	}()
	return
}

func (f *CFeature) YieldFilterPageContextValueStubs(include bool, key string, value interface{}) (pairs chan *feature.ValueStubPair) {

	ctxKeyedValueBucketName := f.makeCtxValBucketName(key)
	ctxKeyedValueBucket := f.KVC().MustBucket(ctxKeyedValueBucketName)
	var searchKey string
	if valueKey, err := kvs.MarshalConcrete(value); err != nil {
		log.ErrorF("error encoding %v key value: %T - %v", key, value, err)
	} else {
		searchKey = valueKey
	}
	pairs = make(chan *feature.ValueStubPair)
	go func() {
		defer close(pairs)

		found := make(map[string]struct{})

		for valueKey := range kvs.YieldFlatList[string](f.contextValueKeyedBucket, key) {

			if include && searchKey != valueKey {
				continue
			} else if !include && searchKey == valueKey {
				continue
			}

			var err error
			var yv interface{}
			if yv, err = kvs.UnmarshalConcrete(valueKey); err != nil {
				log.ErrorF("error unmarshalling concrete value: %q - %v", valueKey, err)
				continue
			}

			for shasum := range kvs.YieldFlatList[string](ctxKeyedValueBucket, valueKey) {
				if _, seen := found[shasum]; !seen {
					if stub := f.getPageStub(f.pageStubsBucket, shasum); stub != nil {
						pairs <- &feature.ValueStubPair{
							Value: yv,
							Stub:  stub,
						}
						found[stub.Shasum] = struct{}{}
					}
				}
			}

		}

	}()
	return
}

func (f *CFeature) FilterPageContextValueStubs(include bool, key string, value interface{}) (stubs feature.PageStubs) {

	ctxKeyedValueBucketName := f.makeCtxValBucketName(key)
	ctxKeyedValueBucket := f.KVC().MustBucket(ctxKeyedValueBucketName)
	var searchKey string
	if valueKey, err := kvs.MarshalConcrete(value); err != nil {
		log.ErrorF("error encoding %v key value: %T - %v", key, value, err)
		return
	} else {
		searchKey = valueKey
	}

	found := make(map[string]struct{})

	for valueKey := range kvs.YieldFlatList[string](f.contextValueKeyedBucket, key) {

		if include && searchKey != valueKey {
			continue
		} else if !include && searchKey == valueKey {
			continue
		}

		for shasum := range kvs.YieldFlatList[string](ctxKeyedValueBucket, valueKey) {
			if _, seen := found[shasum]; !seen {
				if stub := f.getPageStub(f.pageStubsBucket, shasum); stub != nil {
					stubs = append(stubs, stub)
					found[stub.Shasum] = struct{}{}
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
	url = clPath.CleanWithSlash(url)

	if shasum := kvs.GetValue[string](f.redirectionsBucket, url); shasum != "" {
		if stub := f.getPageStub(f.pageStubsBucket, shasum); stub != nil {
			var e error
			if p, e = page.NewPageFromStub(stub, theme); e != nil {
				log.ErrorF("error making redirected page from stub: %v - %v", url, e)
			}
		}
	}

	return
}

func (f *CFeature) FindTranslations(url string) (pages []feature.Page) {
	//f.RLock()
	//defer f.RUnlock()

	url = clPath.CleanWithSlash(url)

	if shasums := kvs.GetFlatList[string](f.translatedByBucket, url); len(shasums) > 0 {
		for _, shasum := range shasums {
			if pg := f.findStubPage(shasum); pg != nil {
				pages = append(pages, pg)
			}
		}
	}

	return
}

func (f *CFeature) FindTranslationUrls(url string) (pages map[language.Tag]string) {
	//f.RLock()
	//defer f.RUnlock()

	pages = make(map[language.Tag]string)

	for _, p := range f.FindTranslations(url) {
		pages[p.LanguageTag()] = p.Url()
	}

	return
}

func (f *CFeature) FindPage(r *http.Request, tag language.Tag, url string) (p feature.Page) {
	//f.RLock()
	//defer f.RUnlock()
	var err error
	url = clPath.CleanWithSlash(url)

	if tag == language.Und {
		if p, err = f.Lookup(f.Enjin.SiteDefaultLanguage(), url); err == nil {
			return
		}
	}

	if p, err = f.Lookup(tag, url); err == nil {
		return
	}
	return
}

func (f *CFeature) Lookup(tag language.Tag, path string) (pg feature.Page, err error) {
	//f.RLock()
	//defer f.RUnlock()

	path = clPath.CleanWithSlash(path)

	if id, ok := pages.ParsePermalink(path); ok {
		var shasum string
		if ee := kvs.GetUnmarshal(f.permalinksBucket, id, &shasum); ee == nil && shasum != "" {
			if p := f.findStubPage(shasum); p != nil {
				pg = p
				return
			}
		}
	}

	process := func(tag language.Tag, path string) (pg feature.Page, err error) {
		if txBucket, ok := f.translationsBucket[tag]; ok {
			var shasum string
			if ee := kvs.GetUnmarshal(txBucket, path, &shasum); ee == nil && shasum != "" {
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
	if pg, err = process(tag, path); err == nil || !errors.Is(err, os.ErrNotExist) {
		return
	}

	if tag != language.Und {
		// path not found, check for (Und) fallback
		if pg, err = process(language.Und, path); err == nil || !errors.Is(err, os.ErrNotExist) {
			return
		}
	}

	err = os.ErrNotExist
	return
}

func (f *CFeature) LookupPrefixed(prefix string) (pages []feature.Page) {
	//f.RLock()
	//defer f.RUnlock()

	prefix = clPath.CleanWithSlash(prefix)

	//allUrls := kvs.GetFlatList[string](f.allUrlsBucket, "all")
	//allUrls := f.allUrlsBucket.Keys("")
	for url := range f.allUrlsBucket.StreamKeys("", nil) {
		if strings.HasPrefix(url, prefix) {
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
