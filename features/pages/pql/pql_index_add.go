//go:build page_pql || pages || all

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

package pql

import (
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/iancoleman/strcase"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/kvs"
	"github.com/go-enjin/be/pkg/values"
)

func (f *CFeature) AddToIndex(stub *feature.PageStub, p feature.Page) (err error) {

	if p.Context().Bool("NoPageIndexing", false) {
		return
	}

	//start := time.Now()
	//defer func() {
	//	if err == nil {
	//		log.DebugF("%v indexed page %v in %v", f.Tag(), p.Url(), time.Now().Sub(start).String())
	//	}
	//}()

	f.Lock()
	defer f.Unlock()
	//isDefaultLocale := p.LanguageTag() == f.Enjin.SiteDefaultLanguage()

	// TODO: figure out a slightly more unique constraint than .Shasum(); copies can have the same shasums!

	if v, ee := f.pageStubsBucket.Get(p.Shasum()); ee == nil && v != nil {
		return
	} else if err = f.addIndexForPageStub(p.Shasum(), stub); err != nil {
		return
	} else if err = f.addIndexForPageUrl(p.Url(), p.Shasum()); err != nil {
		return
	}

	if redirects := p.Redirections(); len(redirects) > 0 {
		if err = f.addIndexForRedirections(p.Shasum(), redirects); err != nil {
			return
		}
	}

	if err = f.addIndexForTranslatedBy(p.Url(), p.Shasum()); err != nil {
		return
	} else if err = f.addIndexForTranslations(p.LanguageTag(), p.Shasum(), p.Url()); err != nil {
		return
	}
	if p.Translates() != "" {
		if err = f.addIndexForTranslations(p.LanguageTag(), p.Shasum(), p.Translates()); err != nil {
			return
		}
	}

	if p.Permalink() != uuid.Nil {
		// long-form root permalink
		permalinkUrl := "/" + p.Permalink().String()
		if err = f.addIndexForPageUrl(permalinkUrl, p.Shasum()); err != nil {
			return
		} else if err = f.addIndexForTranslations(p.LanguageTag(), p.Shasum(), permalinkUrl); err != nil {
			return
		} else if err = f.addIndexForPermalink(p.Permalink().String(), p.Shasum()); err != nil {
			return
		}
		// short-form root permalink
		permalinkUrl = "/" + p.PermalinkSha()
		if err = f.addIndexForPageUrl(permalinkUrl, p.Shasum()); err != nil {
			return
		} else if err = f.addIndexForTranslations(p.LanguageTag(), p.Shasum(), permalinkUrl); err != nil {
			return
		} else if err = f.addIndexForPermalink(p.PermalinkSha(), p.Shasum()); err != nil {
			return
		}
	}

	excluded := make(map[string]struct{})
	for _, key := range append(MustExcludeContextKeys(), append(AlwaysExcludeContextKeys, f.excludeContextKeys...)...) {
		excluded[key] = struct{}{}
	}
	included := make(map[string]struct{})
	for _, key := range append(AlwaysIncludeContextKeys, f.includeContextKeys...) {
		included[key] = struct{}{}
	}

	for pCtxKey, pCtxValue := range p.Context() {

		kebab := strcase.ToKebab(pCtxKey)

		if _, exclude := excluded[kebab]; exclude {
			continue
		} else if _, include := included[kebab]; !include {
			continue
		} else if values.IsEmpty(pCtxValue) {
			continue
		}

		switch t := pCtxValue.(type) {
		case string,
			float32, float64,
			int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64,
			time.Time, time.Duration:
			if err = f.addIndexForPageContextValue(pCtxKey, p.Shasum(), t); err != nil {
				return
			}
			//case []string:
			//	for _, tv := range t {
			//		if err = f.addIndexForPageContextValue(pCtxKey, p.Shasum(), tv); err != nil {
			//			return
			//		}
			//	}
			//case []interface{}:
			//	for _, tv := range t {
			//		if err = f.addIndexForPageContextValue(pCtxKey, p.Shasum(), tv); err != nil {
			//			return
			//		}
			//	}
		}

	}
	return
}

func (f *CFeature) addIndexForPageStub(shasum string, stub *feature.PageStub) (err error) {
	if err = f.pageStubsBucket.Set(shasum, stub); err != nil {
		err = fmt.Errorf("error setting shasum stubs bucket: %v", err)
		return
	}
	return
}

func (f *CFeature) addIndexForPermalink(id string, shasum string) (err error) {
	if err = kvs.AppendToFlatList[string](f.permalinksBucket, id, shasum); err != nil {
		err = fmt.Errorf("error appending page to permalinks list: %v - %v", id, err)
		return
	}
	return
}

func (f *CFeature) addIndexForPageUrl(url, shasum string) (err error) {

	if err = kvs.AppendToFlatList[string](f.pageUrlsBucket, url, shasum); err != nil {
		err = fmt.Errorf("error appending page to page urls list: %v - %v", url, err)
		return
	}

	if err = kvs.AppendToFlatList(f.allUrlsBucket, "all", url); err != nil {
		err = fmt.Errorf("error updating flat-list of known urls: %v", err)
	}
	return
}

func (f *CFeature) addIndexForRedirections(shasum string, redirections []string) (err error) {
	for _, redirect := range redirections {
		if err = kvs.AppendToFlatList[string](f.redirectionsBucket, redirect, shasum); err != nil {
			err = fmt.Errorf("error appending redirect to cached urls list: %v - %v", redirect, err)
			return
		}
	}
	return
}

func (f *CFeature) addIndexForTranslatedBy(url, shasum string) (err error) {
	err = kvs.AppendToFlatList[string](f.translatedByBucket, url, shasum)
	return
}

func (f *CFeature) addIndexForTranslations(lang language.Tag, shasum, translates string) (err error) {
	if _, found := f.translationsBucket[lang]; !found {
		if f.translationsBucket[lang], err = f.KVC().Bucket(f.makeLangBucketName(lang)); err != nil {
			return
		}
	}

	if err = f.translationsBucket[lang].Set(translates, shasum); err != nil {
		err = fmt.Errorf("error setting languages string: [%v] %v - %v", lang, translates, err)
		return
	}
	return
}

func (f *CFeature) addIndexForPageContextValue(key, shasum string, value interface{}) (err error) {

	var valueKey string
	if valueKey, err = kvs.EncodeKeyValue(value); err != nil {
		err = fmt.Errorf("error encoding %v key value: %T - %v", key, value, err)
		return
	}

	// get the specific bucket for the given key
	ctxKeyedValueBucketName := f.makeCtxValBucketName(key)
	ctxKeyedValueBucket := f.KVC().MustBucket(ctxKeyedValueBucketName)

	// check if this value has been added to the list before
	if kvs.FlatListEmpty(ctxKeyedValueBucket, valueKey) == true {
		// value keyed flat list is distinct values only
		if err = kvs.AppendToFlatList[string](f.contextValueKeyedBucket, key, valueKey); err != nil {
			err = fmt.Errorf("error updating flat list: %v - %v", key, err)
			return
		}
	}

	if e := kvs.AppendToFlatList[string](ctxKeyedValueBucket, valueKey, shasum); e != nil {
		err = fmt.Errorf("error appending to bucket flat list: [%v]=\"%v\" - %v", ctxKeyedValueBucketName, value, err)
	}
	return
}
