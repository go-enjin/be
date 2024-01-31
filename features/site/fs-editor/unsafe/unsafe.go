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

package unsafe

import (
	"net/http"

	"github.com/urfave/cli/v2"

	"github.com/go-corelibs/x-text/message"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/menu"
	fs_editor "github.com/go-enjin/be/types/site/fs-editor"
)

var (
	DefaultEditorType = "unsafe"
	DefaultEditorKey  = "unsafe"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "fs-editor-unsafe"

type Feature interface {
	feature.EditorFeature
}

type MakeFeature interface {
	feature.EditorMakeFeature[MakeFeature]

	Make() Feature
}

type CFeature struct {
	fs_editor.CEditorFeature[MakeFeature]
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.SetSiteFeatureKey("unsafe")
	f.SetSiteFeatureIcon("fa-solid fa-dumpster-fire")
	f.SetSiteFeatureLabel(func(printer *message.Printer) (label string) {
		label = printer.Sprintf("Unsafe")
		return
	})
	f.CEditorFeature.Construct(f)
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CEditorFeature.Init(this)
	f.CEditorFeature.EditorKey = DefaultEditorKey
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
			if fsf.IsLocalized() {
				f.EditingFileSystems = append(f.EditingFileSystems, fsf)
				log.DebugF("%v editing filesystem: %v", f.Tag(), fsf.Tag())
			}
		}
	}
	return
}

func (f *CFeature) SiteFeatureMenu(r *http.Request) (m menu.Menu) {
	info := f.SiteFeatureInfo(r)
	m = menu.Menu{
		{
			Text: info.Label,
			Href: f.GetEditorPath(),
			Icon: info.Icon,
		},
	}
	return
}

func (f *CFeature) EditorMenu(r *http.Request) (m menu.Menu) {
	m = f.SiteFeatureMenu(r)
	return
}
