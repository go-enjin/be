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

package themes

import (
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/features/fs/themes"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/menu"
	"github.com/go-enjin/be/types/editor"
)

var (
	DefaultEditorType = "theme"
	DefaultEditorName = "themes"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "editor-themes"

type Feature interface {
	feature.EditorFeature
}

type MakeFeature interface {
	feature.EditorMakeFeature[MakeFeature]

	Make() Feature
}

type CFeature struct {
	editor.CEditorFeature[MakeFeature]
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
	f.CEditorFeature.Init(this)
	f.CEditorFeature.EditorName = DefaultEditorName
	f.CEditorFeature.EditorType = DefaultEditorType
	f.CEditorFeature.EditAnyFileExtension = true
	return
}

func (f *CFeature) Make() (feat Feature) {
	return f
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CEditorFeature.Startup(ctx); err != nil {
		return
	}

	for _, mp := range f.Enjin.Features().List() {
		if fsf, ok := mp.This().(feature.FileSystemFeature); ok {
			if fsf.BaseTag().Equal(themes.Tag) {
				f.EditingFileSystems = append(f.EditingFileSystems, fsf)
				log.DebugF("%v editing theme filesystem: %v", f.Tag(), fsf.Tag())
			}
		}
	}
	return
}

func (f *CFeature) EditorMenu() (m menu.Menu) {
	m = append(m, &menu.Item{
		Text: f.GetEditorName(),
		Href: f.GetEditorPath(),
		Icon: "fa-solid fa-palette",
	})
	return
}