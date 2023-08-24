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

package defaults

import (
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/features/outputs/htmlify"
	"github.com/go-enjin/be/features/pages/formats"
	"github.com/go-enjin/be/features/pages/funcmaps"
	"github.com/go-enjin/be/features/pages/partials"
	"github.com/go-enjin/be/features/pages/permalink"
	"github.com/go-enjin/be/features/pages/query"
	"github.com/go-enjin/be/features/requests/headers/proxy"
	"github.com/go-enjin/be/features/srv/listeners/httpd"
	"github.com/go-enjin/be/features/srv/pages"
	"github.com/go-enjin/be/features/srv/theme/renderer"
	"github.com/go-enjin/be/pkg/feature"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "preset-defaults"

type Feature interface {
	feature.Feature
}

type MakeFeature interface {
	Make() Feature

	SetRenderer(r feature.ThemeRenderer) MakeFeature
	SetListener(l feature.ServiceListener) MakeFeature
}

type CFeature struct {
	feature.CFeature

	fRenderer feature.ThemeRenderer
	fListener feature.ServiceListener
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

func (f *CFeature) SetRenderer(r feature.ThemeRenderer) MakeFeature {
	f.fRenderer = r
	return f
}

func (f *CFeature) SetListener(l feature.ServiceListener) MakeFeature {
	f.fListener = l
	return f
}

func (f *CFeature) Make() (feat Feature) {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	if err = f.CFeature.Build(b); err != nil {
		return
	}

	add := func(feat feature.Feature) (err error) {
		if err = feat.Build(b); err != nil {
			return
		}
		b.AddFeature(feat)
		return
	}

	for _, feat := range []feature.Feature{
		formats.New().Defaults().Make(),
		funcmaps.New().Defaults().Make(),
		partials.New().Make(),
		permalink.New().Make(),
		query.New().Make(),
		pages.New().Make(),
		htmlify.New().Make(),
		proxy.New().Enable().Make(),
	} {
		if err = add(feat); err != nil {
			return
		}
	}

	if f.fRenderer != nil {
		if err = add(f.fRenderer); err != nil {
			return
		}
	} else if err = add(renderer.New().Make()); err != nil {
		return
	}

	if f.fListener != nil {
		if err = add(f.fListener); err != nil {
			return
		}
	} else if err = add(httpd.New().Make()); err != nil {
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