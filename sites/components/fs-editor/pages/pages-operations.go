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
	"strings"
	"time"

	"github.com/iancoleman/strcase"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/editor"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	bePath "github.com/go-enjin/be/pkg/path"
	"github.com/go-enjin/be/types/page/matter"
	"github.com/go-enjin/golang-org-x-text/message"
)

// TODO: restrict front-matter changes for fields with .LockNonEmpty set to true

func SanitizeKeyName(input string) (cleaned string) {
	cleaned = strcase.ToKebab(input)
	cleaned = strings.Join(strings.Split(cleaned, " "), "-")
	return
}

func IsTmplPage(format string) (yes bool) {
	yes = format == "tmpl" || strings.HasSuffix(format, ".tmpl")
	return
}

func AreVariablesAllowed(key, format string, fields context.Fields) (allowed bool) {
	if IsTmplPage(format) {
		// allow custom fields
		return true
	}
	_, allowed = fields.Lookup(key)
	return
}

func (f *CFeature) NotifyErrors(eid string, printer *message.Printer, errs map[string]error) {
	f.Editor.Site().GetContext(eid).SetSpecific("FieldErrors", errs)
	for _, key := range maps.SortedKeys(errs) {
		ee := errs[key]
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf("%[1]s error: %[2]s", key, ee.Error()), true)
	}
	return
}

func (f *CFeature) OpChangeValidate(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string) (err error) {
	printer := lang.GetPrinterFromRequest(r)
	if info.Locked {
		err = errors.New(printer.Sprintf("Cannot make changes, file is locked by another user"))
	} else if _, present := form["matter"]; !present {
		err = errors.New(printer.Sprintf("incomplete form submitted"))
	} else if _, present := form["body"]; !present {
		err = errors.New(printer.Sprintf("incomplete form submitted"))
	}
	return
}

func (f *CFeature) OpChangeHandler(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string) (redirect string) {
	if stop := f.Emit(feature.PreChangeActionSignal, f.Tag().String(), r, pg, ctx, form, info, eid, &redirect); stop {
		return
	}

	printer := lang.GetPrinterFromRequest(r)

	var err error
	var pm *matter.PageMatter
	if pm, err = f.ReadDraftPage(info); err != nil {
		log.ErrorRF(r, "error encoding form context: %v", err)
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf(`error encoding form context: "%[1]s"`, err.Error()), true)
		redirect = f.SelfEditor().GetEditorPath() + "/" + info.EditFilePath()
		return
	}
	fields := f.MakePageContextFields(r, pm.Matter.String("archetype", ""))

	_, target := feature.ParseEditorOpKey(r.PostFormValue("submit"))
	var changeOp, changeType, changeTarget string
	changeOp, changeTarget = feature.ParseEditorOpKey(target)

	switch changeOp {
	case "~add-new":
		changeType, changeTarget = feature.ParseEditorOpKey(changeTarget)

		addNewKey := "change~add-new~key"
		if changeTarget != "" {
			addNewKey += "." + changeTarget
		} else {
			changeTarget = "matter"
		}
		addNewKeyName := r.PostFormValue(addNewKey)
		addNewKeyName = SanitizeKeyName(addNewKeyName)
		form.Delete("change~add-new~key")
		form.Delete("change~add-new~key." + changeTarget)

		if addNewKeyName == "" {
			f.Editor.Site().PushErrorNotice(eid, printer.Sprintf("kebab-cased key name is required"), true)
		} else {

			kvName := "." + changeTarget + "." + addNewKeyName

			var errMsg string
			format := pg.Format()
			if !AreVariablesAllowed(kvName, pg.Format(), fields) {
				// custom field denied
				form.Delete(kvName)
				format = "." + format
				errMsg = printer.Sprintf(`%[1]s page format does not support variables`, format)
			} else if field, ok := fields[addNewKeyName]; ok {
				f.Editor.Site().PushErrorNotice(eid, printer.Sprintf(
					`cannot add key named "%[1]s": is a reserved reserved name`,
					field.Key,
				), true)
			} else if strings.Contains(addNewKeyName, "~") || strings.Contains(addNewKeyName, ".") {
				errMsg = printer.Sprintf(
					`cannot add key named "%[1]s": contains one or more invalid characters (only lower-case letters, numbers and dashes allowed)`,
					addNewKeyName,
				)
			}
			if errMsg != "" {
				changeType = "nop"
				f.Editor.Site().PushErrorNotice(eid, errMsg, true)
			}

			switch changeType {
			case "value":
				_ = form.SetKV(kvName, "")
			case "list":
				// list is ok because the user can edit the contents
				_ = form.SetKV(kvName, []interface{}{""})
			case "dictionary":
				// TODO: find a better way of starting new map values so that there's no empty keys or other cruft
				//       currently there is an extra blank key involved
				_ = form.SetKV(kvName, map[string]interface{}{"": nil})
			default:
				// nop, discard changes
			}

			form.Delete("submit")
		}

	case "~append":
		kvName := "." + changeTarget
		if v := form.Get(kvName); v != nil {
			if list, ok := form.Get(kvName).([]interface{}); ok {
				_ = form.SetKV(kvName, append(list, ""))
			} else {
				log.ErrorRF(r, "~append %s expected []interface{}, received: %T", kvName, v)
			}
		} else if AreVariablesAllowed(kvName, pg.Format(), fields) {
			_ = form.SetKV(kvName, []interface{}{""})
		} else {
			log.DebugRF(r, "dropping denied field: %q", kvName)
		}

	case "~delete":
		form.Delete("." + changeTarget)

	case "~default":
		kvName := "." + changeTarget
		if field, ok := fields.Lookup(kvName); ok {
			_ = form.SetKV(".matter"+kvName, field.DefaultValue)
		} else {
			_ = form.Delete(".matter" + kvName)
		}

	case "~":
		var changeValue string
		changeTarget, changeValue = feature.ParseEditorOpTargetValue(changeTarget)
		_ = form.SetKV(".matter.~."+changeTarget, changeValue)
	}

	var errs map[string]error
	if pm, redirect, errs = f.ParseFormToDraft(pm, fields, form, info, r); redirect != "" {
		return
	} else if len(errs) > 0 {
		f.NotifyErrors(eid, printer, errs)
	}

	switch changeOp {
	case "~delete":
		_, deleteTarget := feature.ParseEditorOpKey(changeTarget) //<trim .matter prefix
		pm.Matter.Delete("." + deleteTarget)
	}

	if err = f.WriteDraftPage(info, pm); err != nil {
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf("error saving draft page changes: \"%[1]s\"", err.Error()), true)
		return
	}

	f.SelfEditor().UpdateFileInfo(info, r)
	f.SelfEditor().UpdateFileInfoForEditing(info, r)

	ctx.SetSpecific("ShowSidebar", pm.Matter.String(".~.show-sidebar", "true"))
	ctx.SetSpecific("SidebarTab", pm.Matter.String(".~.sidebar-tab", "settings"))
	ctx.SetSpecific("SidebarFieldTab", pm.Matter.String(".~.sidebar-field-tab", "page"))
	ctx.SetSpecific("SidebarFieldCategoryTab", pm.Matter.String(".~.sidebar-field-category-tab", "file"))
	ctx.SetSpecific("Page", pm)
	ctx.SetSpecific("Fields", fields)
	ctx.SetSpecific("FieldErrors", errs)
	ctx.SetSpecific("IsTmplPage", IsTmplPage(bePath.Ext(info.File)))
	editorName := f.SelfEditor().GetEditorName()
	filePath := info.EditFilePath()
	pg.SetTitle(printer.Sprintf("Editing %[1]s: %[2]s", editorName, filePath))

	//redirect = f.SelfEditor().GetEditorPath() + "/" + info.EditFilePath()
	//switch target {
	//case "append":
	//case "expand":
	//case "collapse":
	//default:
	//	redirect += "#" + editor.MakeScrollToKey("."+target)
	//}

	f.Emit(feature.ChangeActionSignal, f.Tag().String(), r, pg, ctx, form, info, eid, &redirect)
	return
}

func (f *CFeature) OpFileCommitValidate(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string) (err error) {
	printer := lang.GetPrinterFromRequest(r)
	if info.Locked {
		err = errors.New(printer.Sprintf("%[1]s is locked by another user, cannot make changes", info.Name))
	} else if _, present := form["matter"]; !present {
		err = errors.New(printer.Sprintf("incomplete form submitted"))
	} else if _, present := form["body"]; !present {
		err = errors.New(printer.Sprintf("incomplete form submitted"))
	}
	return
}

func (f *CFeature) OpFileCommitHandler(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string) (redirect string) {
	printer := lang.GetPrinterFromRequest(r)

	var err error
	var pm *matter.PageMatter
	if pm, err = f.ReadDraftPage(info); err != nil {
		log.ErrorRF(r, "error encoding form context: %v", err)
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf(`error encoding form context: "%[1]s"`, err.Error()), true)
		redirect = f.SelfEditor().GetEditorPath() + "/" + info.EditFilePath()
		return
	}
	fields := f.MakePageContextFields(r, pm.Matter.String("archetype", ""))

	_, target := feature.ParseEditorOpKey(r.PostFormValue("submit"))

	switch target {
	default:
		_ = form.SetKV("."+target, true)
	}

	var errs map[string]error
	if pm, redirect, errs = f.ParseFormToDraft(pm, fields, form, info, r); redirect != "" {
		return
	}

	if err = f.WriteDraftPage(info, pm); err != nil {
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf("error saving draft page changes: \"%[1]s\"", err.Error()), true)
		return
	}

	f.SelfEditor().UpdateFileInfo(info, r)
	f.SelfEditor().UpdateFileInfoForEditing(info, r)

	ctx.SetSpecific("ShowSidebar", pm.Matter.String(".~.show-sidebar", "true"))
	ctx.SetSpecific("SidebarTab", pm.Matter.String(".~.sidebar-tab", "settings"))
	ctx.SetSpecific("SidebarFieldTab", pm.Matter.String(".~.sidebar-field-tab", "page"))
	ctx.SetSpecific("SidebarFieldCategoryTab", pm.Matter.String(".~.sidebar-field-category-tab", "file"))
	ctx.SetSpecific("Page", pm)
	ctx.SetSpecific("Fields", fields)
	ctx.SetSpecific("FieldErrors", errs)
	ctx.SetSpecific("IsTmplPage", IsTmplPage(bePath.Ext(info.File)))
	editorName := f.SelfEditor().GetEditorName()
	filePath := info.EditFilePath()
	pg.SetTitle(printer.Sprintf("Editing %[1]s: %[2]s", editorName, filePath))

	if len(errs) > 0 {
		f.NotifyErrors(eid, printer, errs)
	} else {
		f.Editor.Site().PushInfoNotice(eid, printer.Sprintf("%[1]s draft page changes saved.", info.File), true)
	}

	return
}

func (f *CFeature) OpFilePublishValidate(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string) (err error) {
	printer := lang.GetPrinterFromRequest(r)
	if info.Locked {
		err = errors.New(printer.Sprintf("%[1]s is locked by another user, cannot publish changes", info.Name))
	}
	return
}

func (f *CFeature) OpFilePublishHandler(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string) (redirect string) {
	var err error
	printer := lang.GetPrinterFromRequest(r)

	var pm *matter.PageMatter
	if pm, err = f.ReadDraftPage(info); err != nil {
		log.ErrorRF(r, "error encoding form context: %v", err)
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf(`error encoding form context: "%[1]s"`, err.Error()), true)
		redirect = f.SelfEditor().GetEditorPath() + "/" + info.EditFilePath()
		return
	}
	fields := f.MakePageContextFields(r, pm.Matter.String("archetype", ""))

	_, target := feature.ParseEditorOpKey(r.PostFormValue("submit"))

	if tab, _ := form["tab"].(string); tab == "" {
		ctx.SetSpecific("Tab", "general")
	} else {
		ctx.SetSpecific("Tab", tab)
	}

	switch target {
	default:
		_ = form.SetKV("."+target, true)
	}

	if _, hasMatter := form["matter"]; hasMatter {
		if _, hasBody := form["body"]; hasBody {
			var errs map[string]error
			if pm, redirect, errs = f.ParseFormToDraft(pm, fields, form, info, r); redirect != "" {
				return
			} else if len(errs) > 0 {
				f.NotifyErrors(eid, printer, errs)
				ctx.SetSpecific("ShowSidebar", pm.Matter.String(".~.show-sidebar", "true"))
				ctx.SetSpecific("SidebarTab", pm.Matter.String(".~.sidebar-tab", "settings"))
				ctx.SetSpecific("SidebarFieldTab", pm.Matter.String(".~.sidebar-field-tab", "page"))
				ctx.SetSpecific("SidebarFieldCategoryTab", pm.Matter.String(".~.sidebar-field-category-tab", "file"))
				ctx.SetSpecific("Page", pm)
				ctx.SetSpecific("Fields", fields)
				ctx.SetSpecific("FieldErrors", errs)
				ctx.SetSpecific("IsTmplPage", IsTmplPage(bePath.Ext(info.File)))
				editorName := f.SelfEditor().GetEditorName()
				filePath := info.EditFilePath()
				pg.SetTitle(printer.Sprintf("Editing %[1]s: %[2]s", editorName, filePath))

				return
			} else if err = f.WriteDraftPage(info, pm); err != nil {
				f.Editor.Site().PushErrorNotice(eid, printer.Sprintf("error writing draft page: \"%[1]s\"", err.Error()), true)
				return
			}
		}
	}

	if err = f.PublishDraftPage(info); err != nil {
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf("error publishing draft page: \"%[1]s\"", err.Error()), true)
		return
	} else if err = f.SelfEditor().RemoveDraft(info); err != nil {
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf("error removing final draft page: \"%[1]s\"", err.Error()), true)
		return
	} else if err = f.UnLockEditorFile(info.FSID, info.FilePath()); err != nil {
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf("error unlocking page file: \"%[1]s\"", err.Error()), true)
		return
	}

	f.Editor.Site().PushInfoNotice(eid, printer.Sprintf("%[1]s draft page changes published.", info.File), true)
	redirect = f.SelfEditor().GetEditorPath() + "/" + info.EditDirectoryPath()
	return
}

func (f *CFeature) OpFileDeleteHandler(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string) (redirect string) {

	printer := lang.GetPrinterFromRequest(r)

	if lockedBy, locked := f.IsEditorFileLocked(info.FSID, info.FilePath()); locked && eid != lockedBy {
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf("Cannot delete, file is locked by another user"), true)
		return
	}

	_ = f.UnLockEditorFile(info.FSID, info.FilePath())

	var err error
	op, _ := feature.ParseEditorOpKey(r.PostFormValue("submit"))

	if info.HasDraft {
		if err = f.SelfEditor().RemoveDraft(info); err != nil {
			return
		}
	}

	switch op {
	case editor.DeleteDraftActionKey:
		f.Editor.Site().PushInfoNotice(eid, printer.Sprintf("%[1]s draft deleted.", info.File), true)
		if v, ok := form["return"].(string); ok && v == "directory" {
			redirect = f.SelfEditor().GetEditorPath() + "/" + info.EditDirectoryPath()
		}

	default:

		var pm *matter.PageMatter
		if pm, err = f.ReadPageMatter(info); err != nil {
			log.ErrorRF(r, "error reading page matter: %v - %v", info.Name, err)
			f.Editor.Site().PushErrorNotice(eid, printer.Sprintf("error removing file: %[1]s - %[2]s", info.Name, err.Error()), true)
			return
		}

		if err = f.RemovePage(info, pm); err != nil {
			log.ErrorRF(r, "error removing file: %v - %v", info.Name, err)
			f.Editor.Site().PushErrorNotice(eid, printer.Sprintf("error removing file: %[1]s - %[2]s", info.Name, err.Error()), true)
			return
		}

		f.Editor.Site().PushInfoNotice(eid, printer.Sprintf("%[1]s file deleted.", info.File), true)
		redirect = f.SelfEditor().GetEditorPath() + "/" + info.EditDirectoryPath()
	}

	return
}

func (f *CFeature) OpFileIndexValidate(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string) (err error) {
	//printer := lang.GetPrinterFromRequest(r)
	//if info.Locked {
	//	err = errors.New(printer.Sprintf("%[1]s is locked by another user, cannot republish changes", info.Name))
	//}
	return
}

func (f *CFeature) OpFileIndexHandler(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string) (redirect string) {
	printer := lang.GetPrinterFromRequest(r)

	if lockedBy, locked := f.IsEditorFileLocked(info.FSID, info.FilePath()); locked && eid != lockedBy {
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf("Cannot index, file is locked by another user"), true)
		return
	}

	if _, _, err := f.InfoRenderCheck(info); err != nil {
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf(`cannot re-index broken pages`), true)
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf(`page format render error: %[1]s`, err.Error()), true)
		return
	}

	//if stop := f.Emit(feature.PreRepublishFileSignal, f.Tag().String(), r, pg, ctx, form, info, eid, &redirect); stop {
	//	return
	//}

	f.AddIndexing(info)

	_ = f.UnLockEditorFile(info.FSID, info.FilePath())

	f.Editor.Site().PushInfoNotice(eid, printer.Sprintf(`"%[1]s" has been added to page indexing`, info.Name), true)

	//f.Emit(feature.RepublishFileSignal, f.Tag().String(), r, pg, ctx, form, info, eid, &redirect)
	return
}

func (f *CFeature) OpFileDeIndexValidate(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string) (err error) {
	//printer := lang.GetPrinterFromRequest(r)
	//if info.Locked {
	//	err = errors.New(printer.Sprintf("%[1]s is locked by another user, cannot republish changes", info.Name))
	//}
	return
}

func (f *CFeature) OpFileDeIndexHandler(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string) (redirect string) {
	printer := lang.GetPrinterFromRequest(r)

	if lockedBy, locked := f.IsEditorFileLocked(info.FSID, info.FilePath()); locked && eid != lockedBy {
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf("Cannot remove index, file is locked by another user"), true)
		return
	}

	//if stop := f.Emit(feature.PreRepublishFileSignal, f.Tag().String(), r, pg, ctx, form, info, eid, &redirect); stop {
	//	return
	//}

	f.RemoveIndexing(info)

	_ = f.UnLockEditorFile(info.FSID, info.FilePath())

	f.Editor.Site().PushInfoNotice(eid, printer.Sprintf(`"%[1]s" has been removed from page indexing`, info.Name), true)

	//f.Emit(feature.RepublishFileSignal, f.Tag().String(), r, pg, ctx, form, info, eid, &redirect)
	return
}

func (f *CFeature) OpFileTranslateHandler(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string) (redirect string) {
	if stop := f.Emit(feature.PreTranslateFileActionSignal, f.Tag().String(), r, pg, ctx, form, info, eid, &redirect); stop {
		return
	}

	printer := lang.GetPrinterFromRequest(r)

	if _, _, err := f.InfoRenderCheck(info); err != nil {
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf(`cannot translate broken pages`), true)
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf(`page format render error: %[1]s`, err.Error()), true)
		return
	}

	srcUri, dstUri, dstInfo, srcFS, dstFS, srcMP, dstMP, srcExists, dstExists, stop := f.ParseCopyMoveTranslateForm(r, pg, ctx, form, info, eid, &redirect)
	if stop {
		return
	}

	if !srcExists {
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf(`"%[1]s" not found`, srcUri), true)
		return
	} else if dstExists {
		dst := srcUri
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf(`destination "%[1]s" exists already`, dst), true)
		return
	} else if dstFS == nil || dstMP == nil {
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf(`cannot copy "%[2]s" to "%[1]s": filesystem not found`, dstInfo.FSID, srcUri), true)
		return
	} else if dstMP.RWFS == nil {
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf(`cannot copy "%[2]s" to "%[1]s": filesystem is read-only`, dstInfo.FSID, srcUri), true)
		return
	} else if srcFS == nil || srcMP == nil {
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf(`cannot copy "%[2]s" from "%[1]s": filesystem not found`, info.FSID, srcUri), true)
		return
	}

	var err error
	var srcData []byte
	var srcMatter *matter.PageMatter
	var srcShasum, dstShasum string
	var created, updated time.Time
	if _, srcShasum, created, updated, err = srcMP.ROFS.FileStats(info.FilePath()); err != nil {
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf(`error reading "%[1]s" file stats: %[2]s`, srcUri, err.Error()), true)
		return
	} else if srcMatter, err = srcMP.ROFS.ReadPageMatter(info.FilePath()); err != nil {
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf(`error reading "%[1]s" page matter: %[2]s`, srcUri, err.Error()), true)
		return
	}
	srcMatter.Matter.SetSpecific("translates", info.Url())
	if srcData, err = srcMatter.Bytes(); err != nil {
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf(`error preparing "%[1]s" page matter: %[2]s`, srcUri, err.Error()), true)
		return
	} else if err = dstMP.RWFS.WriteFile(dstInfo.FilePath(), srcData, 0664); err != nil {
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf(`error writing "%[1]s": %[2]s`, dstUri, err.Error()), true)
		return
	} else if err = dstMP.RWFS.ChangeTimes(dstInfo.FilePath(), created, updated); err != nil {
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf(`error preserving file timestamps "%[1]s": %[2]s`, dstUri, err.Error()), true)
		return
	} else if dstShasum, err = dstMP.RWFS.Shasum(dstInfo.FilePath()); err != nil {
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf(`error reading "%[1]s" shasum: %[2]s`, dstUri, err.Error()), true)
		return
	} else if srcShasum == dstShasum {
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf(`source and destination file shasums are the same`), true)
		_ = dstMP.RWFS.Remove(dstInfo.FilePath())
		return
	}

	f.AddIndexing(dstInfo)

	if v, _ := form["return"].(string); v == "directory" {
		redirect = f.SelfEditor().GetEditorPath() + "/" + dstInfo.EditDirectoryPath()
	} else {
		redirect = f.SelfEditor().GetEditorPath() + "/" + dstInfo.EditFilePath()
	}

	f.Editor.Site().PushWarnNotice(eid, printer.Sprintf(`"%[1]s" %[2]s page started, please translate and publish`, info.Name, dstInfo.Locale.String()), true)

	f.Emit(feature.TranslateFileActionSignal, f.Tag().String(), r, pg, ctx, form, info, eid, &redirect)
	return
}

func (f *CFeature) OpPageCreateValidate(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string) (err error) {
	//printer := lang.GetPrinterFromRequest(r)
	//if info.Locked {
	//	err = errors.New(printer.Sprintf("%[1]s is locked by another user, cannot republish changes", info.Name))
	//}
	return
}

func (f *CFeature) OpPageCreateHandler(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string) (redirect string) {
	printer := lang.GetPrinterFromRequest(r)
	dstUri, dstArchetype, dstInfo, dstFS, dstMP, dstExists, stop := f.ParseCreatePageForm(r, pg, ctx, form, info, eid, &redirect)
	if stop {
		return
	}

	if dstExists {
		dst := dstUri
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf(`destination "%[1]s" exists already`, dst), true)
		return
	} else if dstFS == nil || dstMP == nil {
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf(`cannot create "%[2]s" on "%[1]s": filesystem not found`, dstInfo.FSID, dstInfo.File), true)
		return
	} else if dstMP.RWFS == nil {
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf(`cannot create "%[2]s" on "%[1]s": filesystem is read-only`, dstInfo.FSID, dstInfo.File), true)
		return
	}

	var err error
	var data []byte
	var format string

	if dstArchetype != "" {
		t := f.Enjin.MustGetTheme()
		if format, data, err = t.MakeArchetype(f.Enjin, dstArchetype); err != nil {
			f.Editor.Site().PushErrorNotice(eid, printer.Sprintf(`error making archetype "%[1]s": %[2]s`, dstArchetype, err.Error()), true)
			return
		}
	}

	realName := bePath.Base(dstInfo.Name) + "." + format
	dstInfo.Name = realName
	dstInfo.File = strings.Replace(dstInfo.File, dstInfo.Name, realName, 1)

	if err = dstMP.RWFS.WriteFile(dstInfo.FilePath(), data, 0664); err != nil {
		f.Editor.Site().PushErrorNotice(eid, printer.Sprintf(`error writing "%[1]s": %[2]s`, dstUri, err.Error()), true)
		return
	}

	f.Editor.Site().PushInfoNotice(eid, printer.Sprintf(`create new page "%[1]s"`, dstUri), true)
	if v, _ := form["return"].(string); v == "directory" {
		redirect = f.SelfEditor().GetEditorPath() + "/" + dstInfo.EditDirectoryPath()
	} else {
		redirect = f.SelfEditor().GetEditorPath() + "/" + dstInfo.EditFilePath()
	}
	return
}