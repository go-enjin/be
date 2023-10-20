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
	"html"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/editor"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/forms/nonce"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/request/argv"
	"github.com/go-enjin/be/pkg/slices"
	"github.com/go-enjin/be/pkg/strings/words"
	"github.com/go-enjin/be/pkg/userbase"
	"github.com/go-enjin/golang-org-x-text/language"
)

func (f *CFeature) PrepareRenderFileEditor(w http.ResponseWriter, r *http.Request) (pg feature.Page, ctx context.Context, info *editor.File, eid string, mountPoints feature.MountPoints, handled bool) {

	eid = userbase.GetCurrentUserEID(r)
	printer := lang.GetPrinterFromRequest(r)

	var err error

	fsid := chi.URLParam(r, "fsid")
	code := chi.URLParam(r, "code")

	if fsid == "" {
		f.Enjin.ServeRedirect(f.GetEditorPath(), w, r)
		handled = true
		return
	}

	info = editor.ParseFile(fsid, code)
	info.File = ""
	info.HasDraft = f.HasDraftLocales(fsid, code)

	if locked, lockedBy, ee := f.IsLocaleLocked(fsid, code); ee == nil && locked {
		info.LockedBy = lockedBy
		if info.Locked = lockedBy != eid; info.Locked {
			info.ReadOnly = true
		}
	} else if err = f.LockLocale(eid, fsid, code); err != nil {
		fsidCode := fsid
		if code != "" {
			fsidCode += "/" + code
		}
		f.Editor.PushErrorNotice(eid, printer.Sprintf(`error locking %[1]s locale for editing: %[2]s`, fsidCode, err.Error()), true)
		handled = true
		return
	} else {
		info.LockedBy = eid
	}

	var mountedPoints feature.MountedPoints
	if mountedPoints = f.FindMountedPoints(fsid); mountedPoints == nil {
		f.Enjin.ServeRedirect(f.GetEditorPath(), w, r)
		handled = true
		return
	}

	//var mountPoints feature.MountPoints
	if countMountedPoints := len(mountedPoints); countMountedPoints > 1 {
		if handled = code == ""; handled {
			f.RenderFileBrowser(w, r)
			return
		}
		if v, present := mountedPoints[code]; present {
			mountPoints = v
		} else {
			handled = true
			f.Enjin.ServeRedirect(f.GetEditorPath()+"/"+fsid, w, r)
			return
		}
	} else if countMountedPoints == 1 {
		if v, present := mountedPoints[code]; present {
			mountPoints = v
		} else {
			handled = true
			f.Enjin.ServeRedirect(f.GetEditorPath()+"/"+fsid, w, r)
			return
		}
	}

	if handled = len(mountPoints) == 0; handled {
		log.DebugRF(r, "mount points not found for: fsid=%v, code=%v", fsid, code)
		f.Enjin.ServeRedirect(f.GetEditorPath()+"/"+fsid, w, r)
		return
	}

	if pg, ctx, err = f.SelfEditor().PrepareEditPage("file-editor", f.EditorType, ""); err != nil {
		handled = true
		log.ErrorRF(r, "error preparing %v editor page: %v", f.Tag(), err)
		f.Enjin.ServeNotFound(w, r)
		return
	}

	if mountPoints.HasRWFS() {
		info.Actions = append(info.Actions,
			editor.MakeCommitFileAction(printer),
			editor.MakeCancelFileAction(printer),
		)
	} else {
		info.ReadOnly = true
	}

	f.UpdateFileInfo(info, r)

	ctx.SetSpecific("LocaleInfo", info)

	ctx.SetSpecific("EditFSID", info.FSID)
	if code != "" {
		ctx.SetSpecific("EditCode", code)
	}
	ctx.SetSpecific("EditPath", info.Path)
	//ctx.SetSpecific("EditFile", info.File)
	//ctx.SetSpecific("FileInfo", info)

	ctx.SetSpecific("LocaleSystems", f.ListLocales(r))

	return
}

func (f *CFeature) RenderFileEditor(w http.ResponseWriter, r *http.Request) {
	var pg feature.Page
	var ctx context.Context
	var handled bool
	//var eid string
	var mountPoints feature.MountPoints
	var info *editor.File

	eid := userbase.GetCurrentUserEID(r)
	fsid := chi.URLParam(r, "fsid")
	code := chi.URLParam(r, "code")
	printer := lang.GetPrinterFromRequest(r)

	if pg, ctx, info, _, mountPoints, handled = f.PrepareRenderFileEditor(w, r); handled {
		return
	}

	reqArgv := argv.Get(r)
	ctx.SetSpecific(argv.RequestConsumedKey, true)
	if reqArgv.NumPerPage == -1 {
		reqArgv.NumPerPage = 10
	}
	if reqArgv.PageNumber == -1 {
		reqArgv.PageNumber = 0
	}

	if ld, ee := f.ReadDraftLocales(fsid, code, mountPoints, !info.HasDraft); ee == nil {

		var searchQuery string
		if len(reqArgv.Argv) > 0 {
			if len(reqArgv.Argv[0]) > 0 {
				searchQuery = reqArgv.Argv[0][0]
			}
		}
		if searchQuery != "" {
			wc := words.DefaultConfig()
			order := append([]string{}, ld.Order...)
			for _, shasum := range ld.Order {
				if msgs, ok := ld.Data[shasum]; ok {
					var found bool
					for _, msg := range msgs {
						var checking []string
						checking = append(checking, msg.ID)
						if msg.Translation.Select == nil {
							if msg.Translation.String != "" {
								checking = append(checking, msg.Translation.String)
							}
						} else {
							for _, text := range msg.Translation.Select.Cases {
								checking = append(checking, text)
							}
						}
						for _, check := range checking {
							score, _ := wc.Search(searchQuery, check)
							if found = score > 0; found {
								break
							}
						}
						if found {
							break
						}
					}
					if found {
						continue
					}
					delete(ld.Data, shasum)
					order, _ = slices.Prune(order, shasum)
				}
			}
			ld.Order = order
		}

		totalMsgs := len(ld.Order)
		var totalPages, startItem, endItem int

		if reqArgv.NumPerPage >= totalMsgs {
			totalPages = 1
			startItem = 1
			endItem = totalMsgs
		} else {
			totalPages = totalMsgs / reqArgv.NumPerPage
			if remainder := totalMsgs % reqArgv.NumPerPage; remainder > 0 {
				totalPages += 1
			}

			// there are more translations than the number per page requested
			var order []string
			var start, end int
			if start = reqArgv.NumPerPage * reqArgv.PageNumber; start >= totalMsgs {
				// redirect to the last page number
				reqArgv.PageNumber = totalPages - 1
				f.Enjin.ServeRedirect(reqArgv.String(), w, r)
				return
			}

			if end = start + reqArgv.NumPerPage; end >= totalMsgs {
				// the end point is past the total number of translations
				order = ld.Order[start:]
				startItem = start + 1
				endItem = totalMsgs
			} else {
				// the start and end are within the given range
				order = ld.Order[start:end]
				startItem = start + 1
				endItem = end
			}

			data := make(map[string]map[language.Tag]*LocaleMessage)
			for _, shasum := range order {
				data[shasum] = ld.Data[shasum]
			}

			ld = &LocaleData{
				FSID:  ld.FSID,
				Code:  ld.Code,
				Data:  data,
				Order: order,
			}
		}

		ctx.SetSpecific("Pagination", &argv.Pagination{
			BasePath:    r.URL.Path,
			PageNumber:  reqArgv.PageNumber + 1,
			NumPerPage:  reqArgv.NumPerPage,
			PageIndex:   reqArgv.PageNumber,
			LastIndex:   totalPages - 1,
			TotalItems:  totalMsgs,
			TotalPages:  totalPages,
			StartItem:   startItem,
			EndItem:     endItem,
			SearchQuery: searchQuery,
		})

		ctx.SetSpecific("LocalesData", ld)
		if table, eee := f.MakeTable(ld); eee == nil {
			ctx.SetSpecific("LocalesTable", table)
		} else {
			log.ErrorF("error making locales table: %allMountedPoints", eee)
		}

	}

	pg.SetTitle(printer.Sprintf("Edit: %[1]s", info.EditCodeFilePath()))
	r = feature.AddUserNotices(r, f.Editor.PullNotices(eid)...)
	f.SelfEditor().ServePreparedEditPage(pg, ctx, w, r)
}

func (f *CFeature) ReceiveFileEditorChanges(w http.ResponseWriter, r *http.Request) {

	var err error
	var pg feature.Page
	var ctx context.Context
	var info *editor.File
	var handled bool
	var eid string
	if pg, ctx, info, eid, _, handled = f.PrepareRenderFileEditor(w, r); handled {
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
			v[i] = html.UnescapeString(forms.StrictSanitize(html.UnescapeString(v[i])))
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