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

package editor

import (
	"net/http"
	"strings"

	"golang.org/x/net/html"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/editor"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/forms/nonce"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	beMime "github.com/go-enjin/be/pkg/mime"
	"github.com/go-enjin/be/pkg/path"
	"github.com/go-enjin/be/pkg/userbase"
)

func (f *CEditorFeature[MakeTypedFeature]) PrepareRenderFileEditor(w http.ResponseWriter, r *http.Request) (pg feature.Page, ctx context.Context, info *editor.File, currentUser string, handled bool) {

	currentUser = userbase.GetCurrentUserEID(r)
	printer := lang.GetPrinterFromRequest(r)

	var err error
	var filePath string

	fsid, code, file, locale := f.ParseEditorUrlParams(r)
	if pg, ctx, err = f.SelfEditor().PrepareEditPage("file-editor", f.EditorType, ""); err != nil {
		log.ErrorRF(r, "error preparing %v editor page: %v", f.Tag(), err)
		//f.Enjin.ServeNotFound(w, r)
		f.RenderFileBrowser(w, r)
		handled = true
		return
	} else if fsid == "" {
		//f.Enjin.ServeRedirect(f.SelfEditor().GetEditorPath(), w, r)
		f.RenderFileBrowser(w, r)
		handled = true
		return
	} else if locale == nil {
		//f.Enjin.ServeRedirect(f.SelfEditor().GetEditorPath()+"/"+fsid, w, r)
		f.RenderFileBrowser(w, r)
		handled = true
		return
	} else if file == "" {
		if r.Method != "POST" {
			f.RenderFileBrowser(w, r)
			handled = true
			return
		}

		info = editor.ParseFile(fsid, locale.String())
		parts := strings.Split(info.Path, "/")
		info.MimeType = beMime.DirectoryMimeType
		info.Code = code
		info.Path = strings.Join(append(parts, info.File), "/")
		info.Name = info.File
		info.File = ""

		ctx.SetSpecific("EditFSID", info.FSID)
		ctx.SetSpecific("EditLang", info.Locale.String())
		ctx.SetSpecific("EditPath", info.Path)
		ctx.SetSpecific("EditFile", info.Name)
		ctx.SetSpecific("FileInfo", info)
		return

	} else if filePath = editor.MakeLangCodePath(locale.String(), file); filePath == "" {
		log.ErrorRF(r, "lang-code-path empty: code=%q, file=%q", code, file)
		//f.Enjin.ServeRedirect(f.SelfEditor().GetEditorPath()+"/"+fsid+"/"+code, w, r)
		f.RenderFileBrowser(w, r)
		handled = true
		return
	} else if info = editor.ParseFile(fsid, filePath); info == nil {
		log.ErrorRF(r, "parsed file is nil: fsid=%q, filePath=%q", fsid, filePath)
		//f.Enjin.ServeRedirect(f.SelfEditor().GetEditorPath()+"/"+fsid+"/"+code, w, r)
		f.RenderFileBrowser(w, r)
		handled = true
		return
	} else if info.Code = code; !f.SelfEditor().FileExists(info) {
		if r.Method != "POST" {
			f.RenderFileBrowser(w, r)
			handled = true
			return
		}

		parts := strings.Split(info.Path, "/")
		info.MimeType = beMime.DirectoryMimeType
		info.Path = strings.Join(append(parts, info.File), "/")
		info.Name = info.File
		info.File = ""

		ctx.SetSpecific("EditFSID", info.FSID)
		if info.Locale != nil {
			ctx.SetSpecific("EditLang", info.Locale.String())
		} else {
			ctx.SetSpecific("EditCode", info.Code)
		}
		ctx.SetSpecific("EditPath", info.Path)
		ctx.SetSpecific("EditFile", info.Name)
		ctx.SetSpecific("FileInfo", info)
		return

	} else if !f.EditAnyFileExtension && !path.HasAnyExt(file, f.EditingFileExtensions...) {
		log.ErrorRF(r, "user trying to edit file with unsupported extension: %v - %+v", file, f.EditingFileExtensions)
		f.Editor.PushErrorNotice(currentUser, printer.Sprintf("Unsupported file type."), true)
		f.Enjin.ServeRedirect(f.SelfEditor().GetEditorPath()+"/"+info.EditDirectoryPath(), w, r)
		handled = true
		return
	} else if !f.Enjin.ValidateUserRequest(f.ViewFileAction, w, r) {
		log.WarnRF(r, "user denied: %v", f.ViewFileAction)
		f.Editor.PushErrorNotice(currentUser, printer.Sprintf("Permission denied."), true)
		f.Enjin.ServeRedirect(f.SelfEditor().GetEditorPath()+"/"+info.EditDirectoryPath(), w, r)
		handled = true
		return
	}

	info = f.SelfEditor().PrepareEditableFile(r, info)

	var eid string
	if info.Locked {
		info.ReadOnly = true
		ctx.SetSpecific("EditFileLocked", eid)
	} else if ee := f.LockEditorFile(currentUser, fsid, info.FilePath()); ee != nil {
		info.ReadOnly = true
		f.Editor.PushErrorNotice(eid, printer.Sprintf("error reading file: \"%[1]s\"", ee.Error()), true)
	}

	f.SelfEditor().UpdateFileInfoForEditing(info, r)

	ctx.SetSpecific("EditFSID", info.FSID)
	if info.Locale != nil {
		ctx.SetSpecific("EditLang", info.Locale.String())
	} else {
		ctx.SetSpecific("EditCode", info.Code)
	}
	ctx.SetSpecific("EditPath", info.Path)
	ctx.SetSpecific("EditFile", info.File)
	ctx.SetSpecific("FileInfo", info)

	return
}

func (f *CEditorFeature[MakeTypedFeature]) RenderFileEditor(w http.ResponseWriter, r *http.Request) {

	var pg feature.Page
	var ctx context.Context
	var info *editor.File
	var handled bool
	var eid string
	if pg, ctx, info, eid, handled = f.PrepareRenderFileEditor(w, r); handled {
		return
	}
	printer := lang.GetPrinterFromRequest(r)

	var body string
	if info.HasDraft {
		if data, ee := f.SelfEditor().ReadDraft(info); ee != nil {
			log.ErrorRF(r, "error reading draft: %v - %v", info.FilePath(), ee)
			f.Editor.PushErrorNotice(eid, printer.Sprintf("error reading draft: \"%[1]s\"", ee.Error()), true)
			info.Actions = editor.Actions{}
		} else {
			body = string(data)
			//info.Actions = append(info.Actions, editor.MakePreviewDraftAction(printer))
		}
	} else if data, ee := f.SelfEditor().ReadFile(info); ee != nil {
		log.ErrorRF(r, "error reading file: %v - %v", info.FilePath(), ee)
		f.Editor.PushErrorNotice(eid, printer.Sprintf("error reading file: \"%[1]s\"", ee.Error()), true)
		info.Actions = editor.Actions{}
	} else {
		body = string(data)
	}

	pg.SetTitle(printer.Sprintf("Edit: %[1]s", info.Name))
	ctx.SetSpecific("Body", body)
	r = feature.AddUserNotices(r, f.Editor.PullNotices(eid)...)
	f.SelfEditor().ServePreparedEditPage(pg, ctx, w, r)
}

func (f *CEditorFeature[MakeTypedFeature]) ReceiveFileEditorChanges(w http.ResponseWriter, r *http.Request) {

	var err error
	var pg feature.Page
	var ctx context.Context
	var info *editor.File
	var handled bool
	var eid string
	if pg, ctx, info, eid, handled = f.PrepareRenderFileEditor(w, r); handled {
		return
	}
	printer := lang.GetPrinterFromRequest(r)

	if info.Tilde != "" {
		// deny anything posted to .~stuff
		f.Enjin.ServeRedirect(f.SelfEditor().GetEditorPath()+"/"+info.EditFilePath(), w, r)
		return
	}

	nonceValue := r.PostFormValue("nonce")
	nonceValue = forms.StrictSanitize(nonceValue)
	if !nonce.Validate("file-editor-form", nonceValue) {
		f.Editor.PushErrorNotice(eid, printer.Sprintf("Form expired before submitting, please try again."), true)
		f.Enjin.ServeRedirect(f.SelfEditor().GetEditorPath()+"/"+info.EditFilePath(), w, r)
		return
	}

	//fields := context.GetFields(r)
	action, _ := feature.ParseEditorOpKey(r.PostFormValue("submit"))
	formCtx := map[string]interface{}{}
	for _, k := range maps.SortedKeys(r.Form) {
		v := r.Form[k]
		for i := 0; i < len(v); i++ {
			v[i] = html.UnescapeString(v[i])
		}
		switch len(v) {
		case 0: // nop
		case 1:
			_ = maps.Set(k, v[0], formCtx)
		case 2:
			if v[0] == v[1] {
				_ = maps.Set(k, v[0], formCtx)
			} else {
				_ = maps.Set(k, v, formCtx)
			}
		default:
			_ = maps.Set(k, v, formCtx)
		}
	}

	if modified := f.SelfEditor().PrepareEditableFile(r, info); modified != nil {
		info = modified
	}

	if op, ok := f.FileOperations[action]; ok {
		if !f.Enjin.ValidateUserRequest(op.Action, w, r) {
			log.WarnRF(r, "user denied: %v", op.Action)
			f.Editor.PushErrorNotice(eid, printer.Sprintf("Permission to perform the operation has been denied."), true)
			f.Enjin.ServeRedirect(f.SelfEditor().GetEditorPath()+"/"+info.EditFilePath(), w, r)
			return
		}
		if op.Confirm != "" {
			if _, confirmed := formCtx[op.Confirm]; !confirmed {
				f.Editor.PushErrorNotice(eid, printer.Sprintf("Unconfirmed operation, please confirm before submitting changes."), true)
				f.Enjin.ServeRedirect(f.SelfEditor().GetEditorPath()+"/"+info.EditFilePath(), w, r)
				return
			}
		}
		if op.Validate != nil {
			if err = op.Validate(r, pg, ctx, formCtx, info, eid); err != nil {
				f.Editor.PushErrorNotice(eid, err.Error(), true)
				f.Enjin.ServeRedirect(f.SelfEditor().GetEditorPath()+"/"+info.EditFilePath(), w, r)
				return
			}
		}
		if op.Operation != nil {
			if redirect := op.Operation(r, pg, ctx, formCtx, info, eid); redirect != "" {
				f.Enjin.ServeRedirect(redirect, w, r)
				return
			}
		}
	} else {
		f.Editor.PushErrorNotice(eid, printer.Sprintf("Unknown operation, please try again."), true)
	}

	if v, ok := formCtx["return"].(string); ok {
		if v == "directory" {
			f.Enjin.ServeRedirect(f.SelfEditor().GetEditorPath()+"/"+info.EditDirectoryPath(), w, r)
			return
		}
	}
	f.Enjin.ServeRedirect(f.SelfEditor().GetEditorPath()+"/"+info.EditFilePath(), w, r)
}