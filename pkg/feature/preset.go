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

package feature

import (
	"github.com/go-corelibs/slices"
	"github.com/go-enjin/be/pkg/log"
)

var (
	_ Preset     = (*CPreset[MakePreset])(nil)
	_ MakePreset = (*CPreset[MakePreset])(nil)
)

type Preset interface {
	Label() (name string)
	Preset(b Builder) (err error)
}

type MakePreset interface {
	Make() Preset

	BaseMakePreset[MakePreset]
}

type BaseMakePreset[MakeTypedPreset interface{}] interface {
	// Include specifies additional features to be included during the build phase
	Include(features ...Feature) MakeTypedPreset

	// Prepend includes the specified features before all other features within the preset
	Prepend(features ...Feature) MakeTypedPreset

	// OmitTags specifies preset Feature.Tag()s to be omitted during the build phase
	OmitTags(features ...Tag) MakeTypedPreset

	// Overload provides a replacement for an existing feature (by tag)
	// Note: panics if tag specified is not found
	Overload(tag Tag, feature Feature) MakeTypedPreset
}

type CPreset[MakeTypedPreset interface{}] struct {
	Name     string
	Features Features

	this     interface{}
	tags     Tags
	omitTags Tags
	overload map[Tag]Feature
}

func NewPresetWith(name string, features ...Feature) MakePreset {
	p := &CPreset[MakePreset]{
		Name:     name,
		Features: features,
	}
	p.Init(p)
	return p
}

func (p *CPreset[MakeTypedPreset]) Init(this interface{}) {
	p.this = this
	for _, f := range p.Features {
		p.tags = append(p.tags, f.Tag())
	}
	return
}

func (p *CPreset[MakeTypedPreset]) Prepend(features ...Feature) MakeTypedPreset {
	p.Features = append(features, p.Features...)
	return p.this.(MakeTypedPreset)
}

func (p *CPreset[MakeTypedPreset]) Include(features ...Feature) MakeTypedPreset {
	p.Features = append(p.Features, features...)
	return p.this.(MakeTypedPreset)
}

func (p *CPreset[MakeTypedPreset]) Overload(tag Tag, feature Feature) MakeTypedPreset {
	if p.tags.Has(tag) {
		p.overload[tag] = feature
	} else {
		log.FatalDF(1, "%v preset tag not found: %v", p.Name, tag)
	}
	return p.this.(MakeTypedPreset)
}

func (p *CPreset[MakeTypedPreset]) OmitTags(features ...Tag) MakeTypedPreset {
	p.omitTags = append(p.omitTags, features...)
	return p.this.(MakeTypedPreset)
}

func (p *CPreset[MakeTypedPreset]) Make() Preset {
	return p.this.(Preset)
}

func (p *CPreset[MakeTypedPreset]) Label() (name string) {
	name = p.Name
	return
}

func (p *CPreset[MakeTypedPreset]) Preset(b Builder) (err error) {

	for idx := 0; idx < p.Features.Len(); idx++ {
		p.IncludeFeature(b, p.Features[idx])
	}

	return
}

// IncludeFeature will prepend the given feature to the Builder enjin, taking into account omissions and overloads
func (p *CPreset[MakeTypedPreset]) IncludeFeature(b Builder, f Feature) {
	if slices.Within(f.Tag(), p.omitTags) {
		return
	}
	if of, ok := p.overload[f.Tag()]; ok {
		b.PrependFeature(of)
	} else {
		b.PrependFeature(f)
	}
}
