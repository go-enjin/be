//go:build fs_content || all

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

package content

import (
	"fmt"
	"sync"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/feature/mountable"
	"github.com/go-enjin/be/pkg/kvs"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/page"
	"github.com/go-enjin/be/pkg/pagecache"
)

const Tag feature.Tag = "fs-content"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type Feature interface {
	mountable.Feature[MakeFeature]
}

type MakeFeature interface {
	mountable.MakeFeature[MakeFeature]

	Make() Feature

	AddToIndexProviders(tag ...feature.Tag) MakeFeature
	AddToSearchProviders(tag ...feature.Tag) MakeFeature

	SetKeyValueCache(tag feature.Tag, name string) MakeFeature
}

type CFeature struct {
	mountable.CFeature[MakeFeature]

	kvcTag  feature.Tag
	kvcName string

	indexProviderTags  feature.Tags
	searchProviderTags feature.Tags

	indexProviders  []pagecache.PageIndexFeature
	searchProviders []pagecache.SearchEnjinFeature

	cache kvs.KeyValueCache

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
}

func (f *CFeature) SetKeyValueCache(tag feature.Tag, name string) MakeFeature {
	f.kvcTag = tag
	f.kvcName = name
	return f
}

func (f *CFeature) AddToIndexProviders(tag ...feature.Tag) MakeFeature {
	f.indexProviderTags = append(f.indexProviderTags, tag...)
	return f
}

func (f *CFeature) AddToSearchProviders(tag ...feature.Tag) MakeFeature {
	f.searchProviderTags = append(f.searchProviderTags, tag...)
	return f
}

func (f *CFeature) Make() Feature {
	f.indexProviderTags = f.indexProviderTags.Unique()
	f.searchProviderTags = f.searchProviderTags.Unique()
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}

	for _, ef := range f.Enjin.Features() {
		if kvcf, ok := ef.(kvs.KeyValueCaches); ok && ef.Tag() == f.kvcTag {
			if kvc, ee := kvcf.Get(f.kvcName); ee == nil {
				f.cache = kvc
				break
			}
		}
	}
	if f.cache == nil {
		err = fmt.Errorf("%v feature requires a feature.KeyValueCache to be set with enjin.SetKeyValueCache(tag,name)", f.Tag())
		return
	}

	for _, ef := range f.Enjin.Features() {
		if f.indexProviderTags.Has(ef.Tag()) {
			if pip, ok := ef.Self().(pagecache.PageIndexFeature); ok {
				f.indexProviders = append(f.indexProviders, pip)
			} else {
				err = fmt.Errorf("%v feature is not a pagecache.PageIndexFeature", ef.Tag())
				return
			}
		}
		if f.searchProviderTags.Has(ef.Tag()) {
			if sep, ok := ef.Self().(pagecache.SearchEnjinFeature); ok {
				f.searchProviders = append(f.searchProviders, sep)
			} else {
				err = fmt.Errorf("%v feature is not a pagecache.SearchEnjinFeature", ef.Tag())
				return
			}
		}
	}

	err = f.PopulateIndexes()
	return
}

func (f *CFeature) Shutdown() {
	return
}

func (f *CFeature) PopulateIndexes() (err error) {

	if f.indexProviderTags.Len() == 0 && f.searchProviderTags.Len() == 0 {
		// early out with no work
		log.WarnF("%v feature has not been given any index or search providers", f.Tag())
		return
	}

	start := time.Now()
	var total int

	theme, _ := f.Enjin.GetTheme()
	for _, point := range maps.SortedKeys(f.MountPoints) {
		for _, mp := range f.MountPoints[point] {
			if files, ee := mp.ROFS.ListAllFiles("."); ee == nil {

				// index all URLs from front matter
				// index all front-matter
				for _, file := range files {

					if pm, eee := f.ReadPageMatter(file); eee != nil {
						log.ErrorF("error reading page matter: %v - %v", file, eee)
					} else if pm.Stub != nil {
						if pg, eeee := page.NewFromPageStub(pm.Stub, theme); eeee != nil {
							log.ErrorF("error making page from stub: %v - %v", file, eeee)
						} else {
							for _, pip := range f.indexProviders {
								if eeeee := pip.AddToIndex(pm.Stub, pg); eeeee != nil {
									log.ErrorF("error adding to page index: %v - %v", file, eeeee)
								}
							}
							for _, sep := range f.searchProviders {
								if eeeee := sep.AddToSearchIndex(pm.Stub, pg); eeeee != nil {
									log.ErrorF("error adding to search index: %v - %v", file, eeeee)
								}
							}
							total += 1
							log.DebugF("%v feature indexed page: [%v] %v - %v", f.Tag(), pg.LanguageTag, pg.Url, file)
						}
					}

				}
			}
		}
	}

	end := time.Now()
	log.InfoF("%v feature indexed %d pages in %v", f.Tag(), total, end.Sub(start))

	return
}