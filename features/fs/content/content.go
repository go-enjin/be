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
	"time"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/feature/filesystem"
	"github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/indexing"
	"github.com/go-enjin/be/pkg/kvs"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/page"
	"github.com/go-enjin/be/pkg/userbase"
)

const Tag feature.Tag = "fs-content"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type Feature interface {
	filesystem.Feature[MakeFeature]
}

type MakeFeature interface {
	filesystem.MakeFeature[MakeFeature]

	Make() Feature

	// AddToIndexProviders indexes the content using the
	// feature.PageIndexFeature tags specified
	AddToIndexProviders(tag ...feature.Tag) MakeFeature

	// AddToSearchProviders indexes the content using the
	// feature.SearchEnjinFeature tags specified
	AddToSearchProviders(tag ...feature.Tag) MakeFeature
}

type CFeature struct {
	filesystem.CFeature[MakeFeature]

	indexProviderTags  feature.Tags
	searchProviderTags feature.Tags

	indexProviders  []indexing.PageIndexFeature
	searchProviders []indexing.SearchEnjinFeature

	cache kvs.KeyValueCache
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
	if f.indexProviderTags.Len() == 0 {
		log.FatalDF(1, "%v feature requires at least one feature.PageIndexFeature", f.Tag())
	}
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}

	allFeatures := f.Enjin.Features()

	for _, pif := range feature.FindAllTypedFeatures[indexing.PageIndexFeature](allFeatures) {
		if f.indexProviderTags.Has(pif.(feature.Feature).Tag()) {
			f.indexProviders = append(f.indexProviders, pif)
		}
	}
	if len(f.indexProviders) == 0 {
		err = fmt.Errorf(`%v feature required index provider(s) not found: %+v`, f.Tag(), f.indexProviderTags)
		return
	}

	for _, sef := range feature.FindAllTypedFeatures[indexing.SearchEnjinFeature](allFeatures) {
		if f.searchProviderTags.Has(sef.(feature.Feature).Tag()) {
			f.searchProviders = append(f.searchProviders, sef)
		}
	}

	err = f.PopulateIndexes()
	return
}

func (f *CFeature) Shutdown() {
	return
}

func (f *CFeature) UserActions() (list userbase.Actions) {

	tag := f.Tag().Kebab()
	list = userbase.Actions{
		userbase.NewAction(tag, "view", "page"),
	}

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

	log.DebugF("%v feature adding pages to: %v", f.Tag(), append(f.indexProviderTags, f.searchProviderTags...))

	theme := f.Enjin.MustGetTheme()
	for _, point := range maps.SortedKeys(f.MountPoints) {
		for _, mp := range f.MountPoints[point] {
			if files, ee := mp.ROFS.ListAllFiles("."); ee == nil {

				for _, file := range files {

					if pm, eee := f.ReadPageMatter(file); eee != nil {

						log.ErrorF("error reading page matter: %v - %v", file, eee)

					} else if pmStub, ok := pm.Stub.(*fs.PageStub); ok && pmStub != nil {

						if pg, eeee := page.NewFromPageStub(pmStub, theme); eeee != nil {

							log.ErrorF("error making page from stub: %v - %v", file, eeee)

						} else {

							for _, pip := range f.indexProviders {
								if eeeee := pip.AddToIndex(pmStub, pg); eeeee != nil {
									log.ErrorF("error adding to page index: %v - %v", file, eeeee)
								}
							}
							for _, sep := range f.searchProviders {
								if eeeee := sep.AddToSearchIndex(pmStub, pg); eeeee != nil {
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