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
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/feature/filesystem"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/slices"
	"github.com/go-enjin/be/pkg/theme"
)

const Tag feature.Tag = "fs-theme"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type Feature interface {
	filesystem.Feature[MakeFeature]
}

type MakeFeature interface {
	// SetTheme is a convenience method for setting the current theme during the
	// enjin build phase
	SetTheme(name string) MakeFeature

	// AddTheme is a convenience method for adding themes during the enjin build
	// phase
	AddTheme(t *theme.Theme) MakeFeature

	ThemeEmbedSupport
	ThemeLocalSupport

	Make() Feature
}

type CFeature struct {
	filesystem.CFeature[MakeFeature]

	theme      string
	themes     map[string]*theme.Theme
	orderAdded []string
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
	f.themes = make(map[string]*theme.Theme)
}

func (f *CFeature) SetTheme(name string) MakeFeature {
	f.theme = name
	return f
}

func (f *CFeature) AddTheme(t *theme.Theme) MakeFeature {
	if _, already := f.themes[t.Name]; already {
		log.WarnDF(1, "replacing existing %v theme", t.Name)
		f.orderAdded, _ = slices.Prune(f.orderAdded, t.Name)
	}
	f.orderAdded = append(f.orderAdded, t.Name)
	f.themes[t.Name] = t
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	for _, name := range f.orderAdded {
		if t := f.themes[name]; t != nil {
			b.AddTheme(t)
		}
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