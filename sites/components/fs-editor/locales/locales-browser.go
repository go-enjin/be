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
	"net/http"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/editor"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/mime"
	"github.com/go-enjin/be/pkg/userbase"
)

func (f *CFeature) RenderFileBrowser(w http.ResponseWriter, r *http.Request) {

	if !f.Enjin.ValidateUserRequest(f.ViewBrowserAction, w, r) {
		log.WarnRF(r, "user denied: %v", f.ViewBrowserAction)
		f.Enjin.ServeNotFound(w, r)
		return
	}

	var err error
	var pg feature.Page
	var ctx context.Context

	if pg, ctx, err = f.SelfEditor().PrepareEditPage("fs-editor--file-browser", f.EditorType, ""); err != nil {
		log.ErrorRF(r, "error preparing %v editor page: %v", f.Tag(), err)
		f.Enjin.ServeNotFound(w, r)
		return
	}

	fsid, code, _, _ := f.ParseEditorUrlParams(r)

	eid := userbase.GetCurrentUserEID(r)

	ctx.SetSpecific("EditorEID", eid)
	ctx.SetSpecific("EditFSID", fsid)
	ctx.SetSpecific("EditLang", code)

	info := &editor.File{
		FSID:     fsid,
		Code:     code,
		MimeType: mime.DirectoryMimeType,
	}
	f.UpdatePathInfo(info, r)
	ctx.SetSpecific("BrowseInfo", info)
	ctx.SetSpecific("PageActions", info.Actions)

	var files editor.Files

	var titlePath string
	if fsid == "" {
		titlePath = f.SelfEditor().GetEditorName()
		files = append(files, f.ListLocaleFileSystems(r)...)
	} else if code == "" {
		titlePath = fsid
		files = append(files, &editor.File{
			Name:     "..",
			MimeType: mime.DirectoryMimeType,
		})
		files = append(files, f.ListLocaleFileSystemLocales(r, fsid)...)
	}

	ctx.SetSpecific("EditFiles", files)
	printer := lang.GetPrinterFromRequest(r)
	r = feature.AddUserNotices(r, f.Editor.PullNotices(eid)...)
	pg.SetTitle(printer.Sprintf("Browsing: %[1]s", titlePath))
	f.SelfEditor().ServePreparedEditPage(pg, ctx, w, r)
}