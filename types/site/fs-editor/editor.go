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

package fs_editor

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/go-enjin/golang-org-x-text/language"

	beContext "github.com/go-enjin/be/pkg/context"
	bePkgEditor "github.com/go-enjin/be/pkg/editor"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/menu"
	"github.com/go-enjin/be/types/site"
)

var (
	_ feature.EditorFeature = (*CEditorFeature[feature.EditorMakeFeature[feature.EditorFeature]])(nil)
)

type CEditorFeature[MakeTypedFeature interface{}] struct {
	site.CSiteFeature[MakeTypedFeature]

	EditorKey             string
	EditorType            string
	EditorTags            feature.Tags
	EditingFileSystems    []feature.FileSystemFeature
	EditingFileExtensions []string
	EditAnyFileExtension  bool

	Editor feature.EditorSite

	ViewBrowserAction feature.Action
	ViewFileAction    feature.Action
	CreateFileAction  feature.Action
	UpdateFileAction  feature.Action
	DeleteFileAction  feature.Action

	DefaultOp      string
	FileOperations map[string]*feature.EditorOperation
}

func (f *CEditorFeature[MakeTypedFeature]) Construct(this interface{}) {
	f.CSiteFeature.Construct(this)
	return
}

func (f *CEditorFeature[MakeTypedFeature]) Init(this interface{}) {
	f.CSiteFeature.Init(this)
	f.EditorType = "unimplemented"
	return
}

func (f *CEditorFeature[MakeTypedFeature]) SetEditorName(name string) MakeTypedFeature {
	f.EditorKey = name
	typed, _ := f.This().(MakeTypedFeature)
	return typed
}

func (f *CEditorFeature[MakeTypedFeature]) SetEditorType(editorType string) MakeTypedFeature {
	f.EditorType = editorType
	typed, _ := f.This().(MakeTypedFeature)
	return typed
}

func (f *CEditorFeature[MakeTypedFeature]) SetEditingTags(tags ...feature.Tag) MakeTypedFeature {
	f.EditorTags = tags
	typed, _ := f.This().(MakeTypedFeature)
	return typed
}

func (f *CEditorFeature[MakeTypedFeature]) Build(b feature.Buildable) (err error) {
	if err = f.CSiteFeature.Build(b); err != nil {
		return
	}
	if f.EditorKey == "" {
		f.EditorKey = f.Tag().Spaced()
	}
	return
}

func (f *CEditorFeature[MakeTypedFeature]) UserActions() (list feature.Actions) {
	list = feature.Actions{
		f.ViewBrowserAction,
		f.ViewFileAction,
		f.CreateFileAction,
		f.UpdateFileAction,
		f.DeleteFileAction,
	}
	return
}

func (f *CEditorFeature[MakeTypedFeature]) SiteFeatureInfo(r *http.Request) (info *feature.CSiteFeatureInfo) {
	printer := lang.GetPrinterFromRequest(r)
	info = feature.NewSiteFeatureInfo(
		f.Self().Tag().Kebab(),
		f.SelfEditor().SiteFeatureKey(),
		f.SelfEditor().SiteFeatureIcon(),
		f.SelfEditor().SiteFeatureLabel(printer),
	)
	return
}

func (f *CEditorFeature[MakeTypedFeature]) SiteFeatureKey() (key string) {
	key = f.GetEditorKey()
	return
}

func (f *CEditorFeature[MakeTypedFeature]) SiteFeatureMenu(r *http.Request) (m menu.Menu) {
	info := f.SiteFeatureInfo(r)
	m = menu.Menu{
		{
			Text: info.Label,
			Href: f.SelfEditor().GetEditorPath(),
			Icon: info.Icon,
		},
	}
	return
}

func (f *CEditorFeature[MakeTypedFeature]) SelfEditor() (self feature.EditorFeature) {
	self, _ = f.This().(feature.EditorFeature)
	return
}

func (f *CEditorFeature[MakeTypedFeature]) GetEditorKey() (name string) {
	return f.EditorKey
}

func (f *CEditorFeature[MakeTypedFeature]) GetEditorPath() (path string) {
	return f.Editor.SiteFeaturePath() + "/" + f.EditorKey
}

func (f *CEditorFeature[MakeTypedFeature]) GetEditorMenu() (m menu.Menu) {
	return nil
}

func (f *CEditorFeature[MakeTypedFeature]) EditorMenu(r *http.Request) (m menu.Menu) {
	m = f.SiteFeatureMenu(r)
	return
}

func (f *CEditorFeature[MakeTypedFeature]) SetupEditor(es feature.EditorSite) {
	f.Editor = es

	f.ViewBrowserAction = f.Action("view", "file-browser")
	f.ViewFileAction = f.Action("view", "file-editor")
	f.CreateFileAction = f.Action("create", "file-editor")
	f.UpdateFileAction = f.Action("edit", "file-editor")
	f.DeleteFileAction = f.Action("delete", "file-editor")

	f.DefaultOp = bePkgEditor.CancelActionKey
	f.FileOperations = map[string]*feature.EditorOperation{
		bePkgEditor.UnlockActionKey: {
			Key:       bePkgEditor.UnlockActionKey,
			Confirm:   bePkgEditor.UnlockActionKey + "-confirmed",
			Action:    f.UpdateFileAction,
			Operation: f.SelfEditor().OpFileUnlockHandler,
		},
		bePkgEditor.RetakeActionKey: {
			Key:       bePkgEditor.RetakeActionKey,
			Confirm:   bePkgEditor.RetakeActionKey + "-confirmed",
			Action:    f.UpdateFileAction,
			Operation: f.SelfEditor().OpFileRetakeHandler,
		},
		bePkgEditor.DeleteActionKey: {
			Key:       bePkgEditor.DeleteActionKey,
			Confirm:   bePkgEditor.DeleteActionKey + "-confirmed",
			Action:    f.DeleteFileAction,
			Validate:  f.SelfEditor().OpFileDeleteValidate,
			Operation: f.SelfEditor().OpFileDeleteHandler,
		},
		bePkgEditor.DeletePathActionKey: {
			Key:       bePkgEditor.DeletePathActionKey,
			Confirm:   bePkgEditor.DeletePathActionKey + "-confirmed",
			Action:    f.DeleteFileAction,
			Validate:  f.SelfEditor().OpPathDeleteValidate,
			Operation: f.SelfEditor().OpPathDeleteHandler,
		},
		bePkgEditor.DeleteDraftActionKey: {
			Key:       bePkgEditor.DeleteDraftActionKey,
			Confirm:   bePkgEditor.DeleteDraftActionKey + "-confirmed",
			Action:    f.DeleteFileAction,
			Validate:  f.SelfEditor().OpFileDeleteValidate,
			Operation: f.SelfEditor().OpFileDeleteHandler,
		},
		bePkgEditor.CommitActionKey: {
			Key:       bePkgEditor.CommitActionKey,
			Action:    f.UpdateFileAction,
			Validate:  f.SelfEditor().OpFileCommitValidate,
			Operation: f.SelfEditor().OpFileCommitHandler,
		},
		bePkgEditor.PublishActionKey: {
			Key:       bePkgEditor.PublishActionKey,
			Confirm:   bePkgEditor.PublishActionKey + "-confirmed",
			Action:    f.UpdateFileAction,
			Validate:  f.SelfEditor().OpFilePublishValidate,
			Operation: f.SelfEditor().OpFilePublishHandler,
		},
		bePkgEditor.CancelActionKey: {
			Key:       bePkgEditor.CancelActionKey,
			Action:    f.UpdateFileAction,
			Validate:  f.SelfEditor().OpFileCancelValidate,
			Operation: f.SelfEditor().OpFileCancelHandler,
		},
		bePkgEditor.MoveActionKey: {
			Key:       bePkgEditor.MoveActionKey,
			Confirm:   bePkgEditor.MoveActionKey + "-confirmed",
			Action:    f.DeleteFileAction,
			Validate:  f.SelfEditor().OpFileMoveValidate,
			Operation: f.SelfEditor().OpFileMoveHandler,
		},
		bePkgEditor.CopyActionKey: {
			Key:       bePkgEditor.CopyActionKey,
			Confirm:   bePkgEditor.CopyActionKey + "-confirmed",
			Action:    f.CreateFileAction,
			Validate:  f.SelfEditor().OpFileCopyValidate,
			Operation: f.SelfEditor().OpFileCopyHandler,
		},
		bePkgEditor.TranslateActionKey: {
			Key:       bePkgEditor.TranslateActionKey,
			Confirm:   bePkgEditor.TranslateActionKey + "-confirmed",
			Action:    f.CreateFileAction,
			Validate:  f.SelfEditor().OpFileTranslateValidate,
			Operation: f.SelfEditor().OpFileTranslateHandler,
		},
	}
	return
}

func (f *CEditorFeature[MakeTypedFeature]) SetupEditorRoute(r chi.Router) {
	r.Use(f.Enjin.GetPanicHandler().PanicHandler)
	r.Post("/{fsid:[a-z0-9][-a-z0-9]+?[a-z0-9]*}/{lang:[a-zA-Z][-a-zA-Z]+?[a-zA-Z]*}/*", f.SelfEditor().ReceiveFileEditorChanges)
	r.Get("/{fsid:[a-z0-9][-a-z0-9]+?[a-z0-9]*}/{lang:[a-zA-Z][-a-zA-Z]+?[a-zA-Z]*}/*", f.SelfEditor().RenderFileEditor)
	r.Post("/{fsid:[a-z0-9][-a-z0-9]+?[a-z0-9]*}/{lang:[a-zA-Z][-a-zA-Z]+?[a-zA-Z]*}/", f.SelfEditor().ReceiveFileEditorChanges)
	r.Get("/{fsid:[a-z0-9][-a-z0-9]+?[a-z0-9]*}/{lang:[a-zA-Z][-a-zA-Z]+?[a-zA-Z]*}/", f.SelfEditor().RenderFileBrowser)
	r.Post("/{fsid:[a-z0-9][-a-z0-9]+?[a-z0-9]*}/{lang:[a-zA-Z][-a-zA-Z]+?[a-zA-Z]*}", f.SelfEditor().ReceiveFileEditorChanges)
	r.Get("/{fsid:[a-z0-9][-a-z0-9]+?[a-z0-9]*}/{lang:[a-zA-Z][-a-zA-Z]+?[a-zA-Z]*}", f.SelfEditor().RenderFileBrowser)
	r.Get("/{fsid:[a-z0-9][-a-z0-9]+?[a-z0-9]*}/", f.SelfEditor().RenderFileBrowser)
	r.Get("/{fsid:[a-z0-9][-a-z0-9]+?[a-z0-9]*}/", f.SelfEditor().RenderFileBrowser)
	r.Get("/{fsid:[a-z0-9][-a-z0-9]+?[a-z0-9]*}", f.SelfEditor().RenderFileBrowser)
	r.Get("/", f.SelfEditor().RenderFileBrowser)
}

func (f *CEditorFeature[MakeTypedFeature]) PrepareEditPage(pageType, editorType string, r *http.Request) (pg feature.Page, ctx beContext.Context, err error) {
	printer := lang.GetPrinterFromRequest(r)
	pg, ctx, err = f.Editor.Site().PreparePage(f.Editor.BaseTag().Kebab(), pageType, editorType, f.Editor.SiteFeatureTheme(), r)

	ctx.SetSpecific("EditorType", editorType)
	ctx.SetSpecific("EditorFeatureKey", f.SelfEditor().GetEditorKey())
	ctx.SetSpecific("EditorFeaturePath", f.SelfEditor().GetEditorPath())
	ctx.SetSpecific("EditorFeatureLabel", f.SelfEditor().SiteFeatureLabel(printer))

	fsids := map[string]string{}
	for _, nfo := range f.SelfEditor().ListFileSystems() {
		fsids[nfo.FSID] = nfo.FSBT
	}
	ctx.SetSpecific("AllFSIDs", maps.SortedKeys(fsids))
	ctx.SetSpecific("FSBTLookup", fsids)

	locales := map[string]struct{}{}
	for _, tag := range f.Enjin.SiteLocales() {
		locales[tag.String()] = struct{}{}
	}
	ctx.SetSpecific("AllLocales", maps.SortedKeys(locales))
	return
}

func (f *CEditorFeature[MakeTypedFeature]) ParseEditorUrlParams(r *http.Request) (fsid, code, file string, locale *language.Tag) {
	fsids := make(map[string]struct{})
	for _, efs := range f.EditingFileSystems {
		fsids[efs.Tag().String()] = struct{}{}
	}
	if fsid = chi.URLParam(r, "fsid"); fsid == "" {
		// no fsid, no other url params matter
		return
	} else if _, validFSID := fsids[fsid]; !validFSID {
		fsid = ""
		return
	} else if code = chi.URLParam(r, "lang"); code == "" {
		if code = chi.URLParam(r, "code"); code == "" {
			// no language or code, file doesn't matter
			return
		}
	}
	if tag, err := language.Parse(code); err == nil {
		locale = &tag
	} else {
		locale = &language.Und
	}
	file = chi.URLParam(r, "*")
	return
}

func (f *CEditorFeature[MakeTypedFeature]) ServePreparedEditPage(pg feature.Page, ctx beContext.Context, w http.ResponseWriter, r *http.Request) {
	f.Editor.Site().ServePreparedPage(pg, ctx, f.Editor.SiteFeatureTheme(), w, r)
}

func (f *CEditorFeature[MakeTypedFeature]) ParseCopyMoveTranslateForm(r *http.Request, pg feature.Page, ctx, form beContext.Context, info *bePkgEditor.File, eid string, redirect *string) (srcUri, dstUri string, dstInfo *bePkgEditor.File, srcFS, dstFS feature.FileSystemFeature, srcMP, dstMP *feature.CMountPoint, srcExists, dstExists bool, stop bool) {
	printer := lang.GetPrinterFromRequest(r)

	var param string
	if submit, ok := form["submit"]; ok && submit == bePkgEditor.CopyActionKey {
		param = bePkgEditor.CopyActionKey
	} else if submit == bePkgEditor.MoveActionKey {
		param = bePkgEditor.MoveActionKey
	} else if submit == bePkgEditor.TranslateActionKey {
		param = bePkgEditor.TranslateActionKey
	} else {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`inconsistent operation requested`))
		stop = true
		return
	}

	var fsid, code, filePath, fileName, fullPath, dstPath string
	fsid, _ = form.FirstString(param + "~dst-fsid")
	code, _ = form.FirstString(param + "~dst-lang")
	filePath, _ = form.FirstString(param + "~dst-path")
	if fileName, _ = form.FirstString(param + "~dst-name"); fileName == "" {
		if stop = f.Emit(feature.FileNameRequiredSignal, f.Tag().String(), r, pg, ctx, form, info, eid, redirect); stop {
			return
		}
		f.Editor.Site().PushWarnNotice(eid, true, printer.Sprintf(`a file name is required`))
		return
	}

	fsid = forms.StrictCleanKebabValue(fsid)
	fileName = forms.StrictCleanKebabValue(fileName)
	if filePath = forms.KebabRelativePath(filePath); filePath != "" {
		fullPath = filePath + "/" + fileName
	} else {
		fullPath = fileName
	}

	t := f.Enjin.MustGetTheme()
	if _, matched := t.MatchFormat(info.File); matched != "" {
		fullPath += "." + matched
	}

	if param == bePkgEditor.TranslateActionKey {
		if t, err := language.Parse(code); err != nil {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`invalid language code given`))
			stop = true
			return
		} else {
			dstPath = t.String() + "/" + fullPath
		}
	} else {
		dstPath = info.Locale.String() + "/" + fullPath
	}

	srcUri, dstUri = info.FSID+"://"+info.FilePath(), fsid+"://"+dstPath
	if stop = srcUri == dstUri; stop {
		f.Editor.Site().PushWarnNotice(eid, true, printer.Sprintf(`"%[1]s" and destination are the same, nothing to do!`, srcUri))
		return
	}

	dstInfo = bePkgEditor.ParseFile(fsid, dstPath)

	for _, efs := range f.EditingFileSystems {
		if !dstExists && efs.Tag().String() == dstInfo.FSID {
			dstFS = efs
			for _, mps := range efs.GetMountedPoints() {
				for _, mp := range mps {
					// TODO: figure out mount point prefix
					if dstExists = mp.ROFS.Exists(dstInfo.FilePath()); mp.RWFS != nil {
						dstMP = mp
						break
					}
				}
				if dstExists || dstMP != nil {
					break
				}
			}
		}
		if !srcExists && efs.Tag().String() == info.FSID {
			srcFS = efs
			for _, mps := range efs.GetMountedPoints() {
				for _, mp := range mps {
					// TODO: figure out mount point prefix
					if srcExists = mp.ROFS.Exists(info.FilePath()); srcExists {
						srcMP = mp
						break
					}
				}
				if srcExists && srcMP != nil {
					break
				}
			}
		}
		if dstExists && srcExists {
			break
		}
	}
	return
}