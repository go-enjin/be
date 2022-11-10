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

	"github.com/go-enjin/be/features/pages/formats/html"
	"github.com/go-enjin/be/features/pages/formats/md"
	"github.com/go-enjin/be/features/pages/formats/njn"
	"github.com/go-enjin/be/features/pages/formats/org"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/types/theme-types"
)

var (
	_ Feature              = (*CFeature)(nil)
	_ MakeFeature          = (*CFeature)(nil)
	_ types.FormatProvider = (*CFeature)(nil)
)

type Feature interface {
	feature.Feature
}

type MakeFeature interface {
	Defaults() MakeFeature
	AddFormat(formats ...types.Format) MakeFeature

	Make() Feature
}

type CFeature struct {
	feature.CFeature

	formats map[string]types.Format

	enjin feature.Internals
}

func New() MakeFeature {
	f := new(CFeature)
	f.Init(f)
	return f
}

func (f *CFeature) Defaults() MakeFeature {
	f.AddFormat(
		md.New().Make(),
		org.New().Make(),
		njn.New().Defaults().Make(),
		html.New().Make(),
	)
	return f
}

func (f *CFeature) AddFormat(formats ...types.Format) MakeFeature {
	for _, format := range formats {
		for _, extn := range format.Extensions() {
			f.formats[extn] = format
		}
	}
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.formats = make(map[string]types.Format)
}

func (f *CFeature) Tag() (tag feature.Tag) {
	tag = "PageFormats"
	return
}

func (f *CFeature) Setup(enjin feature.Internals) {
	f.enjin = enjin
	for _, name := range maps.SortedKeys(f.formats) {
		format := f.formats[name]
		if this, ok := format.This().(feature.CanSetupInternals); ok {
			this.Setup(enjin)
		}
	}
}

func (f *CFeature) GetFormat(extn string) (format types.Format) {
	format, _ = f.formats[extn]
	return
}

func (f *CFeature) MatchFormat(filename string) (format types.Format, match string) {
	for extn, frmt := range f.formats {
		if strings.HasSuffix(filename, "."+extn) {
			match = extn
			format = frmt
		}
	}
	return
}