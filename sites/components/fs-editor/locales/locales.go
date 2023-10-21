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

package locales

import (
	"github.com/go-chi/chi/v5"
	"github.com/urfave/cli/v2"

	bePkgEditor "github.com/go-enjin/be/pkg/editor"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/request/argv"
	"github.com/go-enjin/be/types/editor"
)

var (
	DefaultEditorType = "locale"
	DefaultEditorName = "locales"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "editor-locales"

type Feature interface {
	feature.EditorFeature
}

type MakeFeature interface {
	feature.EditorMakeFeature[MakeFeature]

	Make() Feature
}

type CFeature struct {
	editor.CEditorFeature[MakeFeature]

	ViewLocaleAction feature.Action
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
	f.ViewLocaleAction = feature.NewAction(f.Tag().String(), "view", "locale")
	return f
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CEditorFeature.Startup(ctx); err != nil {
		return
	}
	f.EditingFileExtensions = []string{"gotext.json"}
	for _, mp := range f.Enjin.GetLocalesProviders() {
		if fsf, ok := mp.This().(feature.FileSystemFeature); ok {
			f.EditingFileSystems = append(f.EditingFileSystems, fsf)
			log.DebugF("%v editing locale filesystem: %v", f.Tag(), fsf.Tag())
		}
	}
	return
}

func (f *CFeature) SetupEditor(es feature.EditorSystem) {
	f.CEditorFeature.SetupEditor(es)
	f.DefaultOp = bePkgEditor.CancelActionKey
	f.FileOperations = map[string]*feature.EditorOperation{
		bePkgEditor.UnlockActionKey: {
			Key:       bePkgEditor.UnlockActionKey,
			Confirm:   bePkgEditor.UnlockActionKey + "-confirmed",
			Action:    f.UpdateFileAction,
			Validate:  f.OpUnlockValidate,
			Operation: f.OpUnlockHandler,
		},
		bePkgEditor.RetakeActionKey: {
			Key:       bePkgEditor.RetakeActionKey,
			Confirm:   bePkgEditor.RetakeActionKey + "-confirmed",
			Action:    f.UpdateFileAction,
			Validate:  f.OpRetakeValidate,
			Operation: f.OpRetakeHandler,
		},
		bePkgEditor.CommitActionKey: {
			Key:       bePkgEditor.CommitActionKey,
			Action:    f.UpdateFileAction,
			Validate:  f.OpCommitValidate,
			Operation: f.OpCommitHandler,
		},
		bePkgEditor.CancelActionKey: {
			Key:       bePkgEditor.CancelActionKey,
			Action:    f.UpdateFileAction,
			Validate:  f.OpCancelValidate,
			Operation: f.OpCancelHandler,
		},
		bePkgEditor.PublishActionKey: {
			Key:       bePkgEditor.PublishActionKey,
			Action:    f.UpdateFileAction,
			Validate:  f.OpPublishValidate,
			Operation: f.OpPublishHandler,
		},
		bePkgEditor.DeleteDraftActionKey: {
			Key:       bePkgEditor.DeleteDraftActionKey,
			Action:    f.DeleteFileAction,
			Validate:  f.OpDeleteValidate,
			Operation: f.OpDeleteHandler,
		},
		bePkgEditor.ChangeActionKey: {
			Key:       bePkgEditor.ChangeActionKey,
			Action:    f.UpdateFileAction,
			Validate:  f.OpChangeValidate,
			Operation: f.OpChangeHandler,
		},
		bePkgEditor.SearchActionKey: {
			Key:       bePkgEditor.SearchActionKey,
			Action:    f.ViewFileAction,
			Validate:  f.OpSearchValidate,
			Operation: f.OpSearchHandler,
		},
	}
}

func (f *CFeature) SetupEditorRoute(r chi.Router) {
	r.Use(f.Enjin.GetPanicHandler().PanicHandler)
	//r.Use(argv.Middleware)

	r.Post("/{fsid:[a-z0-9][-a-z0-9]+?[a-z0-9]*}/{code:[a-zA-Z][-a-zA-Z]+?[a-zA-Z]*}/*", f.SelfEditor().ReceiveFileEditorChanges)
	r.Post("/{fsid:[a-z0-9][-a-z0-9]+?[a-z0-9]*}/{code:[a-zA-Z][-a-zA-Z]+?[a-zA-Z]*}", f.SelfEditor().ReceiveFileEditorChanges)
	r.Post("/{fsid:[a-z0-9][-a-z0-9]+?[a-z0-9]*}/", f.SelfEditor().ReceiveFileEditorChanges)
	r.Post("/{fsid:[a-z0-9][-a-z0-9]+?[a-z0-9]*}", f.SelfEditor().ReceiveFileEditorChanges)

	for _, path := range []string{
		`/{fsid:[a-z0-9][-a-z0-9]+?[a-z0-9]*}/{code:[a-z0-9][-a-z0-9]+?[a-z0-9]*}`,
		`/{fsid:[a-z0-9][-a-z0-9]+?[a-z0-9]*}`,
	} {
		r.Get(path+"/"+argv.RouteOneArg+"/"+argv.RoutePgntn+"/", f.RenderFileEditor)
		r.Get(path+"/"+argv.RouteOneArg+"/"+argv.RoutePgntn, f.RenderFileEditor)
		r.Get(path+"/"+argv.RoutePgntn+"/", f.RenderFileEditor)
		r.Get(path+"/"+argv.RoutePgntn, f.RenderFileEditor)
		r.Get(path+"/"+argv.RouteOneArg+"/", f.RenderFileEditor)
		r.Get(path+"/"+argv.RouteOneArg, f.RenderFileEditor)
		r.Get(path+"/", f.RenderFileEditor)
		r.Get(path, f.RenderFileEditor)
	}

	r.Get("/", f.SelfEditor().RenderFileBrowser)
}