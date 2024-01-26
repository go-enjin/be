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

package pages

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
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
	DefaultEditorType = "page"
	DefaultEditorKey  = "pages"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "fs-editor-pages"

type Feature interface {
	feature.EditorFeature
}

type MakeFeature interface {
	feature.EditorMakeFeature[MakeFeature]

	Make() Feature

	AddContentFileSystems(tags ...feature.Tag) MakeFeature
}

type CFeature struct {
	fs_editor.CEditorFeature[MakeFeature]

	contentFsTags feature.Tags

	pageFileSystems []feature.PageFileSystemFeature
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.SetSiteFeatureKey("pages")
	f.SetSiteFeatureIcon("fa-solid fa-file-pen")
	f.SetSiteFeatureLabel(func(printer *message.Printer) (label string) {
		label = printer.Sprintf("Pages")
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

func (f *CFeature) AddContentFileSystems(tags ...feature.Tag) MakeFeature {
	f.contentFsTags = append(f.contentFsTags, tags...)
	f.contentFsTags = f.contentFsTags.Unique()
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

	for _, fp := range f.Enjin.GetFormatProviders() {
		f.EditingFileExtensions = append(f.EditingFileExtensions, fp.ListFormats()...)
	}

	for _, tag := range f.contentFsTags {
		if ef, ok := f.Enjin.Features().Get(tag); ok {
			if fsf, ok := ef.This().(feature.FileSystemFeature); ok {
				if psf, ok := fsf.This().(feature.PageFileSystemFeature); ok {
					f.EditingFileSystems = append(f.EditingFileSystems, fsf)
					f.pageFileSystems = append(f.pageFileSystems, psf)
					log.DebugF("%v editing page filesystem: %v", f.Tag(), psf.Tag())
				} else {
					err = fmt.Errorf("%v feature is not a feature.PageFileSystemFeature", tag)
					return
				}
			} else {
				err = fmt.Errorf("%v feature is not a feature.PageFileSystemFeature", tag)
				return
			}
		} else {
			err = fmt.Errorf("%v feature not found", tag)
			return
		}
	}

	return
}

func (f *CFeature) SetupEditorRoute(r chi.Router) {
	f.CEditorFeature.SetupEditorRoute(r)
	r.Post("/*", f.SelfEditor().ReceiveFileEditorChanges)
}

func (f *CFeature) SetupEditor(es feature.EditorSite) {
	f.CEditorFeature.SetupEditor(es)
	f.FileOperations[bePkgEditor.ChangeActionKey] = &feature.EditorOperation{
		Key:       bePkgEditor.ChangeActionKey,
		Action:    f.UpdateFileAction,
		Validate:  f.OpChangeValidate,
		Operation: f.OpChangeHandler,
	}
	f.FileOperations[bePkgEditor.IndexPageActionKey] = &feature.EditorOperation{
		Key:       bePkgEditor.IndexPageActionKey,
		Confirm:   bePkgEditor.IndexPageActionKey + "-confirmed",
		Action:    f.UpdateFileAction,
		Validate:  f.OpFileIndexValidate,
		Operation: f.OpFileIndexHandler,
	}
	f.FileOperations[bePkgEditor.DeIndexPageActionKey] = &feature.EditorOperation{
		Key:       bePkgEditor.DeIndexPageActionKey,
		Confirm:   bePkgEditor.DeIndexPageActionKey + "-confirmed",
		Action:    f.UpdateFileAction,
		Validate:  f.OpFileDeIndexValidate,
		Operation: f.OpFileDeIndexHandler,
	}
	f.FileOperations[bePkgEditor.CreatePageActionKey] = &feature.EditorOperation{
		Key:       bePkgEditor.CreatePageActionKey,
		Confirm:   bePkgEditor.CreatePageActionKey + "-confirmed",
		Action:    f.CreateFileAction,
		Validate:  f.OpPageCreateValidate,
		Operation: f.OpPageCreateHandler,
	}

	f.Connect(feature.PreMoveFileSignal, "pre-move-page-listener", func(signal signaling.Signal, tag string, data []interface{}, argv []interface{}) (stop bool) {
		if _, _, _, _, info, _, _, ok := feature.ParseSignalArgv(argv); ok {
			f.RemoveIndexing(info)
		}
		return
	})
	//f.Connect(feature.MoveFileSignal, "pre-move-page-listener", func(signal signaling.Signal, tag string, data []interface{}, argv []interface{}) (stop bool) {
	//	// add new index?
	//	if _, _, _, _, info, _, _, ok := feature.ParseSignalArgv(argv); ok {
	//		f.AddIndexing(info)
	//	}
	//	return
	//})
	f.Connect(feature.FileNameRequiredSignal, f.Tag().String()+"--file-name-required-listener", func(signal signaling.Signal, tag string, data []interface{}, argv []interface{}) (stop bool) {
		var eid string
		var r *http.Request
		if r, _, _, _, _, eid, _, stop = feature.ParseSignalArgv(argv); stop {
			printer := lang.GetPrinterFromRequest(r)
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`a file name is required; use "~index" for a directory landing page`))
			stop = true
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
