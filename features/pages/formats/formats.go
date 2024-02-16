//go:build !exclude_pages_formats

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

package formats

import (
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/features/pages/formats/html"
	"github.com/go-enjin/be/features/pages/formats/json"
	"github.com/go-enjin/be/features/pages/formats/md"
	"github.com/go-enjin/be/features/pages/formats/njn"
	"github.com/go-enjin/be/features/pages/formats/org"
	"github.com/go-enjin/be/features/pages/formats/text"
	"github.com/go-enjin/be/features/pages/formats/tmpl"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
)

const Tag feature.Tag = "pages-formats"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type Feature interface {
	feature.Feature
	feature.PageFormatProvider
}

type MakeFeature interface {
	Defaults() MakeFeature
	AddFormat(formats ...feature.PageFormat) MakeFeature

	Make() Feature
}

type CFeature struct {
	feature.CFeature

	formats map[string]feature.PageFormat
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
	f.formats = make(map[string]feature.PageFormat)
}

func (f *CFeature) Defaults() MakeFeature {
	f.AddFormat(
		md.New().Make(),
		org.New().Make(),
		njn.New().Defaults().Make(),
		html.New().Make(),
		json.New().Make(),
		text.New().Make(),
		tmpl.New().Make(),
	)
	return f
}

func (f *CFeature) AddFormat(formats ...feature.PageFormat) MakeFeature {
	for _, format := range formats {
		log.DebugF("adding format: %v", format.Label())
		for _, extn := range format.Extensions() {
			f.formats[extn] = format
		}
	}
	return f
}

func (f *CFeature) Make() Feature {
	if len(f.formats) == 0 {
		f.AddFormat(tmpl.New().Make())
	}
	return f
}

func (f *CFeature) Setup(enjin feature.Internals) {
	f.CFeature.Setup(enjin)
	for _, name := range maps.SortedKeys(f.formats) {
		format := f.formats[name]
		if this, ok := format.This().(feature.CanSetupInternals); ok {
			this.Setup(enjin)
		}
	}
}

func (f *CFeature) PostStartup(ctx *cli.Context) (err error) {
	for _, key := range maps.SortedKeys(f.formats) {
		format := f.formats[key]
		if psf, ok := format.This().(feature.PostStartupFeature); ok {
			if err = psf.PostStartup(ctx); err != nil {
				return
			}
		}
	}
	return
}

func (f *CFeature) ListFormats() (names []string) {
	names = maps.SortedKeyLengths(f.formats)
	return
}

func (f *CFeature) GetFormat(extn string) (format feature.PageFormat) {
	format, _ = f.formats[extn]
	return
}

func (f *CFeature) MatchFormat(filename string) (format feature.PageFormat, match string) {
	for _, extn := range f.ListFormats() {
		if filename == extn || strings.HasSuffix(filename, "."+extn) {
			match = extn
			format = f.formats[extn]
			return
		}
	}
	return
}
