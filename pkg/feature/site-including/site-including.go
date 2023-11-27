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

package site_including

import (
	"github.com/go-enjin/be/pkg/feature"
)

type MakeFeature[M interface{}] interface {
	Include(features ...feature.Feature) M
	Including(tags ...feature.Tag) M
}

type CSiteIncluding[T interface{}, M interface{}] struct {
	IncludeFeatures   feature.Features
	IncludingFeatures feature.Tags
	Features          feature.TypedFeatures[T]

	this interface{}
}

func New[T interface{}, M interface{}](this interface{}) (si *CSiteIncluding[T, M]) {
	si = &CSiteIncluding[T, M]{}
	si.InitSiteIncluding(this)
	return
}

func (si *CSiteIncluding[T, M]) InitSiteIncluding(this interface{}) {
	si.this = this
	return
}

func (si *CSiteIncluding[T, M]) Include(features ...feature.Feature) M {
	si.IncludeFeatures = append(si.IncludeFeatures, features...)
	t, _ := si.this.(M)
	return t
}

func (si *CSiteIncluding[T, M]) Including(tags ...feature.Tag) M {
	si.IncludingFeatures = append(si.IncludingFeatures, tags...)
	t, _ := si.this.(M)
	return t
}

func (si *CSiteIncluding[T, M]) BuildSiteIncluding(b feature.Buildable) {
	for _, ef := range si.IncludeFeatures {
		b.AddFeature(ef)
	}
	return
}

func (si *CSiteIncluding[T, M]) StartupSiteIncluding(enjin feature.Internals) {

	for _, v := range si.IncludeFeatures {
		if ef, ok := v.This().(T); ok {
			si.Features = append(si.Features, ef)
		}
	}

	for _, tag := range si.IncludingFeatures {
		if !si.Features.Has(tag) {
			if v, found := enjin.Features().Get(tag); found {
				if ef, ok := v.This().(T); ok {
					si.Features = append(si.Features, ef)
				}
			}
		}
	}

	return
}
