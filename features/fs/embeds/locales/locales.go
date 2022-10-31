//go:build embed_locales || embeds || all

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

package locales

import (
	"embed"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
)

var _embedLocales *Feature

var _ feature.Feature = (*Feature)(nil)

const Tag feature.Tag = "EmbedLocales"

type Feature struct {
	feature.CFeature

	setup map[string]embed.FS
}

type MakeFeature interface {
	feature.MakeFeature

	Include(path string, efs embed.FS) MakeFeature
}

func New() MakeFeature {
	if _embedLocales == nil {
		_embedLocales = new(Feature)
		_embedLocales.Init(_embedLocales)
	}
	return _embedLocales
}

func (f *Feature) Include(path string, efs embed.FS) MakeFeature {
	f.setup[path] = efs
	return f
}

func (f *Feature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.setup = make(map[string]embed.FS)
}

func (f *Feature) Tag() (tag feature.Tag) {
	tag = Tag
	return
}

func (f *Feature) Build(_ feature.Buildable) (err error) {
	return
}

func (f *Feature) Setup(enjin feature.Internals) {
	tag := enjin.SiteDefaultLanguage()
	c := enjin.SiteLangCatalog()
	for path, efs := range f.setup {
		c.IncludeEmbed(tag, path, efs)
	}
}

func (f *Feature) Startup(ctx *cli.Context) (err error) {
	return
}