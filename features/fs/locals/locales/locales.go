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

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "LocalLocales"

type Feature interface {
	feature.Feature
}

type CFeature struct {
	feature.CFeature

	setup []string
}

type MakeFeature interface {
	Include(path string) MakeFeature

	Make() Feature
}

func New() MakeFeature {
	f := new(CFeature)
	f.Init(f)
	return f
}

func (f *CFeature) Include(path string) MakeFeature {
	if !beStrings.StringInStrings(path, f.setup...) {
		f.setup = append(f.setup, path)
	}
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
}

func (f *CFeature) Tag() (tag feature.Tag) {
	tag = Tag
	return
}

func (f *CFeature) Build(_ feature.Buildable) (err error) {
	return
}

func (f *CFeature) Setup(enjin feature.Internals) {
	tag := enjin.SiteDefaultLanguage()
	c := enjin.SiteLangCatalog()
	for _, path := range f.setup {
		c.IncludeLocal(tag, path)
	}
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	return
}