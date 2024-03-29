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
	"bytes"
	htmlTemplate "html/template"
	"net/http"

	"github.com/iancoleman/strcase"

	clPath "github.com/go-corelibs/path"
	"github.com/go-corelibs/x-text/message"
	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/editor"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/menu"
)

func (f *CFeature) ServePreviewEditPage(pg feature.Page, ctx context.Context, w http.ResponseWriter, r *http.Request) {
	handler := f.Enjin.GetServePagesHandler()
	if err := handler.ServePage(pg, f.Enjin.MustGetTheme(), ctx, w, r); err != nil {
		log.ErrorRF(r, "error serving %v editor preview page: %v", f.Tag(), err)
		f.Enjin.ServeInternalServerError(w, r)
	}
}

func (f *CFeature) RenderFilePreview(w http.ResponseWriter, r *http.Request) {
	var pg feature.Page
	var ctx context.Context
	var info *editor.File
	var handled bool
	var eid string
	if pg, ctx, info, eid, handled = f.PrepareRenderFileEditor(w, r); handled {
		return
	}
	printer := message.GetPrinter(r)

	var err error
	var list menu.EditMenu
	var data []byte
	if info.HasDraft {
		if data, err = f.SelfEditor().ReadDraft(info); err != nil {
			log.ErrorRF(r, "error reading draft: %v", err)
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error reading draft: "%v"`, err))
			f.Enjin.ServeRedirect(f.SelfEditor().GetEditorPath()+"/"+info.EditDirectoryPath(), w, r)
			return
		}
	} else if data, err = f.SelfEditor().ReadFile(info); err != nil {
		log.ErrorRF(r, "error reading draft: %v", err)
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error reading file: "%v"`, err))
		f.Enjin.ServeRedirect(f.SelfEditor().GetEditorPath()+"/"+info.EditDirectoryPath(), w, r)
		return
	}

	if list, err = menu.NewEditMenuFromJson(data); err != nil {
		log.ErrorRF(r, "error reading draft: %v", err)
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error encoding menu: "%v"`, err))
		f.Enjin.ServeRedirect(f.SelfEditor().GetEditorPath()+"/"+info.EditDirectoryPath(), w, r)
		return
	}

	siteMenu := context.New()
	basename := clPath.Base(info.File)
	menuName := strcase.ToCamel(basename)
	siteMenu[menuName] = list.AsMenu()

	ctx.SetSpecific("TopMenu", list)
	ctx.SetSpecific("SiteMenu", siteMenu)
	ctx.SetSpecific("SelfEditorPath", f.SelfEditor().GetEditorPath())

	if info.Locked {
		r = feature.AddWarnNotice(r, false,
			printer.Sprintf("Menu Preview"),
			feature.UserNoticeLink{
				Text: printer.Sprintf("click here to return to the file-browser"),
				Href: f.SelfEditor().GetEditorPath() + "/" + info.EditDirectoryPath(),
			},
		)
	} else {
		r = feature.AddWarnNotice(r, false,
			printer.Sprintf("Menu Preview"),
			feature.UserNoticeLink{
				Text: printer.Sprintf("click here to continue editing"),
				Href: f.SelfEditor().GetEditorPath() + "/" + info.EditFilePath(),
			},
		)
	}

	pg.SetType("page")
	pg.SetLayout("defaults")
	pg.SetFormat("tmpl")

	var content string
	buf := bytes.Buffer{}
	var tt *htmlTemplate.Template
	if tt, err = f.Editor.SiteFeatureTheme().NewHtmlTemplate(f.Enjin, "lorem-ipsum.tmpl", f.Enjin.Context(r)); err != nil {
		content = "<p>" + err.Error() + "</p>"
	} else if tt, err = tt.Parse(`{{ template "partials/fs-editor/lorem-ipsum.tmpl" . }}`); err != nil {
		content = "<p>" + err.Error() + "</p>"
	} else if err = tt.Execute(&buf, f.Enjin.Context(r)); err != nil {
		content = "<p>" + err.Error() + "</p>"
	} else {
		content = buf.String()
	}
	pg.SetContent(content)

	f.ServePreviewEditPage(pg, ctx, w, r)
	return
}

func (f *CFeature) RenderFileEditor(w http.ResponseWriter, r *http.Request) {
	var pg feature.Page
	var ctx context.Context
	var info *editor.File
	var err error
	var eid string
	var handled bool
	if pg, ctx, info, eid, handled = f.PrepareRenderFileEditor(w, r); handled {
		return
	}
	printer := message.GetPrinter(r)

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
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error reading draft: "%v"`, err))
			f.Enjin.ServeRedirect(f.SelfEditor().GetEditorPath()+"/"+info.EditDirectoryPath(), w, r)
			return
		}
	} else if data, err = f.SelfEditor().ReadFile(info); err != nil {
		log.ErrorRF(r, "error reading draft: %v", err)
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error reading file: "%v"`, err))
		f.Enjin.ServeRedirect(f.SelfEditor().GetEditorPath()+"/"+info.EditDirectoryPath(), w, r)
		return
	}

	var list menu.EditMenu
	if list, err = menu.NewEditMenuFromJson(data); err != nil {
		log.ErrorRF(r, "error reading draft: %v", err)
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error encoding menu: "%v"`, err))
		f.Enjin.ServeRedirect(f.SelfEditor().GetEditorPath()+"/"+info.EditDirectoryPath(), w, r)
		return
	}

	ctx.SetSpecific("TopMenu", list)
	ctx.SetSpecific("ShowSidebar", "true")
	ctx.SetSpecific("SidebarTab", "details")
	ctx.SetSpecific("SelfEditorPath", f.SelfEditor().GetEditorPath())
	ctx.SetSpecific("TranslatedLocales", f.GetTranslatedLocales(info))
	ctx.SetSpecific("UntranslatedLocales", f.GetUntranslatedLocales(info))
	r = feature.AddUserNotices(r, f.Editor.Site().PullNotices(eid)...)
	pg.SetTitle(printer.Sprintf("Edit: %[1]s", info.Name))
	f.SelfEditor().ServePreparedEditPage(pg, ctx, w, r)
}
