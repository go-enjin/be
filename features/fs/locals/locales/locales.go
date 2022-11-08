//go:build local_locales || locals || all

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
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

var _ feature.Feature = (*Feature)(nil)

const Tag feature.Tag = "LocalLocales"

type Feature struct {
	feature.CFeature

	setup []string
}

type MakeFeature interface {
	feature.MakeFeature

	Include(path string) MakeFeature
}

func New() MakeFeature {
	f := new(Feature)
	f.Init(f)
	return f
}

func (f *Feature) Include(path string) MakeFeature {
	if !beStrings.StringInStrings(path, f.setup...) {
		f.setup = append(f.setup, path)
	}
	return f
}

func (f *Feature) Init(this interface{}) {
	f.CFeature.Init(this)
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
	for _, path := range f.setup {
		c.IncludeLocal(tag, path)
	}
}

func (f *Feature) Startup(ctx *cli.Context) (err error) {
	return
}