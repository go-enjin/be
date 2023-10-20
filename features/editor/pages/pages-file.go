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
	"errors"
	"net/http"

	beContext "github.com/go-enjin/be/pkg/context"
	bePkgEditor "github.com/go-enjin/be/pkg/editor"
	beErrors "github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	bePath "github.com/go-enjin/be/pkg/path"
	"github.com/go-enjin/be/types/page"
	"github.com/go-enjin/be/types/page/matter"
)

func (f *CFeature) ServePreparedEditPage(pg feature.Page, ctx beContext.Context, w http.ResponseWriter, r *http.Request) {
	handler := f.Enjin.GetServePagesHandler()
	t := f.Enjin.MustGetTheme()
	ctx.SetSpecific("PageArchetypes", t.ListArchetypes())
	if err := handler.ServePage(pg, f.Editor.EditorTheme(), ctx, w, r); err != nil {
		log.ErrorRF(r, "error serving %v editor generic page: %v", f.Tag(), err)
		f.Enjin.ServeInternalServerError(w, r)
	}
}

func (f *CFeature) ServePreviewEditPage(pg feature.Page, ctx beContext.Context, w http.ResponseWriter, r *http.Request) {
	printer := lang.GetPrinterFromRequest(r)

	if ee := f.PageRenderCheck(pg); ee != nil {
		var contents string
		var enjErr *beErrors.EnjinError
		if errors.As(ee, &enjErr) {
			contents = string(enjErr.Html())
		} else {
			contents = "<p>" + printer.Sprintf(`(this page failed to render)`) + "</p>"
		}
		pg.SetContent(contents)
		pg.SetArchetype("")
		pg.SetFormat("html")
	}

	handler := f.Enjin.GetServePagesHandler()
	if err := handler.ServePage(pg, f.Enjin.MustGetTheme(), ctx, w, r); err != nil {
		log.ErrorRF(r, "error serving %v editor preview page: %v", f.Tag(), err)
		f.Enjin.ServeInternalServerError(w, r)
	}
}

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
		f.Editor.PushErrorNotice(eid, printer.Sprintf(`error reading draft page matter: "%[1]s"`, err.Error()), false)
		r = feature.AddUserNotices(r, f.Editor.PullNotices(eid)...)
		f.Enjin.ServeRedirect(f.SelfEditor().GetEditorPath()+"/"+info.EditFilePath(), w, r)
		return
	}

	var p feature.Page
	if p, err = page.NewFromPageMatter(pm, f.Editor.EditorTheme(), f.Enjin.Context()); err != nil {
		f.Editor.PushErrorNotice(eid, printer.Sprintf(`error preparing draft page preview: "%[1]s"`, err.Error()), false)
		r = feature.AddUserNotices(r, f.Editor.PullNotices(eid)...)
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
			f.Editor.PushErrorNotice(eid, printer.Sprintf(`error reading draft: "%v"`, err), true)
			f.Enjin.ServeRedirect(f.SelfEditor().GetEditorPath()+"/"+info.EditDirectoryPath(), w, r)
			return
		}
	} else if data, err = f.SelfEditor().ReadFile(info); err != nil {
		log.ErrorRF(r, "error reading draft: %v", err)
		f.Editor.PushErrorNotice(eid, printer.Sprintf(`error reading file: "%v"`, err), true)
		f.Enjin.ServeRedirect(f.SelfEditor().GetEditorPath()+"/"+info.EditDirectoryPath(), w, r)
		return
	}

	var pm *matter.PageMatter
	if pm, err = matter.ParsePageMatter(info.FSID, info.FilePath(), info.Created, info.Updated, data); err != nil {
		log.ErrorRF(r, "error parsing page-matter: %v", err)
		f.Editor.PushErrorNotice(eid, printer.Sprintf(`error parsing page-matter: "%v"`, err), true)
		f.Enjin.ServeRedirect(f.SelfEditor().GetEditorPath()+"/"+info.EditDirectoryPath(), w, r)
		return
	}

	if r, err = f.FinalizeRenderFileEditor(r, eid, pg, pm, ctx, info); err != nil {
		log.ErrorRF(r, "error finalizing edit page: %v", err)
		f.Editor.PushErrorNotice(eid, printer.Sprintf(`error finalizing edit page: "%v"`, err), true)
		f.Enjin.ServeRedirect(f.SelfEditor().GetEditorPath()+"/"+info.EditDirectoryPath(), w, r)
		return
	}

	f.SelfEditor().ServePreparedEditPage(pg, ctx, w, r)
}

func (f *CFeature) MakePageArchetypeContextFields(name string) (fields beContext.Fields) {

	tc := f.Enjin.MustGetTheme().GetConfig()
	fields = beContext.Fields{}
	basename := bePath.Base(name)

	if found, ok := tc.Supports.Archetypes[basename]; ok {
		// general fields for any format of archetype
		for k, v := range found {
			fields[k] = v
		}
	}

	if basename != name {
		if found, ok := tc.Supports.Archetypes[name]; ok {
			// fields for a specific archetype, clobbering generals
			for k, v := range found {
				fields[k] = v
			}
		}
	}

	return
}

func (f *CFeature) MakePageContextFields(r *http.Request, archetype string) (fields beContext.Fields) {
	fields = f.Enjin.MakePageContextFields(r)
	for k, v := range f.MakePageArchetypeContextFields(archetype) {
		if _, present := fields[k]; !present {
			// no clobbering allowed here
			fields[k] = v
		}
	}
	return
}

func (f *CFeature) PageRenderCheck(p feature.Page) (err error) {
	renderer := f.Enjin.GetThemeRenderer(p.Context())
	_, _, err = renderer.PrepareRenderPage(f.Enjin.MustGetTheme(), f.Enjin.Context(), p)
	return
}

func (f *CFeature) InfoRenderCheck(info *bePkgEditor.File) (p feature.Page, pm *matter.PageMatter, err error) {
	if info.HasDraft {
		if pm, err = f.ReadDraftMatter(info); err != nil {
			return
		}
	} else {
		if pm, err = f.ReadPageMatter(info); err != nil {
			return
		}
	}
	if p, err = page.NewFromPageMatter(pm.Copy(), f.Enjin.MustGetTheme(), f.Enjin.Context()); err != nil {
		return
	}
	err = f.PageRenderCheck(p)
	return
}

func (f *CFeature) FinalizeRenderFileEditor(r *http.Request, eid string, pg feature.Page, pm *matter.PageMatter, ctx beContext.Context, info *bePkgEditor.File) (modified *http.Request, err error) {
	printer := lang.GetPrinterFromRequest(r)

	var p feature.Page
	if p, err = page.NewFromPageMatter(pm.Copy(), f.Enjin.MustGetTheme(), f.Enjin.Context()); err != nil {
		return
	}

	if ee := f.PageRenderCheck(p); ee != nil {
		info.Actions = info.Actions.Prune(
			bePkgEditor.PreviewDraftActionKey,
			bePkgEditor.PublishActionKey,
			bePkgEditor.DeIndexPageActionKey,
			bePkgEditor.TranslateActionKey,
		)
		var enjErr *beErrors.EnjinError
		if errors.As(ee, &enjErr) {
			f.Editor.PushErrorNotice(eid, printer.Sprintf(`page format error: %[1]s - %[2]s`, enjErr.Title, forms.StrictSanitize(enjErr.Summary)), false)
			info.Actions = append(info.Actions, bePkgEditor.MakeViewErrorAction(printer))
		} else {
			f.Editor.PushErrorNotice(eid, printer.Sprintf(`page render error: %[1]s`, forms.StrictSanitize(ee.Error())), false)
		}
	}

	ctx.SetSpecific("ShowSidebar", pm.Matter.String(".~.show-sidebar", "true"))
	ctx.SetSpecific("SidebarTab", pm.Matter.String(".~.sidebar-tab", "details"))
	ctx.SetSpecific("SidebarFieldTab", pm.Matter.String(".~.sidebar-field-tab", "page"))
	ctx.SetSpecific("SidebarFieldCategoryTab", pm.Matter.String(".~.sidebar-field-category-tab", "file"))
	ctx.SetSpecific("Page", pm)
	ctx.SetSpecific("CPage", p)
	var archetype string
	if archetype = p.Archetype(); archetype != "" {
		archetype += "." + p.Format()
	}
	ctx.SetSpecific("Fields", f.MakePageContextFields(r, archetype))
	ctx.SetSpecific("IsTmplPage", IsTmplPage(bePath.Ext(info.File)))
	ctx.SetSpecific("SelfEditorPath", f.SelfEditor().GetEditorPath())
	ctx.SetSpecific("TranslatedLocales", f.GetTranslatedLocales(info))
	ctx.SetSpecific("UntranslatedLocales", f.GetUntranslatedLocales(info))
	pg.SetTitle(printer.Sprintf("Edit: %[1]s", info.Name))
	modified = feature.AddUserNotices(r, f.Editor.PullNotices(eid)...)
	return
}