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

func (f *CFeature) RemoveFromIndex(stub *feature.PageStub, p feature.Page) (err error) {

	//start := time.Now()
	//defer func() {
	//	if err == nil {
	//		log.DebugF("%v indexed page %v in %v", f.Tag(), p.Url(), time.Now().Sub(start).String())
	//	}
	//}()

	f.Lock()
	defer f.Unlock()
	//isDefaultLocale := p.LanguageTag() == f.Enjin.SiteDefaultLanguage()

	if v, ee := f.pageStubsBucket.Get(p.Shasum()); ee != nil || v == nil {
		return
	} else if err = f.removeIndexForPageStub(p.Shasum()); err != nil {
		return
	} else if err = f.removeIndexForPageUrl(p.Url(), p.Shasum()); err != nil {
		return
	}

	if redirects := p.Redirections(); len(redirects) > 0 {
		if err = f.removeIndexForRedirections(p.Shasum(), redirects); err != nil {
			return
		}
	}

	if err = f.removeIndexForTranslatedBy(p.Url(), p.Shasum()); err != nil {
		return
	} else if err = f.removeIndexForTranslations(p.LanguageTag(), p.Shasum(), p.Url()); err != nil {
		return
	}
	if p.Translates() != "" {
		if err = f.removeIndexForTranslations(p.LanguageTag(), p.Shasum(), p.Translates()); err != nil {
			return
		}
	}

	if p.Permalink() != uuid.Nil {
		// long-form root permalink
		permalinkUrl := "/" + p.Permalink().String()
		if err = f.removeIndexForPageUrl(permalinkUrl, p.Shasum()); err != nil {
			return
		} else if err = f.removeIndexForTranslations(p.LanguageTag(), p.Shasum(), permalinkUrl); err != nil {
			return
		} else if err = f.removeIndexForPermalink(p.Permalink().String(), p.Shasum()); err != nil {
			return
		}
		// short-form root permalink
		permalinkUrl = "/" + p.PermalinkSha()
		if err = f.removeIndexForPageUrl(permalinkUrl, p.Shasum()); err != nil {
			return
		} else if err = f.removeIndexForTranslations(p.LanguageTag(), p.Shasum(), permalinkUrl); err != nil {
			return
		} else if err = f.removeIndexForPermalink(p.PermalinkSha(), p.Shasum()); err != nil {
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
			if err = f.removeIndexForPageContextValue(pCtxKey, p.Shasum(), t); err != nil {
				return
			}
			//case []string:
			//	for _, tv := range t {
			//		if err = f.removeIndexForPageContextValue(pCtxKey, p.Shasum(), tv); err != nil {
			//			return
			//		}
			//	}
			//case []interface{}:
			//	for _, tv := range t {
			//		if err = f.removeIndexForPageContextValue(pCtxKey, p.Shasum(), tv); err != nil {
			//			return
			//		}
			//	}
		}

	}
	return
}

func (f *CFeature) removeIndexForPageStub(shasum string) (err error) {
	if err = f.pageStubsBucket.Delete(shasum); err != nil {
		err = fmt.Errorf("error setting shasum stubs bucket: %v", err)
		return
	}
	return
}

func (f *CFeature) removeIndexForPermalink(id string, shasum string) (err error) {
	if err = kvs.RemoveFromFlatList[string](f.permalinksBucket, id, shasum); err != nil {
		err = fmt.Errorf("error appending page to permalinks list: %v - %v", id, err)
		return
	}
	return
}

func (f *CFeature) removeIndexForPageUrl(url, shasum string) (err error) {

	if err = kvs.RemoveFromFlatList[string](f.pageUrlsBucket, url, shasum); err != nil {
		err = fmt.Errorf("error appending page to page urls list: %v - %v", url, err)
		return
	}

	if err = kvs.RemoveFromFlatList[string](f.allUrlsBucket, "all", url); err != nil {
		err = fmt.Errorf("error updating flat-list of known urls: %v", err)
	}

	return
}

func (f *CFeature) removeIndexForRedirections(shasum string, redirections []string) (err error) {
	for _, redirect := range redirections {
		if err = kvs.RemoveFromFlatList[string](f.redirectionsBucket, redirect, shasum); err != nil {
			err = fmt.Errorf("error appending redirect to cached urls list: %v - %v", redirect, err)
			return
		}
	}
	return
}

func (f *CFeature) removeIndexForTranslatedBy(url, shasum string) (err error) {
	err = kvs.RemoveFromFlatList[string](f.translatedByBucket, url, shasum)
	return
}

func (f *CFeature) removeIndexForTranslations(lang language.Tag, shasum, translates string) (err error) {
	if _, found := f.translationsBucket[lang]; !found {
		return
	}

	if err = f.translationsBucket[lang].Delete(translates); err != nil {
		err = fmt.Errorf("error setting languages string: [%v] %v - %v", lang, translates, err)
		return
	}
	return
}

func (f *CFeature) removeIndexForPageContextValue(key, shasum string, value interface{}) (err error) {

	var valueKey string
	if valueKey, err = kvs.EncodeKeyValue(value); err != nil {
		err = fmt.Errorf("error encoding %v key value: %T - %v", key, value, err)
		return
	}

	// get the specific bucket for the given key
	ctxKeyedValueBucketName := f.makeCtxValBucketName(key)
	ctxKeyedValueBucket := f.KVC().MustBucket(ctxKeyedValueBucketName)

	// TODO: remove valueKey from f.contextValueKeyedBucket, must check if there are no shasums for the valueKey first,
	//       otherwise this drops valueKey for all other pages with the same key and value
	//if err = kvs.RemoveFromFlatList[string](f.contextValueKeyedBucket, key, valueKey); err != nil {
	//	err = fmt.Errorf("error removing from flat list: %v - %v", key, err)
	//	return
	//}

	if e := kvs.RemoveFromFlatList[string](ctxKeyedValueBucket, valueKey, shasum); e != nil {
		err = fmt.Errorf("error removing from bucket flat list: [%v]=\"%v\" - %v", ctxKeyedValueBucketName, value, err)
	}

	return
}
