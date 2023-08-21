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

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/kvs"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/types/page"
)

func (f *CFeature) processPageStub(shasum string, stub *feature.PageStub) (err error) {
	if err = f.pageStubsBucket.Set(shasum, stub); err != nil {
		err = fmt.Errorf("error setting shasum stubs bucket: %v", err)
		return
	}
	return
}

func (f *CFeature) processPageUrl(url, shasum string) (err error) {
	//if err = kvs.AppendToSlice[string](f.pageUrlsBucket, url, shasum); err != nil {
	//	err = fmt.Errorf("error appending page to page urls list: %v - %v", url, err)
	//	return
	//}
	if err = kvs.AppendToFlatList[string](f.pageUrlsBucket, url, shasum); err != nil {
		err = fmt.Errorf("error appending page to page urls list: %v - %v", url, err)
		return
	}

	// TODO: figure out a better "all urls" bucket pattern

	//if err = kvs.AppendToSlice[string](f.allUrlsBucket, "all", url); err != nil {
	//	err = fmt.Errorf("error appending page to all urls list: %v - %v", url, err)
	//	return
	//}

	if err = kvs.AppendToFlatList(f.allUrlsBucket, "url", url); err != nil {
		err = fmt.Errorf("error updating flat-list of known urls: %v", err)
	}
	return
}

func (f *CFeature) processRedirections(shasum string, redirections []string) (err error) {
	for _, redirect := range redirections {
		//if err = kvs.AppendToSlice[string](f.redirectionsBucket, redirect, shasum); err != nil {
		//	err = fmt.Errorf("error appending redirect to cached urls list: %v - %v", redirect, err)
		//	return
		//}
		if err = kvs.AppendToFlatList[string](f.redirectionsBucket, redirect, shasum); err != nil {
			err = fmt.Errorf("error appending redirect to cached urls list: %v - %v", redirect, err)
			return
		}
	}
	return
}

func (f *CFeature) processTranslatedBy(url, shasum string) (err error) {
	//err = kvs.AppendToSlice[string](f.translatedByBucket, url, shasum)
	err = kvs.AppendToFlatList[string](f.translatedByBucket, url, shasum)
	return
}

func (f *CFeature) processTranslations(lang language.Tag, shasum, translates string) (err error) {
	if _, found := f.translationsBucket[lang]; !found {
		if f.translationsBucket[lang], err = f.cache.Bucket(f.makeLangBucketName(lang)); err != nil {
			return
		}
	}
	//if err = kvs.AppendToSlice[string](f.translationsBucket[lang], translates, shasum); err != nil {
	//	err = fmt.Errorf("error appending to languages string list: [%v] %v - %v", lang, translates, err)
	//	return
	//}
	//if err = kvs.AppendToFlatList[string](f.translationsBucket[lang], translates, shasum); err != nil {
	//	err = fmt.Errorf("error appending to languages string list: [%v] %v - %v", lang, translates, err)
	//	return
	//}
	if err = f.translationsBucket[lang].Set(translates, shasum); err != nil {
		err = fmt.Errorf("error setting languages string: [%v] %v - %v", lang, translates, err)
		return
	}
	return
}

func (f *CFeature) processPageContextValue(key, shasum string, value interface{}) (err error) {

	var valueKey string
	if valueKey, err = kvs.EncodeKeyValue(value); err != nil {
		err = fmt.Errorf("error encoding %v key value: %T - %v", key, value, err)
		return
	}

	// get the specific bucket for the given key
	ctxKeyedValueBucketName := f.makeCtxValBucketName(key)
	ctxKeyedValueBucket := f.cache.MustBucket(ctxKeyedValueBucketName)

	// check if this value has been added to the list before
	if kvs.FlatListEmpty(ctxKeyedValueBucket, valueKey) == true {
		// value keyed flat list is distinct values only
		if err = kvs.AppendToFlatList[interface{}](f.contextValueKeyedBucket, key, value); err != nil {
			err = fmt.Errorf("error updating flat list: %v - %v", key, err)
			return
		}
	}

	if e := kvs.AppendToFlatList[string](ctxKeyedValueBucket, valueKey, shasum); e != nil {
		err = fmt.Errorf("error appending to bucket flat list: %v/%v[%v]=\"%v\" - %v", f.kvcTag, f.kvcName, ctxKeyedValueBucketName, value, err)
	}
	return
}

func (f *CFeature) findStubPage(shasum string) (pg feature.Page) {
	if stub := f.findStub(shasum); stub != nil {
		theme, _ := f.Enjin.GetTheme()
		if p, e := page.NewPageFromStub(stub, theme); e != nil {
			log.ErrorF("error making page from stub: %v - %v", stub.Source, e)
		} else {
			pg = p
		}
	}
	return
}

func (f *CFeature) findStub(shasum string) (stub *feature.PageStub) {
	if vStub, e := f.pageStubsBucket.Get(shasum); e == nil {
		if s, ok := vStub.(*feature.PageStub); ok {
			stub = s
		} else {
			log.ErrorF("expected: *matter.PageStub, received: %T from stubs bucket: %v", vStub, gPageStubsBucketName)
		}
	}
	return
}

func (f *CFeature) makeCtxValBucketName(key string) (name string) {
	name = f.Tag().String() + "__" + key
	return
}

func (f *CFeature) makeLangBucketName(lang language.Tag) (name string) {
	name = fmt.Sprintf("%s__%s", gPageTranslationsBucketName, lang)
	return
}