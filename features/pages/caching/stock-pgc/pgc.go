//go:build stock_pgc || pages || all

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

package pgc

import (
	"fmt"
	"sync"

	"github.com/go-enjin/golang-org-x-text/language"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/page"
	"github.com/go-enjin/be/pkg/pagecache"
	types "github.com/go-enjin/be/pkg/types/theme-types"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "PagesCaching"

type Feature interface {
	feature.Feature
	pagecache.CacheEnjinFeature
}

type CFeature struct {
	feature.CFeature

	cli   *cli.Context
	enjin feature.Internals

	formats  types.FormatProvider
	langMode lang.Mode
	fallback language.Tag
	search   pagecache.SearchEnjinFeature

	caches map[string]*Cache

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
	f.caches = make(map[string]*Cache)
}

func (f *CFeature) Tag() (tag feature.Tag) {
	tag = Tag
	return
}

func (f *CFeature) Setup(enjin feature.Internals) {
	f.enjin = enjin
	if t, err := f.enjin.GetTheme(); err != nil {
		log.FatalF("error getting enjin theme: %v", err)
	} else {
		f.formats = t
	}
	f.langMode = f.enjin.SiteLanguageMode()
	f.fallback = f.enjin.SiteDefaultLanguage()
	var ok bool
	for _, feat := range f.enjin.Features() {
		if f.search, ok = feat.(pagecache.SearchEnjinFeature); ok {
			break
		}
	}
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	f.cli = ctx
	return
}

func (f *CFeature) NewCache(bucket string) (err error) {
	if _, exists := f.caches[bucket]; exists {
		err = fmt.Errorf("cache already exists: %v", bucket)
		return
	}
	f.caches[bucket] = newCache(f.enjin, f.formats, f.langMode, f.fallback, f.search)
	return
}

func (f *CFeature) Mounted(bucket, path string) (ok bool) {
	if cache, exists := f.caches[bucket]; exists {
		ok = cache.Mounted(path)
	} else {
		log.FatalF("bucket not found: %v", bucket)
	}
	return
}

func (f *CFeature) Mount(bucket, mount, path string, mfs fs.FileSystem) {
	if cache, exists := f.caches[bucket]; exists {
		cache.Mount(mount, path, mfs)
	} else {
		log.FatalF("bucket not found: %v", bucket)
	}
	return
}

func (f *CFeature) Rebuild(bucket string) (ok bool) {
	if cache, exists := f.caches[bucket]; exists {
		if _, errs := cache.Rebuild(); len(errs) > 10 {
			log.ErrorF("errors (%d) during cache rebuilding: (too many to output)\n%v", len(errs), errs[0])
		} else if len(errs) > 0 {
			log.ErrorF("errors (%d) during cache rebuilding:\n%v", len(errs), errs)
		}
	} else {
		log.FatalF("bucket not found: %v", bucket)
	}
	return
}

func (f *CFeature) Lookup(bucket string, tag language.Tag, url string) (mount, path string, p *page.Page, err error) {
	if cache, exists := f.caches[bucket]; exists {
		mount, path, p, err = cache.Lookup(tag, url)
	} else {
		log.FatalF("bucket not found: %v", bucket)
	}
	return
}

func (f *CFeature) LookupTranslations(bucket, url string) (pgs []*page.Page) {
	if cache, exists := f.caches[bucket]; exists {
		pgs = cache.LookupTranslations(url)
	} else {
		log.FatalF("bucket not found: %v", bucket)
	}
	return
}

func (f *CFeature) LookupRedirect(bucket, url string) (p *page.Page, ok bool) {
	if cache, exists := f.caches[bucket]; exists {
		p, ok = cache.LookupRedirect(url)
	} else {
		log.FatalF("bucket not found: %v", bucket)
	}
	return
}

func (f *CFeature) LookupPrefix(bucket, prefix string) (found []*page.Page) {
	if cache, exists := f.caches[bucket]; exists {
		found = cache.LookupPrefix(prefix)
	} else {
		log.FatalF("bucket not found: %v", bucket)
	}
	return
}

func (f *CFeature) TotalCached(bucket string) (count uint64) {
	if cache, exists := f.caches[bucket]; exists {
		count = cache.TotalCached
	} else {
		log.FatalF("bucket not found: %v", bucket)
	}
	return
}