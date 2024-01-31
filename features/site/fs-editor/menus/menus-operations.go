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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-corelibs/x-text/message"
	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/editor"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/menu"
	clPath "github.com/go-corelibs/path"
)

func (f *CFeature) OpChangeValidate(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string) (err error) {
	printer := message.GetPrinter(r)
	if info.Locked {
		err = errors.New(printer.Sprintf("%[1]s is locked by another user, cannot make changes", info.Name))
	} else if _, present := form["menu"]; !present {
		err = errors.New(printer.Sprintf("incomplete form submitted"))
	}
	return
}

func (f *CFeature) OpChangeHandler(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string) (redirect string) {
	printer := message.GetPrinter(r)

	_, target := feature.ParseEditorOpKey(r.PostFormValue("submit"))

	switch target {
	case "append":
	case "expand":
	case "collapse":
	default:
		_ = form.SetKV("."+target, true)
	}

	var parsed menu.EditMenu
	if v, ok := form["menu"].([]interface{}); ok {
		if parsed, redirect = f.ParseFormToDraft(v, info, r); redirect != "" {
			return
		}
	}

	parsed.SanitizeAll()
	parsed = parsed.ProcessAllChanges()

	switch target {
	case "expand":
		parsed.ExpandAll()
	case "collapse":
		parsed.CollapseAll()
	case "append":
		parsed = append(parsed, &menu.EditItem{})
	}

	fileContents := parsed.String()
	if err := f.SelfEditor().WriteDraft(info, []byte(fileContents)); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf("error saving draft changes: \"%[1]s\"", err.Error()))
		return
	}

	f.SelfEditor().UpdateFileInfo(info, r)
	f.SelfEditor().UpdateFileInfoForEditing(info, r)

	ctx.SetSpecific("TopMenu", parsed)
	redirect = f.SelfEditor().GetEditorPath() + "/" + info.EditFilePath()

	switch target {
	case "append":
	case "expand":
	case "collapse":
	default:
		redirect += "#" + editor.MakeScrollToKey("."+target)
	}
	return
}

func (f *CFeature) OpFileCommitValidate(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string) (err error) {
	printer := message.GetPrinter(r)
	if info.Locked {
		err = errors.New(printer.Sprintf("%[1]s is locked by another user, cannot make changes", info.Name))
	} else if _, present := form["menu"]; !present {
		err = errors.New(printer.Sprintf("incomplete form submitted"))
	}
	return
}

func (f *CFeature) OpFileCommitHandler(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string) (redirect string) {
	printer := message.GetPrinter(r)
	var parsed menu.EditMenu
	if v, ok := form["menu"].([]interface{}); ok {
		if parsed, redirect = f.ParseFormToDraft(v, info, r); redirect != "" {
			return
		}
	}

	parsed.SanitizeAll()
	parsed = parsed.ProcessAllChanges()

	fileContents := parsed.String()
	if err := f.SelfEditor().WriteDraft(info, []byte(fileContents)); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf("error saving draft changes: \"%[1]s\"", err.Error()))
		return
	}

	f.SelfEditor().UpdateFileInfo(info, r)
	f.SelfEditor().UpdateFileInfoForEditing(info, r)

	ctx.SetSpecific("TopMenu", parsed)
	f.Editor.Site().PushInfoNotice(eid, true, printer.Sprintf("%[1]s draft changes saved.", info.File))
	return
}

func (f *CFeature) OpFilePublishValidate(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string) (err error) {
	printer := message.GetPrinter(r)
	if info.Locked {
		err = fmt.Errorf("%s", printer.Sprintf("%[1]s is locked by another user, cannot publish changes", info.Name))
	}
	return
}

func (f *CFeature) OpFilePublishHandler(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string) (redirect string) {
	var err error
	printer := message.GetPrinter(r)

	if v, ok := form["menu"].([]interface{}); ok {
		var parsed menu.EditMenu
		if parsed, redirect = f.ParseFormToDraft(v, info, r); redirect != "" {
			return
		}
		parsed.SanitizeAll()
		parsed = parsed.ProcessAllChanges()

		var cleaned menu.Menu
		if cleaned, redirect = f.ParseDraftToMenu(parsed, info, r); redirect != "" {
			return
		}

		fileContents := cleaned.String()
		if err = f.SelfEditor().WriteDraft(info, []byte(fileContents)); err != nil {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf("error saving final draft changes: \"%[1]s\"", err.Error()))
			return
		}
	}

	var data []byte
	cleaned := menu.Menu{}
	if data, err = f.SelfEditor().ReadDraft(info); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf("error reading final draft: \"%[1]s\"", err.Error()))
		return
	} else if err = json.Unmarshal(data, &cleaned); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf("error decoding draft data: \"%[1]s\"", err.Error()))
		return
	} else if data, err = json.MarshalIndent(cleaned, "", "\t"); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf("error encoding draft menu: \"%[1]s\"", err.Error()))
		return
	} else if err = f.SelfEditor().WriteFile(info, data); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf("error writing file: \"%[1]s\"", err.Error()))
		return
	} else if err = f.SelfEditor().RemoveDraft(info); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf("error removing final draft: \"%[1]s\"", err.Error()))
		return
	} else if err = f.SelfEditor().UnLockEditorFile(info.FSID, info.FilePath()); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf("error unlocking file: \"%[1]s\"", err.Error()))
		return
	}

	f.Editor.Site().PushInfoNotice(eid, true, printer.Sprintf("%[1]s draft changes published.", info.File))
	redirect = f.SelfEditor().GetEditorPath() + "/" + info.EditDirectoryPath()
	return
}

func (f *CFeature) OpMenuCreateValidate(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string) (err error) {
	//printer := message.GetPrinter(r)
	//if info.Locked {
	//	err = errors.New(printer.Sprintf("%[1]s is locked by another user, cannot republish changes", info.Name))
	//}
	return
}

func (f *CFeature) OpMenuCreateHandler(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string) (redirect string) {
	printer := message.GetPrinter(r)
	dstUri, dstInfo, dstFS, dstMP, dstExists, stop := f.ParseCreateMenuForm(r, pg, ctx, form, info, eid, &redirect)
	if stop {
		return
	}

	if dstExists {
		dst := dstUri
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`destination "%[1]s" exists already`, dst))
		return
	} else if dstFS == nil || dstMP == nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`cannot create "%[2]s" on "%[1]s": filesystem not found`, dstInfo.FSID, dstInfo.File))
		return
	} else if dstMP.RWFS == nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`cannot create "%[2]s" on "%[1]s": filesystem is read-only`, dstInfo.FSID, dstInfo.File))
		return
	}

	var err error
	data := []byte("[]")

	realName := clPath.Base(dstInfo.Name) + ".json"
	dstInfo.Name = realName
	dstInfo.File = strings.Replace(dstInfo.File, dstInfo.Name, realName, 1)

	if err = dstMP.RWFS.WriteFile(dstInfo.FilePath(), data, 0664); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error writing "%[1]s": %[2]s`, dstUri, err.Error()))
		return
	}

	f.Editor.Site().PushInfoNotice(eid, true, printer.Sprintf(`create new page "%[1]s"`, dstUri))
	if v, _ := form["return"].(string); v == "directory" {
		redirect = f.SelfEditor().GetEditorPath() + "/" + dstInfo.EditDirectoryPath()
	} else {
		redirect = f.SelfEditor().GetEditorPath() + "/" + dstInfo.EditFilePath()
	}
	return
}
