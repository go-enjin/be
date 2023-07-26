//go:build page_context || pages || all

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

package context

import (
	"net/http"

	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "pages-context"

type Feature interface {
	feature.Feature
	feature.PageContextModifier
}

type MakeFeature interface {
	Make() Feature

	Set(key string, value interface{}) MakeFeature
	Flag(custom cli.Flag) MakeFeature
}

type CFeature struct {
	feature.CFeature

	ctx    context.Context
	custom []cli.Flag

	cli *cli.Context
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
	f.ctx = context.New()
	f.custom = make([]cli.Flag, 0)
}

func (f *CFeature) Set(key string, value interface{}) MakeFeature {
	f.ctx.Set(key, value)
	return f
}

func (f *CFeature) Flag(custom cli.Flag) MakeFeature {
	f.custom = append(f.custom, custom)
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	if len(f.custom) > 0 {
		b.AddFlags(f.custom...)
	}
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	f.cli = ctx
	return
}

func (f *CFeature) FilterPageContext(tCtx context.Context, pCtx context.Context, r *http.Request) (out context.Context) {
	out = tCtx.Copy()
	for _, custom := range f.custom {
		for _, name := range custom.Names() {
			out.Set(strcase.ToCamel(name), f.cli.Generic(name))
			log.DebugRF(r, "setting page context: %v => %v", strcase.ToCamel(name), f.cli.Generic(name))
		}
	}
	return
}