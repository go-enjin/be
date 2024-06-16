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

package feature

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/go-corelibs/x-text/language"
	"github.com/go-corelibs/x-text/message"

	clContext "github.com/go-corelibs/context"
	"github.com/go-enjin/be/pkg/feature/signaling"
	"github.com/go-enjin/be/pkg/menu"
	"github.com/go-enjin/be/types/page/matter"
)

type EditorFeature interface {
	SiteFeature
	signaling.Signaling
	UserActionsProvider

	SelfEditor() (self EditorFeature)
	SetupEditor(editor EditorSite)
	SetupEditorRoute(r chi.Router)
	EditorMenu(r *http.Request) (m menu.Menu)

	GetEditorKey() (name string)
	GetEditorPath() (path string)
	GetEditorMenu() (m menu.Menu)

	PrepareEditPage(pageType, editorType string, r *http.Request) (pg Page, ctx clContext.Context, err error)
	ParseEditorUrlParams(r *http.Request) (fsid, code, file string, locale *language.Tag)
	ServePreparedEditPage(pg Page, ctx clContext.Context, w http.ResponseWriter, r *http.Request)

	UpdatePathInfo(info *EditorFile, r *http.Request)
	UpdateFileInfo(info *EditorFile, r *http.Request)
	UpdateFileInfoForEditing(info *EditorFile, r *http.Request)
	PrepareEditableFile(r *http.Request, info *EditorFile) (editFile *EditorFile)

	ListFileSystems() (list EditorFiles)
	ListFileSystemLocales(fsid string) (list EditorFiles)
	ListFileSystemDirectories(r *http.Request, fsid, code, dirs string) (list EditorFiles)
	ListFileSystemFiles(r *http.Request, fsid, code, dirs string) (list EditorFiles)
	ProcessMountPointFile(r *http.Request, printer *message.Printer, eid, mpfBTag, mpfTag, code, dirs, file string, mountPoint *CMountPoint, draftWork bool) (ef *EditorFile, ignored bool)

	LockEditorFile(eid, fsid, filePath string) (err error)
	IsEditorFileLocked(fsid, filePath string) (eid string, locked bool)
	UnLockEditorFile(fsid, filePath string) (err error)

	FileExists(info *EditorFile) (exists bool)
	ReadFile(info *EditorFile) (data []byte, err error)
	WriteFile(info *EditorFile, data []byte) (err error)
	RemoveFile(info *EditorFile) (err error)
	RemoveDirectory(info *EditorFile) (err error)

	DraftExists(info *EditorFile) (present bool)
	ReadDraft(info *EditorFile) (contents []byte, err error)
	ReadDraftMatter(info *EditorFile) (pm *matter.PageMatter, err error)
	WriteDraft(info *EditorFile, contents []byte) (err error)
	RemoveDraft(info *EditorFile) (err error)
	PublishDraft(info *EditorFile) (err error)

	RenderFileBrowser(w http.ResponseWriter, r *http.Request)
	RenderFileEditor(w http.ResponseWriter, r *http.Request)
	ReceiveFileEditorChanges(w http.ResponseWriter, r *http.Request)

	OpFileUnlockHandler(r *http.Request, pg Page, ctx, form clContext.Context, info *EditorFile, eid string) (redirect string)
	OpFileRetakeHandler(r *http.Request, pg Page, ctx, form clContext.Context, info *EditorFile, eid string) (redirect string)
	OpFileDeleteValidate(r *http.Request, pg Page, ctx, form clContext.Context, info *EditorFile, eid string) (err error)
	OpFileDeleteHandler(r *http.Request, pg Page, ctx, form clContext.Context, info *EditorFile, eid string) (redirect string)
	OpFileCommitValidate(r *http.Request, pg Page, ctx, form clContext.Context, info *EditorFile, eid string) (err error)
	OpFileCommitHandler(r *http.Request, pg Page, ctx, form clContext.Context, info *EditorFile, eid string) (redirect string)
	OpFilePublishValidate(r *http.Request, pg Page, ctx, form clContext.Context, info *EditorFile, eid string) (err error)
	OpFilePublishHandler(r *http.Request, pg Page, ctx, form clContext.Context, info *EditorFile, eid string) (redirect string)
	OpFileCancelValidate(r *http.Request, pg Page, ctx, form clContext.Context, info *EditorFile, eid string) (err error)
	OpFileCancelHandler(r *http.Request, pg Page, ctx, form clContext.Context, info *EditorFile, eid string) (redirect string)
	OpFileMoveValidate(r *http.Request, pg Page, ctx, form clContext.Context, info *EditorFile, eid string) (err error)
	OpFileMoveHandler(r *http.Request, pg Page, ctx, form clContext.Context, info *EditorFile, eid string) (redirect string)
	OpFileCopyValidate(r *http.Request, pg Page, ctx, form clContext.Context, info *EditorFile, eid string) (err error)
	OpFileCopyHandler(r *http.Request, pg Page, ctx, form clContext.Context, info *EditorFile, eid string) (redirect string)
	OpFileTranslateValidate(r *http.Request, pg Page, ctx, form clContext.Context, info *EditorFile, eid string) (err error)
	OpFileTranslateHandler(r *http.Request, pg Page, ctx, form clContext.Context, info *EditorFile, eid string) (redirect string)

	OpPathDeleteValidate(r *http.Request, pg Page, ctx, form clContext.Context, info *EditorFile, eid string) (err error)
	OpPathDeleteHandler(r *http.Request, pg Page, ctx, form clContext.Context, info *EditorFile, eid string) (redirect string)
}

type EditorMakeFeature[MakeTypedFeature interface{}] interface {
	SiteMakeFeature[MakeTypedFeature]

	SetEditorName(name string) MakeTypedFeature
	SetEditorType(editorType string) MakeTypedFeature
	SetEditingTags(tags ...Tag) MakeTypedFeature
}
