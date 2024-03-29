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

package fs_editor

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-corelibs/x-text/message"
	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/editor"
	"github.com/go-enjin/be/pkg/feature"
)

func (f *CEditorFeature[MakeTypedFeature]) OpFileUnlockHandler(r *http.Request, pg feature.Page, ctx, form beContext.Context, info *editor.File, eid string) (redirect string) {
	if stop := f.Emit(feature.PreUnlockFileSignal, f.Tag().String(), r, pg, ctx, form, info, eid, &redirect); stop {
		return
	}
	var err error
	printer := message.GetPrinter(r)
	cannotUnlockErr := printer.Sprintf("Cannot unlock another user's file lock")
	if info.Locked {
		f.Editor.Site().PushErrorNotice(eid, true, cannotUnlockErr)
		return
	} else if other, locked := f.IsEditorFileLocked(info.FSID, info.FilePath()); locked && other != eid {
		f.Editor.Site().PushErrorNotice(eid, true, cannotUnlockErr)
		return
	} else if err = f.UnLockEditorFile(info.FSID, info.FilePath()); err != nil {
		return
	}
	f.Editor.Site().PushInfoNotice(eid, true, printer.Sprintf("%[1]s unlocked for others to edit.", info.File))
	f.Emit(feature.UnlockFileSignal, f.Tag().String(), r, pg, ctx, form, info, eid, &redirect)
	return
}

func (f *CEditorFeature[MakeTypedFeature]) OpFileRetakeHandler(r *http.Request, pg feature.Page, ctx, form beContext.Context, info *editor.File, eid string) (redirect string) {
	if stop := f.Emit(feature.PreRetakeFileSignal, f.Tag().String(), r, pg, ctx, form, info, eid, &redirect); stop {
		return
	}
	printer := message.GetPrinter(r)
	var err error
	if err = f.LockEditorFile(eid, info.FSID, info.FilePath()); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf("error locking file for editing: \"%[1]s\"", err.Error()))
		return
	}
	f.Editor.Site().PushInfoNotice(eid, true, printer.Sprintf("%[1]s editing taken over.", info.File))
	f.Emit(feature.RetakeFileSignal, f.Tag().String(), r, pg, ctx, form, info, eid, &redirect)
	return
}

func (f *CEditorFeature[MakeTypedFeature]) OpFileDeleteValidate(r *http.Request, pg feature.Page, ctx, form beContext.Context, info *editor.File, eid string) (err error) {
	printer := message.GetPrinter(r)
	if info.Locked {
		err = errors.New(printer.Sprintf("%[1]s is locked by another user, cannot delete", info.Name))
	}
	return
}

func (f *CEditorFeature[MakeTypedFeature]) OpFileDeleteHandler(r *http.Request, pg feature.Page, ctx, form beContext.Context, info *editor.File, eid string) (redirect string) {
	if stop := f.Emit(feature.PreDeleteFileSignal, f.Tag().String(), r, pg, ctx, form, info, eid, &redirect); stop {
		return
	}
	printer := message.GetPrinter(r)

	if lockedBy, locked := f.IsEditorFileLocked(info.FSID, info.FilePath()); locked && eid != lockedBy {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf("Cannot delete, file is locked by another user"))
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
		f.Editor.Site().PushInfoNotice(eid, true, printer.Sprintf("%[1]s draft deleted.", info.File))

	default:
		if err = f.SelfEditor().RemoveFile(info); err != nil {
			return
		}
		f.Editor.Site().PushInfoNotice(eid, true, printer.Sprintf("%[1]s file deleted.", info.File))
		redirect = f.SelfEditor().GetEditorPath() + "/" + info.EditDirectoryPath()
	}

	f.Emit(feature.DeletePathSignal, f.Tag().String(), r, pg, ctx, form, info, eid, &redirect)
	return
}

func (f *CEditorFeature[MakeTypedFeature]) OpPathDeleteValidate(r *http.Request, pg feature.Page, ctx, form beContext.Context, info *editor.File, eid string) (err error) {
	printer := message.GetPrinter(r)
	if foundFiles := f.SelfEditor().ListFileSystemFiles(r, info.FSID, info.Locale.String(), info.BaseNamePath()); len(foundFiles) > 0 {
		err = errors.New(printer.Sprintf(`cannot delete "%[1]s": directory not empty`, info.Name))
	} else if foundDirs := f.SelfEditor().ListFileSystemDirectories(r, info.FSID, info.Locale.String(), info.BaseNamePath()); len(foundDirs) > 0 {
		err = errors.New(printer.Sprintf(`cannot delete "%[1]s": directory not empty`, info.Name))
	}
	return
}

func (f *CEditorFeature[MakeTypedFeature]) OpPathDeleteHandler(r *http.Request, pg feature.Page, ctx, form beContext.Context, info *editor.File, eid string) (redirect string) {
	if stop := f.Emit(feature.PreDeletePathSignal, f.Tag().String(), r, pg, ctx, form, info, eid, &redirect); stop {
		return
	}
	printer := message.GetPrinter(r)
	if err := f.SelfEditor().RemoveDirectory(info); err != nil {
		return
	}
	f.Editor.Site().PushInfoNotice(eid, true, printer.Sprintf("%[1]s directory deleted.", info.Name))
	redirect = f.SelfEditor().GetEditorPath() + "/" + info.EditParentDirectoryPath()
	f.Emit(feature.DeletePathSignal, f.Tag().String(), r, pg, ctx, form, info, eid, &redirect)
	return
}

func (f *CEditorFeature[MakeTypedFeature]) OpFileCommitValidate(r *http.Request, pg feature.Page, ctx, form beContext.Context, info *editor.File, eid string) (err error) {
	printer := message.GetPrinter(r)
	if info.Locked {
		err = errors.New(printer.Sprintf("%[1]s is locked by another user, cannot make changes", info.Name))
	} else if _, present := form["body"]; !present {
		err = errors.New(printer.Sprintf("incomplete form submitted"))
	}
	return
}

func (f *CEditorFeature[MakeTypedFeature]) OpFileCommitHandler(r *http.Request, pg feature.Page, ctx, form beContext.Context, info *editor.File, eid string) (redirect string) {
	if stop := f.Emit(feature.PreCommitFileSignal, f.Tag().String(), r, pg, ctx, form, info, eid, &redirect); stop {
		return
	}
	printer := message.GetPrinter(r)

	body, _ := form["body"].(string)
	body = strings.ReplaceAll(body, "\r", "")

	if err := f.SelfEditor().WriteDraft(info, []byte(body)); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf("error saving to draft file: \"%[1]s\"", err.Error()))
		return
	}

	f.SelfEditor().UpdateFileInfo(info, r)
	f.SelfEditor().UpdateFileInfoForEditing(info, r)

	ctx.SetSpecific("Body", body)
	f.Editor.Site().PushInfoNotice(eid, true, printer.Sprintf("%[1]s draft changes saved.", info.File))
	f.Emit(feature.CommitFileSignal, f.Tag().String(), r, pg, ctx, form, info, eid, &redirect)
	return
}

func (f *CEditorFeature[MakeTypedFeature]) OpFilePublishValidate(r *http.Request, pg feature.Page, ctx, form beContext.Context, info *editor.File, eid string) (err error) {
	printer := message.GetPrinter(r)
	if info.Locked {
		err = errors.New(printer.Sprintf("%[1]s is locked by another user, cannot publish changes", info.Name))
	}
	return
}

func (f *CEditorFeature[MakeTypedFeature]) OpFilePublishHandler(r *http.Request, pg feature.Page, ctx, form beContext.Context, info *editor.File, eid string) (redirect string) {
	if stop := f.Emit(feature.PrePublishFileSignal, f.Tag().String(), r, pg, ctx, form, info, eid, &redirect); stop {
		return
	}
	printer := message.GetPrinter(r)

	var err error
	if body, present := form["body"].(string); present {
		body = strings.ReplaceAll(body, "\r", "")
		if err = f.SelfEditor().WriteDraft(info, []byte(body)); err != nil {
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf("error saving final changes to draft: \"%[1]s\"", err.Error()))
			return
		}
		ctx.SetSpecific("Body", body)
	}

	//if err = f.SelfEditor().PublishDraft(info); err != nil {
	//	f.Editor.Site().PushErrorNotice(eid, err.Error(), true)
	//	return
	//}

	var data []byte
	if data, err = f.SelfEditor().ReadDraft(info); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf("error reading final draft: \"%[1]s\"", err.Error()))
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
	f.Emit(feature.PublishFileSignal, f.Tag().String(), r, pg, ctx, form, info, eid, &redirect)
	return
}

func (f *CEditorFeature[MakeTypedFeature]) OpFileCancelValidate(r *http.Request, pg feature.Page, ctx, form beContext.Context, info *editor.File, eid string) (err error) {
	printer := message.GetPrinter(r)
	if info.Locked {
		err = errors.New(printer.Sprintf("file is locked by another user, cannot move"))
	}
	return
}

func (f *CEditorFeature[MakeTypedFeature]) OpFileCancelHandler(r *http.Request, pg feature.Page, ctx, form beContext.Context, info *editor.File, eid string) (redirect string) {
	if stop := f.Emit(feature.PrePublishFileSignal, f.Tag().String(), r, pg, ctx, form, info, eid, &redirect); stop {
		return
	}
	_ = f.OpFileUnlockHandler(r, pg, ctx, form, info, eid)
	redirect = f.SelfEditor().GetEditorPath() + "/" + info.EditDirectoryPath()
	f.Emit(feature.PublishFileSignal, f.Tag().String(), r, pg, ctx, form, info, eid, &redirect)
	return
}

func (f *CEditorFeature[MakeTypedFeature]) OpFileMoveValidate(r *http.Request, pg feature.Page, ctx, form beContext.Context, info *editor.File, eid string) (err error) {
	printer := message.GetPrinter(r)
	if info.Locked {
		err = errors.New(printer.Sprintf("file is locked by another user, cannot move"))
	}
	return
}

func (f *CEditorFeature[MakeTypedFeature]) OpFileMoveHandler(r *http.Request, pg feature.Page, ctx, form beContext.Context, info *editor.File, eid string) (redirect string) {
	if stop := f.Emit(feature.PreMoveFileSignal, f.Tag().String(), r, pg, ctx, form, info, eid, &redirect); stop {
		return
	}
	printer := message.GetPrinter(r)

	srcUri, dstUri, dstInfo, srcFS, dstFS, srcMP, dstMP, srcExists, dstExists, stop := f.ParseCopyMoveTranslateForm(r, pg, ctx, form, info, eid, &redirect)
	if stop {
		return
	}
	if !srcExists {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`"%[1]s" not found`, srcUri))
		return
	} else if dstExists {
		dst := srcUri
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`destination "%[1]s" exists already`, dst))
		return
	} else if dstFS == nil || dstMP == nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`cannot move "%[2]s" to "%[1]s": filesystem not found`, dstInfo.FSID, srcUri))
		return
	} else if dstMP.RWFS == nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`cannot move "%[2]s" to "%[1]s": filesystem is read-only`, dstInfo.FSID, srcUri))
		return
	} else if srcFS == nil || srcMP == nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`cannot move "%[2]s" from "%[1]s": filesystem not found`, info.FSID, srcUri))
		return
	} else if srcMP.RWFS == nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`cannot move "%[2]s" from "%[1]s": filesystem is read-only`, info.FSID, srcUri))
		return
	}

	var err error
	var srcData []byte
	var srcShasum, dstShasum string
	var created, updated time.Time
	if _, srcShasum, created, updated, err = srcMP.RWFS.FileStats(info.FilePath()); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error reading "%[1]s" file stats: %[2]s`, srcUri, err.Error()))
		return
	} else if srcData, err = srcMP.RWFS.ReadFile(info.FilePath()); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error reading "%[1]s": %[2]s`, srcUri, err.Error()))
		return
	} else if err = dstMP.RWFS.WriteFile(dstInfo.FilePath(), srcData, 0664); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error writing "%[1]s": %[2]s`, dstUri, err.Error()))
		return
	} else if err = dstMP.RWFS.ChangeTimes(dstInfo.FilePath(), created, updated); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error preserving file timestamps "%[1]s": %[2]s`, dstUri, err.Error()))
		return
	} else if dstShasum, err = dstMP.RWFS.Shasum(dstInfo.FilePath()); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error reading "%[1]s" shasum: %[2]s`, dstUri, err.Error()))
		return
	} else if srcShasum != dstShasum {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`source and destination file shasums differ`))
		return
	} else if err = f.UnLockEditorFile(info.FSID, info.FilePath()); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error unlocking source file "%[1]s" before deleting during move: %[2]s`, srcUri, err.Error()))
		return
	} else if err = srcMP.RWFS.Remove(info.FilePath()); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error removing "%[1]s": %[2]s`, srcUri, err.Error()))
		return
	}

	f.Editor.Site().PushInfoNotice(eid, true, printer.Sprintf(`moved "%[1]s" to "%[2]s"`, srcUri, dstUri))
	if v, _ := form["return"].(string); v == "directory" {
		redirect = f.SelfEditor().GetEditorPath() + "/" + dstInfo.EditDirectoryPath()
	} else {
		redirect = f.SelfEditor().GetEditorPath() + "/" + dstInfo.EditFilePath()
	}

	f.Emit(feature.MoveFileSignal, f.Tag().String(), r, pg, ctx, form, dstInfo, eid, &redirect)
	return
}

func (f *CEditorFeature[MakeTypedFeature]) OpFileCopyValidate(r *http.Request, pg feature.Page, ctx, form beContext.Context, info *editor.File, eid string) (err error) {
	return
}

func (f *CEditorFeature[MakeTypedFeature]) OpFileCopyHandler(r *http.Request, pg feature.Page, ctx, form beContext.Context, info *editor.File, eid string) (redirect string) {
	if stop := f.Emit(feature.PreCopyFileSignal, f.Tag().String(), r, pg, ctx, form, info, eid, &redirect); stop {
		return
	}
	printer := message.GetPrinter(r)

	srcUri, dstUri, dstInfo, srcFS, dstFS, srcMP, dstMP, srcExists, dstExists, stop := f.ParseCopyMoveTranslateForm(r, pg, ctx, form, info, eid, &redirect)
	if stop {
		return
	}

	if !srcExists {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`"%[1]s" not found`, srcUri))
		return
	} else if dstExists {
		dst := srcUri
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`destination "%[1]s" exists already`, dst))
		return
	} else if dstFS == nil || dstMP == nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`cannot copy "%[2]s" to "%[1]s": filesystem not found`, dstInfo.FSID, srcUri))
		return
	} else if dstMP.RWFS == nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`cannot copy "%[2]s" to "%[1]s": filesystem is read-only`, dstInfo.FSID, srcUri))
		return
	} else if srcFS == nil || srcMP == nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`cannot copy "%[2]s" from "%[1]s": filesystem not found`, info.FSID, srcUri))
		return
	}

	var err error
	var srcData []byte
	var srcShasum, dstShasum string
	var created, updated time.Time
	if _, srcShasum, created, updated, err = srcMP.ROFS.FileStats(info.FilePath()); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error reading "%[1]s" file stats: %[2]s`, srcUri, err.Error()))
		return
	} else if srcData, err = srcMP.ROFS.ReadFile(info.FilePath()); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error reading "%[1]s": %[2]s`, srcUri, err.Error()))
		return
	} else if err = dstMP.RWFS.WriteFile(dstInfo.FilePath(), srcData, 0664); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error writing "%[1]s": %[2]s`, dstUri, err.Error()))
		return
	} else if err = dstMP.RWFS.ChangeTimes(dstInfo.FilePath(), created, updated); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error preserving file timestamps "%[1]s": %[2]s`, dstUri, err.Error()))
		return
	} else if dstShasum, err = dstMP.RWFS.Shasum(dstInfo.FilePath()); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error reading "%[1]s" shasum: %[2]s`, dstUri, err.Error()))
		return
	} else if srcShasum != dstShasum {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`source and destination file shasums differ`))
		return
	}

	f.Editor.Site().PushInfoNotice(eid, true, printer.Sprintf(`copied "%[1]s" to "%[2]s"`, srcUri, dstUri))
	if v, _ := form["return"].(string); v == "directory" {
		redirect = f.SelfEditor().GetEditorPath() + "/" + dstInfo.EditDirectoryPath()
	} else {
		redirect = f.SelfEditor().GetEditorPath() + "/" + dstInfo.EditFilePath()
	}
	f.Emit(feature.CopyFileSignal, f.Tag().String(), r, pg, ctx, form, info, eid, &redirect)
	return
}

func (f *CEditorFeature[MakeTypedFeature]) OpFileTranslateValidate(r *http.Request, pg feature.Page, ctx, form beContext.Context, info *editor.File, eid string) (err error) {
	//printer := message.GetPrinter(r)
	//if info.Locked {
	//	err = errors.New(printer.Sprintf("%[1]s is locked by another user, cannot republish changes", info.Name))
	//}
	return
}

func (f *CEditorFeature[MakeTypedFeature]) OpFileTranslateHandler(r *http.Request, pg feature.Page, ctx, form beContext.Context, info *editor.File, eid string) (redirect string) {
	if stop := f.Emit(feature.PreTranslateFileActionSignal, f.Tag().String(), r, pg, ctx, form, info, eid, &redirect); stop {
		return
	}

	printer := message.GetPrinter(r)

	srcUri, dstUri, dstInfo, srcFS, dstFS, srcMP, dstMP, srcExists, dstExists, stop := f.ParseCopyMoveTranslateForm(r, pg, ctx, form, info, eid, &redirect)
	if stop {
		return
	}

	if !srcExists {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`"%[1]s" not found`, srcUri))
		return
	} else if dstExists {
		dst := srcUri
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`destination "%[1]s" exists already`, dst))
		return
	} else if dstFS == nil || dstMP == nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`cannot copy "%[2]s" to "%[1]s": filesystem not found`, dstInfo.FSID, srcUri))
		return
	} else if dstMP.RWFS == nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`cannot copy "%[2]s" to "%[1]s": filesystem is read-only`, dstInfo.FSID, srcUri))
		return
	} else if srcFS == nil || srcMP == nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`cannot copy "%[2]s" from "%[1]s": filesystem not found`, info.FSID, srcUri))
		return
	}

	var err error
	var srcData []byte
	var srcShasum, dstShasum string
	var created, updated time.Time
	if _, srcShasum, created, updated, err = srcMP.ROFS.FileStats(info.FilePath()); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error reading "%[1]s" file stats: %[2]s`, srcUri, err.Error()))
		return
	} else if srcData, err = srcMP.ROFS.ReadFile(info.FilePath()); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error reading "%[1]s": %[2]s`, srcUri, err.Error()))
		return
	} else if err = dstMP.RWFS.WriteFile(dstInfo.FilePath(), srcData, 0664); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error writing "%[1]s": %[2]s`, dstUri, err.Error()))
		return
	} else if err = dstMP.RWFS.ChangeTimes(dstInfo.FilePath(), created, updated); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error preserving file timestamps "%[1]s": %[2]s`, dstUri, err.Error()))
		return
	} else if dstShasum, err = dstMP.RWFS.Shasum(dstInfo.FilePath()); err != nil {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error reading "%[1]s" shasum: %[2]s`, dstUri, err.Error()))
		return
	} else if srcShasum != dstShasum {
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`source and destination file shasums differ`))
		return
	}

	if v, _ := form["return"].(string); v == "directory" {
		redirect = f.SelfEditor().GetEditorPath() + "/" + dstInfo.EditDirectoryPath()
	} else {
		redirect = f.SelfEditor().GetEditorPath() + "/" + dstInfo.EditFilePath()
	}

	f.Editor.Site().PushInfoNotice(eid, true, printer.Sprintf(`"%[1]s" %[2]s translation started`, info.Name, dstInfo.Locale.String()))

	f.Emit(feature.TranslateFileActionSignal, f.Tag().String(), r, pg, ctx, form, info, eid, &redirect)
	return
}
