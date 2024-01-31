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
	"errors"
	"net/http"
	"strconv"
	"strings"

	cllang "github.com/go-corelibs/lang"
	"github.com/go-corelibs/slices"
	"github.com/go-corelibs/x-text/language"
	"github.com/go-corelibs/x-text/message"
	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/editor"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
)

func (f *CFeature) OpRetakeValidate(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string) (err error) {
	//printer := message.GetPrinter(r)
	//if info.Locked {
	//	err = errors.New(printer.Sprintf("Cannot take over editing, locale is locked by another user"))
	//}
	return
}

func (f *CFeature) OpRetakeHandler(r *http.Request, pg feature.Page, ctx context.Context, form context.Context, info *editor.File, eid string) (redirect string) {
	//log.DebugRF(r, "retake editing: info=%#+v; form=%#+v", info, form)
	printer := message.GetPrinter(r)
	if err := f.LockLocale(eid, info.FSID, info.Code); err != nil {
		log.ErrorRF(r, "error locking %v locale for editing by others: %v", info.FSID+"/"+info.Code, err)
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error taking over editing: %[1]s`, err.Error()))
		return
	}
	//if info.Code != "" {
	//	redirect = f.GetEditorPath() + "/" + info.FSID
	//} else {
	//	redirect = f.GetEditorPath()
	//}
	return
}

func (f *CFeature) OpUnlockValidate(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string) (err error) {
	printer := message.GetPrinter(r)
	if info.Locked {
		err = errors.New(printer.Sprintf("Cannot unlock, locale is locked by another user"))
	}
	return
}

func (f *CFeature) OpUnlockHandler(r *http.Request, pg feature.Page, ctx context.Context, form context.Context, info *editor.File, eid string) (redirect string) {
	//log.DebugRF(r, "unlock editing: info=%#+v; form=%#+v", info, form)
	printer := message.GetPrinter(r)
	if err := f.UnlockLocales(info.FSID, info.Code); err != nil {
		log.ErrorRF(r, "error unlocking %v locale for editing by others: %v", info.FSID+"/"+info.Code, err)
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error unlocking locale for editing by others: %[1]s`, err.Error()))
		return
	}
	if info.Code != "" {
		redirect = f.GetEditorPath() + "/" + info.FSID
	} else {
		redirect = f.GetEditorPath()
	}
	return
}

func (f *CFeature) OpCancelValidate(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string) (err error) {
	printer := message.GetPrinter(r)
	if info.Locked {
		err = errors.New(printer.Sprintf("Cannot cancel, locale is locked by another user"))
	}
	return
}

func (f *CFeature) OpCancelHandler(r *http.Request, pg feature.Page, ctx context.Context, form context.Context, info *editor.File, eid string) (redirect string) {
	//log.DebugRF(r, "cancel editing: info=%#+v; form=%#+v", info, form)
	if err := f.UnlockLocales(info.FSID, info.Code); err != nil {
		log.ErrorRF(r, "error unlocking %v locale for editing by others: %v", info.FSID+"/"+info.Code, err)
	}
	if info.Code != "" {
		redirect = f.GetEditorPath() + "/" + info.FSID
	} else {
		redirect = f.GetEditorPath()
	}
	return
}

func (f *CFeature) OpCommitValidate(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string) (err error) {
	printer := message.GetPrinter(r)
	if info.Locked {
		err = errors.New(printer.Sprintf("Cannot save changes, locale is locked by another user"))
	}
	return
}

func (f *CFeature) OpCommitHandler(r *http.Request, pg feature.Page, ctx context.Context, form context.Context, info *editor.File, eid string) (redirect string) {
	//log.DebugRF(r, "commit editing: info=%#+v; form=%#+v", info, form)
	translations, _ := form["tx"].(map[string]interface{})
	printer := message.GetPrinter(r)

	var err error
	var ld *LocaleData
	mountPoints := f.FindMountPoints(info.FSID, info.Code)
	if ld, err = f.ReadDraftLocales(info.FSID, info.Code, mountPoints, false); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf("error reading locales data: %[1]s", err.Error()))
		return
	}

	if err = f.performDraftChanges(r, translations, ld); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error performing draft changes: %[1]s`, err.Error()))
		return
	}

	if err = f.WriteDraftLocales(ld, mountPoints); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error writing draft changes: %[1]s`, err.Error()))
		return
	}

	f.Editor.Site().PushInfoNotice(eid, true, printer.Sprintf(`draft locale changes saved`))
	redirect = f.makePaginationRedirect(info, form)
	return
}

func (f *CFeature) OpPublishValidate(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string) (err error) {
	printer := message.GetPrinter(r)
	if info.Locked {
		err = errors.New(printer.Sprintf("Cannot publish, locale is locked by another user"))
	}
	return
}

func (f *CFeature) OpPublishHandler(r *http.Request, pg feature.Page, ctx context.Context, form context.Context, info *editor.File, eid string) (redirect string) {
	//log.DebugRF(r, "publish editing: info=%#+v; form=%#+v", info, form)

	translations, _ := form["tx"].(map[string]interface{})
	printer := message.GetPrinter(r)

	var err error
	var ld *LocaleData
	mountPoints := f.FindMountPoints(info.FSID, info.Code)
	if ld, err = f.ReadDraftLocales(info.FSID, info.Code, mountPoints, false); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf("error reading locales data: %[1]s", err.Error()))
		return
	}

	if err = f.performDraftChanges(r, translations, ld); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error performing draft changes: %[1]s`, err.Error()))
		return
	}

	if err = f.WriteLocales(ld, mountPoints); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error writing draft changes: %[1]s`, err.Error()))
		return
	}

	f.DeleteDraftLocales(mountPoints)

	if !f.Enjin.HotReloading() {
		f.Enjin.ReloadLocales()
	}

	if err = f.UnlockLocales(info.FSID, info.Code); err != nil {
		log.ErrorRF(r, "error unlocking %v locale for editing by others: %v", info.FSID+"/"+info.Code, err)
	}

	f.Editor.Site().PushInfoNotice(eid, true, printer.Sprintf(`draft locale changes published`))

	if info.Code != "" {
		redirect = f.GetEditorPath() + "/" + info.FSID
	} else {
		redirect = f.GetEditorPath()
	}
	return
}

func (f *CFeature) OpDeleteValidate(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string) (err error) {
	printer := message.GetPrinter(r)
	if info.Locked {
		err = errors.New(printer.Sprintf("Cannot delete draft changes, locale is locked by another user"))
	}
	return
}

func (f *CFeature) OpDeleteHandler(r *http.Request, pg feature.Page, ctx context.Context, form context.Context, info *editor.File, eid string) (redirect string) {
	//log.DebugRF(r, "delete editing: info=%#+v; form=%#+v", info, form)
	mountPoints := f.FindMountPoints(info.FSID, info.Code)
	f.DeleteDraftLocales(mountPoints)
	if err := f.UnlockLocales(info.FSID, info.Code); err != nil {
		log.ErrorRF(r, "error unlocking %v locale for editing by others: %v", info.FSID+"/"+info.Code, err)
	}
	if info.Code != "" {
		redirect = f.GetEditorPath() + "/" + info.FSID
	} else {
		redirect = f.GetEditorPath()
	}
	return
}

func (f *CFeature) OpChangeValidate(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string) (err error) {
	printer := message.GetPrinter(r)
	if info.Locked {
		err = errors.New(printer.Sprintf("Cannot change, locale is locked by another user"))
	}
	return
}

func (f *CFeature) OpChangeHandler(r *http.Request, pg feature.Page, ctx context.Context, form context.Context, info *editor.File, eid string) (redirect string) {
	//log.DebugRF(r, "change editing: info=%#+v; form=%#+v", info, form)
	//if err := f.UnlockLocales(info.FSID, info.Code); err != nil {
	//	log.ErrorRF(r, "error unlocking %v locale for editing by others: %v", info.FSID+"/"+info.Code, err)
	//}

	var err error
	var ld *LocaleData

	translations, _ := form["tx"].(map[string]interface{})
	printer := message.GetPrinter(r)
	mountPoints := f.FindMountPoints(info.FSID, info.Code)

	if ld, err = f.ReadDraftLocales(info.FSID, info.Code, mountPoints, false); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf("error reading locales data: %[1]s", err.Error()))
		return
	}

	if err = f.performDraftChanges(r, translations, ld); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error performing draft changes: %[1]s`, err.Error()))
		return
	}

	changed := true
	submit := r.PostFormValue("submit")
	_, target := feature.ParseEditorOpKey(submit)
	changeOp, changeTarget := feature.ParseEditorOpKey(target)
	log.DebugRF(r, `target=%v; changeOp=%v; changeTarget=%v`, target, changeOp, changeTarget)

	switch changeOp {

	case "copy-translation":

		var shasum, dstPath string
		if parts := strings.Split(changeTarget, "."); len(parts) != 1 {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`incomplete form submission`))
			return
		} else if shasum = parts[0]; len(shasum) != 10 {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`incomplete form submission`))
			return
		} else if dstPath, _ = form[submit+"~dst-locale-system"].(string); dstPath == "" {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`incomplete form submission`))
			return
		}
		var dstFsid, dstCode string
		if parts := strings.Split(dstPath, "/"); len(parts) == 1 {
			dstFsid = parts[0]
		} else if len(parts) == 2 {
			dstFsid, dstCode = parts[0], parts[1]
		} else {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`incorrect form submission`))
			return
		}

		dstMountPoints := f.FindMountPoints(dstFsid, dstCode)

		var dstLd *LocaleData
		if dstLd, err = f.ReadDraftLocales(info.FSID, info.Code, dstMountPoints, false); err != nil {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf("destination not found, cannot copy"))
			return
		}

		var ok bool
		var msgs map[language.Tag]*LocaleMessage
		if msgs, ok = ld.Data[shasum]; !ok {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`message not found, cannot copy`))
			return
		}

		dstLd.Data[shasum] = msgs
		if !slices.Within(shasum, dstLd.Order) {
			dstLd.Order = append(dstLd.Order, shasum)
		}
		if err = f.WriteDraftLocales(dstLd, dstMountPoints); err != nil {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error writing draft changes: %[1]s`, err.Error()))
			return
		}

		changed = false
		f.Editor.Site().PushInfoNotice(eid, true, printer.Sprintf(`translation copied to: %[1]s`, dstPath))
		redirect = f.GetEditorPath() + "/" + dstPath

	case "add-translation":

		var key, comment string
		if key, _ = form["add-translation.key"].(string); key == "" {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`translation key is required to create a new entry`))
			return
		} else if comment, _ = form["add-translation.comment"].(string); comment == "" {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`translation comment is required to create a new entry`))
			return
		}

		msg := ParseNewMessage(key, comment)
		locales := f.Enjin.SiteLocales()
		if _, present := ld.Data[msg.Shasum]; present {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`translation key exists already`))
			return
		}
		ld.Data[msg.Shasum] = map[language.Tag]*LocaleMessage{}
		for _, tag := range locales {
			ld.Data[msg.Shasum][tag] = msg.Copy()
		}
		ld.Order = append(ld.Order, msg.Shasum)

	case "delete-translation":

		var shasum string
		if parts := strings.Split(changeTarget, "."); len(parts) != 1 {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`incomplete form submission`))
			return
		} else if shasum = parts[0]; len(shasum) != 10 {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`incomplete form submission`))
			return
		}

		if _, present := ld.Data[shasum]; !present {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`translation not found, nothing to delete`))
			return
		}

		delete(ld.Data, shasum)
		ld.Order = slices.Remove(ld.Order, slices.IndexOf(ld.Order, shasum))

	case "add-translation-case":

		var tag language.Tag
		var shasum, key string
		if parts := strings.Split(changeTarget, "."); len(parts) != 2 {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`incomplete form submission`))
			return
		} else if tag, err = language.Parse(parts[0]); err != nil {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`incorrect form submission`))
			return
		} else if shasum = parts[1]; len(shasum) != 10 {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`incomplete form submission`))
			return
		} else if key, _ = form["add-translation-case."+tag.String()+"."+shasum].(string); key == "" {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`incomplete form submission`))
			return
		} else if key = cllang.ParsePluralCaseKey(key); key == "" {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`invalid plural translation case key`))
			return
		}

		if msg, ok := ld.Data[shasum][tag]; ok {
			if msg.Translation.Select != nil {
				if _, present := msg.Translation.Select.Cases[key]; present {
					f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`translation case "%[1]s" exists already`, key))
					return
				}
				msg.Translation.Select.Cases[key] = msg.Key
			}
		}

	case "delete-translation-case":

		var idx int
		var tag language.Tag
		var shasum, key string
		if parts := strings.Split(changeTarget, "."); len(parts) != 3 {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`incomplete form submission`))
			return
		} else if tag, err = language.Parse(parts[0]); err != nil {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`incorrect form submission`))
			return
		} else if shasum = parts[1]; len(shasum) != 10 {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`incomplete form submission`))
			return
		} else if idx, err = strconv.Atoi(parts[2]); err != nil {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`incorrect form submission`))
			return
		} else if key, _ = form["delete-translation-case."+tag.String()+"."+shasum+"."+strconv.Itoa(idx)].(string); key == "" {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`incomplete form submission`))
			return
		} else if key = cllang.ParsePluralCaseKey(key); key == "" {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`invalid plural translation case key`))
			return
		} else if key == "other" {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`"other" translation case is required, cannot delete`))
			return
		}

		if msg, ok := ld.Data[shasum][tag]; ok {
			if msg.Translation.Select != nil {
				if _, present := msg.Translation.Select.Cases[key]; present {
					delete(msg.Translation.Select.Cases, key)
					log.DebugRF(r, "deleted translation case: %q - %q", msg.ID, key)
				}
			}
		}

	case "pluralize-translation":

		var shasum string
		if parts := strings.Split(changeTarget, "."); len(parts) != 1 {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`incomplete form submission`))
			return
		} else if shasum = parts[0]; len(shasum) != 10 {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`incomplete form submission`))
			return
		}

		var defArg string
		defTag := f.Enjin.SiteDefaultLanguage()
		if defMsg, ok := ld.Data[shasum][defTag]; ok {
			numerics := defMsg.Placeholders.Numeric()
			if len(numerics) > 0 {
				defArg = numerics[0].ID
			}
		}
		if defArg == "" {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`translation cannot be converted to plural form`))
			return
		}

		for _, msg := range ld.Data[shasum] {
			msg.Translation.Select = &Select{
				Arg:     defArg,
				Feature: "plural",
				Cases: map[string]string{
					"other": msg.Translation.String,
				},
			}
			msg.Translation.String = ""
		}

	case "flatten-translation":

		var shasum string
		if parts := strings.Split(changeTarget, "."); len(parts) != 1 {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`incomplete form submission`))
			return
		} else if shasum = parts[0]; len(shasum) != 10 {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`incomplete form submission`))
			return
		}

		for _, msg := range ld.Data[shasum] {
			if msg.Translation.Select != nil {
				if other, ok := msg.Translation.Select.Cases["other"]; ok {
					msg.Translation.String = other
					msg.Translation.Select = nil
				} else if len(msg.Translation.Select.Cases) > 0 {
					for _, text := range msg.Translation.Select.Cases {
						msg.Translation.String = text
						msg.Translation.Select = nil
						break
					}
				} else {
					msg.Translation.String = msg.Key
					msg.Translation.Select = nil
				}
			}
		}

	default:
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`unknown operation`))
		return

	}

	if changed {
		if err = f.WriteDraftLocales(ld, mountPoints); err != nil {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error writing draft changes: %[1]s`, err.Error()))
			return
		}
		f.Editor.Site().PushInfoNotice(eid, true, printer.Sprintf(`draft locale changes saved`))
	}

	redirect = f.makePaginationRedirect(info, form)
	return
}

func (f *CFeature) OpSearchValidate(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string) (err error) {
	//printer := message.GetPrinter(r)
	//if searchQuery, _ := form["search.query"]; searchQuery == "" {
	//	err = errors.New(printer.Sprintf("missing search query"))
	//}
	return
}

func (f *CFeature) OpSearchHandler(r *http.Request, pg feature.Page, ctx context.Context, form context.Context, info *editor.File, eid string) (redirect string) {
	//log.DebugRF(r, "searching: info=%#+v; form=%#+v", info, form)
	//translations, _ := form["tx"].(map[string]interface{})
	//printer := message.GetPrinter(r)
	//
	//var searchQuery string
	//if searchQuery, _ = form["search.query"].(string); searchQuery == "" {
	//	redirect = f.GetEditorPath() + "/" + info.EditCodeFilePath()
	//	return
	//}

	submit := r.PostFormValue("submit")
	_, target := feature.ParseEditorOpKey(submit)
	changeOp, changeTarget := feature.ParseEditorOpKey(target)
	log.DebugRF(r, `target=%v; changeOp=%v; changeTarget=%v`, target, changeOp, changeTarget)

	if changeOp == "cancel" {
		form.Delete("search.query")
	}

	redirect = f.makePaginationRedirect(info, form)
	return
}
