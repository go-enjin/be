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

package menus

import (
	"fmt"
	"net/http"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/golang-org-x-text/message"

	bePkgEditor "github.com/go-enjin/be/pkg/editor"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/feature/signaling"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/menu"
	fs_editor "github.com/go-enjin/be/types/site/fs-editor"
)

var (
	DefaultEditorType = "menu"
	DefaultEditorKey  = "menus"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "fs-editor-menus"

type Feature interface {
	feature.EditorFeature
}

type MakeFeature interface {
	feature.EditorMakeFeature[MakeFeature]

	AddMenuFileSystems(features ...feature.Feature) MakeFeature
	AddMenuFileSystemsByTag(tags ...feature.Tag) MakeFeature

	Make() Feature
}

type CFeature struct {
	fs_editor.CEditorFeature[MakeFeature]

	extraTags feature.Tags
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.SetSiteFeatureKey("menus")
	f.SetSiteFeatureIcon("fa-solid fa-bars-staggered")
	f.SetSiteFeatureLabel(func(printer *message.Printer) (label string) {
		label = printer.Sprintf("Menus")
		return
	})
	f.CEditorFeature.Construct(f)
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CEditorFeature.Init(this)
	f.CEditorFeature.EditorKey = DefaultEditorKey
	f.CEditorFeature.EditorType = DefaultEditorType
	return
}

func (f *CFeature) AddMenuFileSystems(features ...feature.Feature) MakeFeature {
	for _, ef := range features {
		if efs, ok := ef.This().(feature.FileSystemFeature); ok {
			f.EditingFileSystems = append(f.EditingFileSystems, efs)
		}
	}
	return f
}

func (f *CFeature) AddMenuFileSystemsByTag(tags ...feature.Tag) MakeFeature {
	f.extraTags = f.extraTags.Append(tags...)
	return f
}

func (f *CFeature) Make() (feat Feature) {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	if err = f.CEditorFeature.Build(b); err != nil {
		return
	}
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CEditorFeature.Startup(ctx); err != nil {
		return
	}
	f.EditingFileExtensions = []string{"json"}
	for _, mp := range f.Enjin.GetMenuProviders() {
		if fsf, ok := mp.This().(feature.FileSystemFeature); ok {
			f.EditingFileSystems = append(f.EditingFileSystems, fsf)
			log.DebugF("%v editing filesystem: %v", f.Tag(), fsf.Tag())
		} else {
			err = fmt.Errorf("not a feature.FileSystemFeature: %q", mp.Tag())
			return
		}
	}
	for _, extra := range f.extraTags {
		if ef, ok := f.Enjin.Features().Get(extra); ok {
			if efs, ok := ef.This().(feature.FileSystemFeature); ok {
				f.EditingFileSystems = append(f.EditingFileSystems, efs)
			} else {
				err = fmt.Errorf("not a feature.FileSystemFeature: %q", extra)
				return
			}
		} else {
			err = fmt.Errorf("feature not found: %q", extra)
			return
		}
	}
	return
}

func (f *CFeature) SetupEditor(es feature.EditorSite) {
	f.CEditorFeature.SetupEditor(es)

	f.FileOperations[bePkgEditor.ChangeActionKey] = &feature.EditorOperation{
		Key:       bePkgEditor.ChangeActionKey,
		Action:    f.UpdateFileAction,
		Operation: f.OpChangeHandler,
	}
	f.FileOperations[bePkgEditor.CreateMenuActionKey] = &feature.EditorOperation{
		Key:       bePkgEditor.CreateMenuActionKey,
		Confirm:   bePkgEditor.CreateMenuActionKey + "-confirmed",
		Action:    f.CreateFileAction,
		Validate:  f.OpMenuCreateValidate,
		Operation: f.OpMenuCreateHandler,
	}

	f.Connect(feature.FileNameRequiredSignal, f.Tag().String()+"--file-name-required-listener", func(signal signaling.Signal, tag string, data []interface{}, argv []interface{}) (stop bool) {
		var eid string
		var r *http.Request
		if r, _, _, _, _, eid, _, stop = feature.ParseSignalArgv(argv); stop {
			t := f.Enjin.MustGetTheme()
			var filenames string
			supported := t.GetConfig().Supports.Menus
			for idx, key := range supported.Keys() {
				if idx > 0 {
					filenames += ", "
				}
				filenames += key + ".json"
			}
			printer := lang.GetPrinterFromRequest(r)
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`a file name is required; %[1]s supports the following menu files: %[2]s`, t.Name(), filenames))
		}
		return
	})
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
