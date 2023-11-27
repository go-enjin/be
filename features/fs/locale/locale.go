//go:build !exclude_fs_locale

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

package locale

import (
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/golang-org-x-text/message/catalog"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/feature/filesystem"
	pkgLangCatalog "github.com/go-enjin/be/pkg/lang/catalog"
	"github.com/go-enjin/be/pkg/maps"
)

const Tag feature.Tag = "fs-locale"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type Feature interface {
	filesystem.Feature[MakeFeature]
	feature.LocalesProvider
}

type MakeFeature interface {
	filesystem.MakeFeature[MakeFeature]

	AddLanguageCatalog(ctlgs ...catalog.Catalog) MakeFeature

	Make() Feature
}

type CFeature struct {
	filesystem.CFeature[MakeFeature]

	catalogs []catalog.Catalog
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
	f.CFeature.Localized = true
}

func (f *CFeature) AddLanguageCatalog(ctlgs ...catalog.Catalog) MakeFeature {
	f.catalogs = append(f.catalogs, ctlgs...)
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	return
}

func (f *CFeature) Setup(enjin feature.Internals) {
	f.CFeature.Setup(enjin)
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	err = f.CFeature.Startup(ctx)
	return
}

func (f *CFeature) Shutdown() {
	return
}

func (f *CFeature) AddLocales(c pkgLangCatalog.Catalog) {
	tag := f.Enjin.SiteDefaultLanguage()
	c.AddCatalog(f.catalogs...)
	for _, point := range maps.SortedKeys(f.MountPoints) {
		for _, mp := range f.MountPoints[point] {
			c.AddLocalesFromFS(tag, mp.ROFS)
		}
	}
}
