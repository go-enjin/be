//go:build page_metrics || pages || all

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

package metrics

import (
	"net/http"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/maths"
	"github.com/go-enjin/be/pkg/slices"
	"github.com/go-enjin/be/pkg/strings/words"

	"github.com/go-enjin/golang-org-x-text/message"
)

var (
	AverageWordsPerMinute = 238.0
	SlowerWordsPerMinute  = 200.0
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "page-metrics"

type Feature interface {
	feature.Feature
	feature.PageContextModifier
}

type MakeFeature interface {
	Make() Feature

	AddPageType(types ...string) MakeFeature
	SetPageType(types ...string) MakeFeature

	AddArchetype(archetypes ...string) MakeFeature
	SetArchetype(archetypes ...string) MakeFeature
}

type CFeature struct {
	feature.CFeature

	pageTypes  []string
	archetypes []string
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
	return
}

func (f *CFeature) AddPageType(types ...string) MakeFeature {
	f.pageTypes = append(f.pageTypes, types...)
	return f
}

func (f *CFeature) SetPageType(types ...string) MakeFeature {
	f.pageTypes = types
	return f
}

func (f *CFeature) AddArchetype(archetypes ...string) MakeFeature {
	f.archetypes = append(f.archetypes, archetypes...)
	return f
}

func (f *CFeature) SetArchetype(archetypes ...string) MakeFeature {
	f.archetypes = archetypes
	return f
}

func (f *CFeature) Make() (feat Feature) {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	if err = f.CFeature.Build(b); err != nil {
		return
	}
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	return
}

func (f *CFeature) Shutdown() {
	f.CFeature.Shutdown()
}

func (f *CFeature) FilterPageContext(themeCtx, pageCtx context.Context, r *http.Request) (themeOut context.Context) {
	themeOut = themeCtx
	additions := f.UpdatePageContext(pageCtx, r)
	themeOut.Apply(additions)
	return
}

func (f *CFeature) UpdatePageContext(pageCtx context.Context, r *http.Request) (additions context.Context) {
	var printer *message.Printer
	printer = lang.GetPrinterFromRequest(r)

	additions = context.New()

	pageType := pageCtx.String("Type", "page")
	archetype := pageCtx.String("Archetype", "")

	if slices.Within(pageType, f.pageTypes) {
	} else if archetype != "" && slices.Within(archetype, f.archetypes) {
	} else {
		return
	}

	content := pageCtx.String("Content", "nil")

	wordCount := words.Count(content, nil)
	additions.SetSpecific("WordCount", wordCount)

	wordCountLabel := printer.Sprintf("%[1]d words", wordCount)
	additions.SetSpecific("CountedWords", wordCountLabel)

	fastTime := float64(wordCount) / AverageWordsPerMinute
	fastMinutes := maths.RoundDown(fastTime)
	slowMinutes := maths.RoundUp(float64(wordCount) / SlowerWordsPerMinute)

	additions.SetSpecific("ReadingTime", time.Duration(fastTime*float64(time.Minute)))

	minutes := printer.Sprintf("%[1]d-%[2]d minutes", fastMinutes, slowMinutes)
	additions.SetSpecific("ReadingMinutes", minutes)
	return
}