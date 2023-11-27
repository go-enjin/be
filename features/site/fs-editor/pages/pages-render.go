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
	errors2 "errors"
	"net/http"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/editor"
	"github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/path"
	"github.com/go-enjin/be/types/page"
	"github.com/go-enjin/be/types/page/matter"
)

func (f *CFeature) FinalizeRenderFileEditor(r *http.Request, eid string, pg feature.Page, pm *matter.PageMatter, ctx context.Context, info *editor.File) (modified *http.Request, err error) {
	printer := lang.GetPrinterFromRequest(r)

	var p feature.Page
	if p, err = page.NewFromPageMatter(pm.Copy(), f.Enjin.MustGetTheme(), f.Enjin.Context()); err != nil {
		return
	}

	if ee := f.PageRenderCheck(p); ee != nil {
		info.Actions = info.Actions.Prune(
			editor.PreviewDraftActionKey,
			editor.PublishActionKey,
			editor.DeIndexPageActionKey,
			editor.TranslateActionKey,
		)
		var enjErr *errors.EnjinError
		if errors2.As(ee, &enjErr) {
			f.Editor.Site().PushErrorNotice(eid, false, printer.Sprintf(`page format error: %[1]s - %[2]s`, enjErr.Title, forms.StrictSanitize(enjErr.Summary)))
			info.Actions = append(info.Actions, editor.MakeViewErrorAction(printer))
		} else {
			f.Editor.Site().PushErrorNotice(eid, false, printer.Sprintf(`page render error: %[1]s`, forms.StrictSanitize(ee.Error())))
		}
	}

	ctx.SetSpecific("ShowSidebar", pm.Matter.String(".~.show-sidebar", "true"))
	ctx.SetSpecific("SidebarTab", pm.Matter.String(".~.sidebar-tab", "details"))
	ctx.SetSpecific("SidebarFieldTab", pm.Matter.String(".~.sidebar-field-tab", "page"))
	ctx.SetSpecific("SidebarFieldCategoryTab", pm.Matter.String(".~.sidebar-field-category-tab", "file"))
	ctx.SetSpecific("Page", pm)
	ctx.SetSpecific("CPage", p)
	//var archetype string
	//if archetype = p.Archetype(); archetype != "" {
	//	archetype += "." + p.Format()
	//}
	//ctx.SetSpecific("Fields", f.MakePageContextFields(r, archetype))
	ctx.SetSpecific("Fields", f.MakePageContextFields(r, p.Archetype()))
	ctx.SetSpecific("IsTmplPage", IsTmplPage(path.Ext(info.File)))
	ctx.SetSpecific("SelfEditorPath", f.SelfEditor().GetEditorPath())
	ctx.SetSpecific("TranslatedLocales", f.GetTranslatedLocales(info))
	ctx.SetSpecific("UntranslatedLocales", f.GetUntranslatedLocales(info))
	if errs := f.Editor.Site().GetContext(eid).Get("FieldErrors"); errs != nil {
		f.Editor.Site().DeleteContextKeys(eid, "FieldErrors")
		ctx.SetSpecific("FieldErrors", errs)
	}
	pg.SetTitle(printer.Sprintf("Edit: %[1]s", info.Name))
	modified = feature.AddUserNotices(r, f.Editor.Site().PullNotices(eid)...)
	return
}
