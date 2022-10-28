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
	"github.com/go-enjin/be/features/pages/formats/html"
	"github.com/go-enjin/be/features/pages/formats/md"
	"github.com/go-enjin/be/features/pages/formats/njn"
	"github.com/go-enjin/be/features/pages/formats/org"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/page"
	"github.com/go-enjin/be/pkg/types/theme-types"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

var _instance *CFeature

type Feature interface {
	feature.Feature
}

type MakeFeature interface {
	feature.MakeFeature

	Defaults() MakeFeature
	AddFormat(format types.Format) MakeFeature
}

type CFeature struct {
	feature.CFeature
}

func New() MakeFeature {
	if _instance == nil {
		_instance = new(CFeature)
		_instance.Init(_instance)
	}
	return _instance
}

func (f *CFeature) Defaults() MakeFeature {
	page.AddFormat(md.New().Make())
	page.AddFormat(org.New().Make())
	page.AddFormat(njn.New().Defaults().Make())
	page.AddFormat(html.New().Make())
	return f
}

func (f *CFeature) AddFormat(format types.Format) MakeFeature {
	page.AddFormat(format)
	return f
}

func (f *CFeature) RemoveFormat(name string) MakeFeature {
	page.RemoveFormat(name)
	return f
}

func (f *CFeature) Tag() (tag feature.Tag) {
	tag = "PageFormats"
	return
}