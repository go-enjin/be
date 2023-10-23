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
	"net/http"

	beContext "github.com/go-enjin/be/pkg/context"
	bePkgEditor "github.com/go-enjin/be/pkg/editor"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/types/page"
	"github.com/go-enjin/be/types/page/matter"
)

func (f *CFeature) RenderFilePreview(w http.ResponseWriter, r *http.Request) {
	var ctx beContext.Context
	var info *bePkgEditor.File
	var handled bool
	var eid string
	if _, ctx, info, eid, handled = f.PrepareRenderFileEditor(w, r); handled {
		return
	}
	printer := lang.GetPrinterFromRequest(r)

	var err error
	var pm *matter.PageMatter
	if pm, err = f.SelfEditor().ReadDraftMatter(info); err != nil {
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf(`error reading draft page matter: "%[1]s"`, err.Error()), false)
		r = feature.AddUserNotices(r, f.Editor.Site().PullNotices(eid)...)
		f.Enjin.ServeRedirect(f.SelfEditor().GetEditorPath()+"/"+info.EditFilePath(), w, r)
		return
	}

	var p feature.Page
	if p, err = page.NewFromPageMatter(pm, f.Editor.SiteFeatureTheme(), f.Enjin.Context()); err != nil {
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf(`error preparing draft page preview: "%[1]s"`, err.Error()), false)
		r = feature.AddUserNotices(r, f.Editor.Site().PullNotices(eid)...)
		f.Enjin.ServeRedirect(f.SelfEditor().GetEditorPath()+"/"+info.EditFilePath(), w, r)
		return
	}

	if info.Locked {
		r = feature.AddUserNotices(r, feature.MakeWarnNotice(
			printer.Sprintf("Draft Preview"),
			false,
			feature.UserNoticeLink{
				Text: printer.Sprintf("click here to return to the file-browser"),
				Href: f.SelfEditor().GetEditorPath() + "/" + info.EditDirectoryPath(),
			},
		))
	} else {
		r = feature.AddUserNotices(r, feature.MakeWarnNotice(
			printer.Sprintf("Draft Preview"),
			false,
			feature.UserNoticeLink{
				Text: printer.Sprintf("click here to continue editing"),
				Href: f.SelfEditor().GetEditorPath() + "/" + info.EditFilePath(),
			},
		))
	}

	f.ServePreviewEditPage(p, ctx, w, r)
	return
}

func (f *CFeature) RenderFileEditor(w http.ResponseWriter, r *http.Request) {
	var pg feature.Page
	var ctx beContext.Context
	var info *bePkgEditor.File
	var err error
	var eid string
	var handled bool
	if pg, ctx, info, eid, handled = f.PrepareRenderFileEditor(w, r); handled {
		return
	}
	printer := lang.GetPrinterFromRequest(r)

	if info.Tilde != "" {
		if info.Tilde != "draft" || !info.HasDraft {
			// redirect if not ~draft or has no draft
			f.Enjin.ServeRedirect(f.SelfEditor().GetEditorPath()+"/"+info.EditFilePath(), w, r)
			return
		}
		f.RenderFilePreview(w, r)
		return
	}

	var data []byte
	if info.HasDraft {
		if data, err = f.SelfEditor().ReadDraft(info); err != nil {
			log.ErrorRF(r, "error reading draft: %v", err)
			f.Editor.Site().PushErrorNotice(eid, printer.Sprintf(`error reading draft: "%v"`, err), true)
			f.Enjin.ServeRedirect(f.SelfEditor().GetEditorPath()+"/"+info.EditDirectoryPath(), w, r)
			return
		}
	} else if data, err = f.SelfEditor().ReadFile(info); err != nil {
		log.ErrorRF(r, "error reading draft: %v", err)
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf(`error reading file: "%v"`, err), true)
		f.Enjin.ServeRedirect(f.SelfEditor().GetEditorPath()+"/"+info.EditDirectoryPath(), w, r)
		return
	}

	var pm *matter.PageMatter
	if pm, err = matter.ParsePageMatter(info.FSID, info.FilePath(), info.Created, info.Updated, data); err != nil {
		log.ErrorRF(r, "error parsing page-matter: %v", err)
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf(`error parsing page-matter: "%v"`, err), true)
		f.Enjin.ServeRedirect(f.SelfEditor().GetEditorPath()+"/"+info.EditDirectoryPath(), w, r)
		return
	}

	if r, err = f.FinalizeRenderFileEditor(r, eid, pg, pm, ctx, info); err != nil {
		log.ErrorRF(r, "error finalizing edit page: %v", err)
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf(`error finalizing edit page: "%v"`, err), true)
		f.Enjin.ServeRedirect(f.SelfEditor().GetEditorPath()+"/"+info.EditDirectoryPath(), w, r)
		return
	}

	f.SelfEditor().ServePreparedEditPage(pg, ctx, w, r)
}