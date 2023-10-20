//go:build fs_theme || all

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

package themes

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/feature/filesystem"
	"github.com/go-enjin/be/pkg/lang/catalog"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/types/theme"
)

const Tag feature.Tag = "fs-theme"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type Feature interface {
	filesystem.Feature[MakeFeature]
	feature.LocalesProvider
}

type MakeFeature interface {
	// SetTheme is a convenience method for setting the current theme during the
	// enjin build phase
	SetTheme(name string) MakeFeature

	// Include themes loaded with a different instance of this themes feature
	Include(other Feature) MakeFeature

	ThemeEmbedSupport
	ThemeLocalSupport

	Make() Feature
}

type CFeature struct {
	filesystem.CFeature[MakeFeature]

	theme   string
	loading []*loadTheme
	loaded  []feature.Theme
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
	f.CFeature.Localized = false // themes are not localized filesystems!
}

func (f *CFeature) SetTheme(name string) MakeFeature {
	f.theme = name
	return f
}

func (f *CFeature) Include(other Feature) MakeFeature {
	if other != nil {
		if of, ok := other.(*CFeature); ok {
			for _, otherLoad := range of.loading {
				f.loading = append(f.loading, otherLoad)
				log.DebugDF(1, "%v including %v theme: %v - %v", f.Tag(), other.Tag(), otherLoad.support, otherLoad.path)
			}
		} else {
			log.FatalDF(1, "unsupported themes.Feature implementation: %T", other)
		}
	}
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	for _, ts := range f.loading {
		var t feature.Theme
		if t, err = theme.New(f.Tag().String(), ts.path, ts.themeFs, ts.staticFs, ts.rwfs != nil); err != nil {
			err = fmt.Errorf("error loading %v theme: %v - %v", ts.support, ts.path, err)
			return
		}
		mount := "/" + t.Name()
		f.CFeature.MountPoints[mount] = f.CFeature.MountPoints[mount].Append(&feature.CMountPoint{
			Path:  ts.path,
			Mount: mount,
			ROFS:  ts.themeFs,
			RWFS:  ts.rwfs,
		})
		if t.StaticFS() != nil {
			b.RegisterPublicFileSystem("/", t.StaticFS())
		}
		b.AddTheme(t)
		f.loaded = append(f.loaded, t)
		log.DebugF("loaded %v theme: %v", ts.support, t.Name())
	}

	if f.theme != "" {
		b.SetTheme(f.theme)
	}
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

func (f *CFeature) AddLocales(c catalog.Catalog) {
	defTag := f.Enjin.SiteDefaultLanguage()
	for _, t := range f.loaded {
		if locale, ok := t.Locales(); ok {
			c.AddLocalesFromFS(defTag, locale)
		}
	}
	return
}