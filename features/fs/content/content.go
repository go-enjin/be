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
	"runtime"
	"runtime/debug"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"

	"github.com/go-corelibs/maps"
	"github.com/go-enjin/be/pkg/editor"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/feature/filesystem"
	uses_actions "github.com/go-enjin/be/pkg/feature/uses-actions"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/signals"
	"github.com/go-enjin/be/types/page"
)

var (
	DefaultGCPercent = -1
)

const Tag feature.Tag = "fs-content"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type Feature interface {
	filesystem.Feature[MakeFeature]
	feature.PageFileSystemFeature
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

	// SetStartupGC specifies the debug.SetGCPercent() value to use during startup indexing. The setting is restored
	// when the indexing is complete. The default is -1 which disables setting anything at all. The Go default GC
	// percent is 100. See: https://tip.golang.org/doc/gc-guide for more detail on what this does.
	SetStartupGC(percent int) MakeFeature

	SetStartupIndexing(enabled bool) MakeFeature

	SkipStartupIndexing(list ...feature.Tag) MakeFeature
}

type CFeature struct {
	filesystem.CFeature[MakeFeature]
	uses_actions.CUsesActions

	gcPercent int

	indexProviderTags    feature.Tags
	indexingProviderTags feature.Tags
	searchProviderTags   feature.Tags

	indexProviders    []feature.PageIndexFeature
	indexingProviders []feature.PageIndexingFeature
	searchProviders   []feature.SearchEnjinFeature

	startupIndexing     bool
	skipStartupIndexing feature.Tags

	cache feature.KeyValueCache
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
	f.CUsesActions.ConstructUsesActions(f)
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.gcPercent = DefaultGCPercent
	f.startupIndexing = true
}

func (f *CFeature) AddToIndexProviders(tag ...feature.Tag) MakeFeature {
	f.indexProviderTags = append(f.indexProviderTags, tag...)
	return f
}

func (f *CFeature) AddToSearchProviders(tag ...feature.Tag) MakeFeature {
	f.searchProviderTags = append(f.searchProviderTags, tag...)
	return f
}

func (f *CFeature) SetStartupGC(percent int) MakeFeature {
	f.gcPercent = percent
	return f
}

func (f *CFeature) SetStartupIndexing(enabled bool) MakeFeature {
	f.startupIndexing = enabled
	return f
}

func (f *CFeature) SkipStartupIndexing(list ...feature.Tag) MakeFeature {
	f.skipStartupIndexing = f.skipStartupIndexing.Append(list...)
	return f
}

func (f *CFeature) Make() Feature {
	f.indexProviderTags = f.indexProviderTags.Unique()
	f.searchProviderTags = f.searchProviderTags.Unique()
	return f
}

func (f *CFeature) makeFlagName() (category, indexing, skip string) {
	category = f.Tag().Kebab()
	indexing = strcase.ToKebab(category + "-startup-indexing")
	skip = strcase.ToKebab(category + "-skip-startup-indexing")
	return
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	category, indexingFlag, skipFlag := f.makeFlagName()
	b.AddFlags(
		&cli.BoolFlag{
			Name:     indexingFlag,
			Usage:    "toggle indexing content on startup",
			EnvVars:  b.MakeEnvKeys(indexingFlag),
			Value:    f.startupIndexing,
			Category: category,
		},
		&cli.StringSliceFlag{
			Name:     skipFlag,
			Usage:    "skip indexing features on startup",
			EnvVars:  b.MakeEnvKeys(skipFlag),
			Value:    cli.NewStringSlice(f.skipStartupIndexing.Strings()...),
			Category: category,
		},
	)
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}

	allFeatures := f.Enjin.Features().List()

	var indexingProviderTags feature.Tags
	for _, pif := range feature.FilterTyped[feature.PageIndexingFeature](allFeatures) {
		if tag := pif.Tag(); f.indexProviderTags.Has(tag) {
			f.indexProviderTags = f.indexProviderTags.Prune(tag)
			f.indexingProviders = append(f.indexingProviders, pif)
			indexingProviderTags = append(indexingProviderTags, tag)
			log.DebugF("%v feature found indexing provider: %v", f.Tag(), tag)
		} else {
			log.WarnF("%v feature ignoring indexing provider: %v", f.Tag(), tag)
		}
	}

	f.indexingProviderTags = indexingProviderTags // tags order matches providers order

	var indexProviderTags feature.Tags
	for _, pif := range feature.FilterTyped[feature.PageIndexFeature](allFeatures) {
		if tag := pif.Tag(); f.indexProviderTags.Has(tag) {
			f.indexProviders = append(f.indexProviders, pif)
			indexProviderTags = append(indexProviderTags, tag)
			log.DebugF("%v feature found index provider: %v", f.Tag(), tag)
		} else {
			log.WarnF("%v feature ignoring index provider: %v", f.Tag(), tag)
		}
	}

	f.indexProviderTags = indexProviderTags // tags order matches providers order

	var searchProviderTags feature.Tags
	for _, sef := range feature.FilterTyped[feature.SearchEnjinFeature](allFeatures) {
		tag := sef.Tag()
		if f.searchProviderTags.Has(tag) {
			f.searchProviders = append(f.searchProviders, sef)
			searchProviderTags = append(searchProviderTags, tag)
			log.DebugF("%v feature found search provider: %v", f.Tag(), tag)
		} else {
			log.WarnF("%v feature ignoring search provider: %v", f.Tag(), tag)
		}
	}

	f.searchProviderTags = searchProviderTags

	_, indexingFlag, skipFlag := f.makeFlagName()
	f.startupIndexing = ctx.Bool(indexingFlag)
	for _, name := range ctx.StringSlice(skipFlag) {
		tag := feature.Tag(name)
		if f.indexingProviderTags.Has(tag) || f.indexProviderTags.Has(tag) || f.searchProviderTags.Has(tag) {
			f.skipStartupIndexing = append(f.skipStartupIndexing, tag)
			continue
		}
		err = fmt.Errorf("%v skip indexing feature not found: %v", f.Tag(), name)
		return
	}
	var all bool
	for _, tag := range append(f.indexingProviderTags, append(f.indexProviderTags, f.searchProviderTags...)...) {
		if all = f.skipStartupIndexing.Has(tag); !all {
			break
		}
	}
	if all && f.startupIndexing {
		f.startupIndexing = false
		log.WarnF("skipping all startup indexing")
	}
	return
}

func (f *CFeature) PostStartup(ctx *cli.Context) (err error) {
	if f.startupIndexing {
		err = f.PopulateIndexes()
	}
	return
}

func (f *CFeature) UserActions() (list feature.Actions) {
	list = feature.Actions{
		f.Action("view", "page"),
	}
	return
}

func (f *CFeature) PopulateIndexes() (err error) {

	if f.indexingProviderTags.Len() == 0 && f.indexProviderTags.Len() == 0 && f.searchProviderTags.Len() == 0 {
		// early out with no work
		log.WarnF("%v feature has not been given any index or search providers", f.Tag())
		return
	}

	start := time.Now()
	var total int

	log.DebugF("%v feature adding pages to: %v", f.Tag(), append(f.indexingProviderTags, append(f.indexProviderTags, f.searchProviderTags...)...))

	var previousGOGC int
	if f.gcPercent != -1 {
		previousGOGC = debug.SetGCPercent(f.gcPercent) //< slow go runtime GC down a bit?
	}

	theme := f.Enjin.MustGetTheme()
	ectx := f.Enjin.Context(nil)
	for _, point := range maps.SortedKeyLengths(f.MountPoints) {
		for _, mp := range f.MountPoints[point] {
			if files, ee := mp.ROFS.ListAllFiles("."); ee == nil {

				fileCount := len(files)
				fileStartTime := time.Now()

				batchTotal := 0
				batchCount := 0
				batchStart := fileStartTime
				prevBatch := time.Duration(0)

				numBatches := 10
				if fileCount > 10000 {
					numBatches = fileCount / 5000
				} else if fileCount > 1000 {
					numBatches = fileCount / 500
				} else if fileCount > 100 {
					numBatches = 10
				} else {
					numBatches = 1
				}
				batchMax := fileCount / numBatches
				batchTrack := make(map[feature.Tag]time.Duration)

				for _, file := range files {

					if _, wf, ok := editor.ParseEditorWorkFile(file); ok {
						log.TraceF("%v feature ignoring editor work-file:  (%s) %v", f.Tag().String(), wf, file)
						continue
					}

					if pm, eee := f.ReadMountPageMatter(mp, file); eee != nil {

						log.ErrorF("error reading page matter: %v - %v", file, eee)

					} else if pmStub, ok := pm.Stub.(*feature.PageStub); ok && pmStub != nil {

						if pg, eeee := page.NewPageFromStub(pmStub, theme, ectx); eeee != nil {

							log.ErrorF("error making page from stub: %v - %v", file, eeee)

						} else {

							for _, pip := range f.indexingProviders {
								tag := pip.Tag()
								if f.skipStartupIndexing.Has(tag) {
									continue
								}
								pipStart := time.Now()
								var eeeee error
								var report feature.PageIndexReport
								if report, eeeee = pip.AddToIndexWithReport(pmStub, pg); eeeee != nil {
									log.ErrorF("error adding to page %q index: %v - %v", pip.Tag(), file, eeeee)
								} else {
									// log.DebugF("%v indexed %v", pip.(feature.Feature).Tag(), pg.Url)
								}
								batchTrack[tag] += time.Now().Sub(pipStart)
								for _, rtag := range maps.SortedKeys(report) {
									batchTrack[tag+":"+rtag] += report[rtag]
								}
							}

							for _, pip := range f.indexProviders {
								tag := pip.Tag()
								if f.skipStartupIndexing.Has(tag) {
									continue
								}
								pipStart := time.Now()
								if eeeee := pip.AddToIndex(pmStub, pg); eeeee != nil {
									log.ErrorF("error adding to page %q index: %v - %v", pip.Tag(), file, eeeee)
								} else {
									// log.DebugF("%v indexed %v", pip.(feature.Feature).Tag(), pg.Url)
								}
								batchTrack[tag] += time.Now().Sub(pipStart)
							}

							for _, sep := range f.searchProviders {
								tag := sep.Tag()
								if f.skipStartupIndexing.Has(tag) {
									continue
								}
								sepStart := time.Now()
								if eeeee := sep.AddToSearchIndex(pmStub, pg); eeeee != nil {
									log.ErrorF("error adding to search %q index: %v - %v", sep.Tag(), file, eeeee)
								} else {
									// log.DebugF("%v indexed %v", sep.(feature.Feature).Tag(), pg.Url)
								}
								batchTrack[tag] += time.Now().Sub(sepStart)
							}

							caiStart := time.Now()
							f.Enjin.Emit(signals.ContentAddIndexing, f.Tag().String(), file, theme, pmStub, pg)
							if caiTime := time.Now().Sub(caiStart); caiTime.Milliseconds() > 25 {
								batchTrack["signaling"] += caiTime
							}

							total += 1
							batchTotal += 1

							if batchTotal >= batchMax {
								batchCount += 1
								now := time.Now()
								dur := now.Sub(batchStart)
								trackSummary := make(map[feature.Tag]string)
								for k, v := range batchTrack {
									trackSummary[k] = v.String()
								}
								delta := dur - prevBatch

								totalTime := now.Sub(fileStartTime).Nanoseconds()
								avgPerFile := totalTime / int64(total)
								remainingFiles := int64(fileCount - total)
								eta := time.Duration(remainingFiles * avgPerFile)
								tSummary := fmt.Sprintf("%v", trackSummary)
								tSummary = tSummary[3:] // trim "map" prefix
								log.WarnF(
									"%v indexing eta: %v; batch %d/%d (%d) in %v +%v (%d/%d in %v) - %+v",
									f.Tag(),
									eta,
									batchCount, numBatches, batchTotal, dur, delta,
									total, fileCount, now.Sub(start),
									tSummary,
								)
								batchStart = now
								batchTotal = 0
								batchTrack = make(map[feature.Tag]time.Duration)
								prevBatch = dur
							}
						}
					}

				}
			}
		}
	}

	if f.gcPercent != -1 {
		debug.SetGCPercent(previousGOGC)
	}

	gcStart := time.Now()
	runtime.GC()

	end := time.Now()
	log.InfoF("%v feature indexed %d pages in %v (gc: %v)", f.Tag(), total, end.Sub(start), end.Sub(gcStart))

	return
}

func (f *CFeature) AddIndexing(filePath string) {

	if _, _, ok := editor.ParseEditorWorkFile(filePath); ok {
		return
	}

	theme := f.Enjin.MustGetTheme()
	ectx := f.Enjin.Context(nil)
	for _, point := range maps.SortedKeyLengths(f.MountPoints) {
		for _, mp := range f.MountPoints[point] {

			if pm, eee := f.ReadMountPageMatter(mp, filePath); eee == nil {

				if stub, ok := pm.Stub.(*feature.PageStub); ok && stub != nil {

					if p, eeee := page.NewPageFromStub(stub, theme, ectx); eeee == nil {

						for _, indexer := range f.searchProviders {
							if err := indexer.AddToSearchIndex(stub, p); err != nil {
								log.ErrorF("error adding search indexing: %v - %v", p.Url(), err)
							}
						}
						for _, indexer := range f.indexProviders {
							if err := indexer.AddToIndex(stub, p); err != nil {
								log.ErrorF("error adding page indexing: %v - %v", p.Url(), err)
							}
						}

						f.Enjin.Emit(signals.ContentAddIndexing, f.Tag().String(), filePath, theme, stub, p)

					}

					return
				}

			}

		}
	}

	return
}

func (f *CFeature) RemoveIndexing(filePath string) {

	if _, _, ok := editor.ParseEditorWorkFile(filePath); ok {
		return
	}

	theme := f.Enjin.MustGetTheme()
	ectx := f.Enjin.Context(nil)
	for _, point := range maps.SortedKeyLengths(f.MountPoints) {
		for _, mp := range f.MountPoints[point] {

			if pm, eee := f.ReadMountPageMatter(mp, filePath); eee == nil {

				if stub, ok := pm.Stub.(*feature.PageStub); ok && stub != nil {

					if p, eeee := page.NewPageFromStub(stub, theme, ectx); eeee == nil {

						for _, indexer := range f.searchProviders {
							indexer.RemoveFromSearchIndex(stub, p)
						}
						for _, indexer := range f.indexProviders {
							if err := indexer.RemoveFromIndex(stub, p); err != nil {
								log.ErrorF("error removing page indexing: %v - %v", p.Url(), err)
							}
						}

						f.Enjin.Emit(signals.ContentRemoveIndexing, f.Tag().String(), filePath, theme, stub, p)

					}

					return
				}

			}

		}
	}

	return
}
