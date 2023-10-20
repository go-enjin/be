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
	"net/http"

	"github.com/urfave/cli/v2"

	bePkgEditor "github.com/go-enjin/be/pkg/editor"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/feature/signaling"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/types/editor"
)

var (
	DefaultEditorType = "menu"
	DefaultEditorName = "menus"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "editor-menus"

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
	return
}

func (f *CFeature) Make() (feat Feature) {
	return f
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
		}
	}
	return
}

func (f *CFeature) SetupEditor(es feature.EditorSystem) {
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
			f.Editor.PushErrorNotice(eid, printer.Sprintf(`a file name is required; %[1]s supports the following menu files: %[2]s`, t.Name(), filenames), true)
		}
		return
	})
}