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

func (f *CFeature) getPageStub(store feature.KeyValueStore, shasum string) (stub *feature.PageStub) {
	stub = &feature.PageStub{}
	if err := kvs.GetUnmarshal(store, shasum, stub); err != nil {
		stub = nil
	}
	return
}

func (f *CFeature) findStub(shasum string) (stub *feature.PageStub) {
	stub = f.getPageStub(f.pageStubsBucket, shasum)
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
