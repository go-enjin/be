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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/editor"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/hash/sha"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/lang/catalog"
	"github.com/go-enjin/be/pkg/log"
	beMime "github.com/go-enjin/be/pkg/mime"
	bePath "github.com/go-enjin/be/pkg/path"
	"github.com/go-enjin/be/pkg/userbase"
)

func (f *CFeature) FindFS(fsid string) (found feature.FileSystemFeature) {

	for _, efs := range f.EditingFileSystems {
		if efs.Tag().String() == fsid {
			found = efs
			return
		}
	}

	return
}

func (f *CFeature) FindMountedPoints(fsid string) (mp feature.MountedPoints) {
	if found := f.FindFS(fsid); found != nil {
		mp = make(feature.MountedPoints)
		for mount, mountedPoints := range found.GetMountedPoints() {
			name := strings.TrimPrefix(mount, "/")
			for _, mountPoint := range mountedPoints {
				mp[name] = append(mp[name], mountPoint)
			}
		}
	}
	return
}

func (f *CFeature) FindMountPoints(fsid, code string) (mountPoints feature.MountPoints) {

	var mountedPoints feature.MountedPoints
	if mountedPoints = f.FindMountedPoints(fsid); mountedPoints == nil {
		return
	}

	//var mountPoints feature.MountPoints
	if countMountedPoints := len(mountedPoints); countMountedPoints > 1 {
		if v, present := mountedPoints[code]; present {
			mountPoints = v
		}
	} else if countMountedPoints == 1 {
		if v, present := mountedPoints[code]; present {
			mountPoints = v
		}
	}

	return
}

func (f *CFeature) WriteLocales(ld *LocaleData, mountPoints feature.MountPoints) (err error) {
	lookup := ld.MakeGoTextData()
	if len(lookup) == 0 {
		for _, tag := range f.Enjin.SiteLocales() {
			lookup[tag] = &catalog.GoText{Language: tag.String()}
		}
	}
	for _, mountPoint := range mountPoints {
		if mountPoint.RWFS != nil {
			var prefix string
			if mountPoint.RWFS.Exists("locales") {
				prefix += "locales/"
			}
			for tag, gtd := range lookup {
				var data []byte
				if data, err = json.MarshalIndent(gtd, "", "    "); err != nil {
					return
				}
				msgDst := prefix + tag.String() + "/messages.gotext.json"
				if err = mountPoint.RWFS.WriteFile(msgDst, data, 0660); err != nil {
					return
				}
				outDst := prefix + tag.String() + "/out.gotext.json"
				if err = mountPoint.RWFS.WriteFile(outDst, data, 0660); err != nil {
					return
				}
			}
			break
		}
	}
	return
}

func (f *CFeature) WriteDraftLocales(ld *LocaleData, mountPoints feature.MountPoints) (err error) {
	lookup := ld.MakeGoTextData()
	if len(lookup) == 0 {
		for _, tag := range f.Enjin.SiteLocales() {
			lookup[tag] = &catalog.GoText{Language: tag.String()}
		}
	}
	for _, mountPoint := range mountPoints {
		if mountPoint.RWFS != nil {
			var prefix string
			if mountPoint.RWFS.Exists("locales") {
				prefix += "locales/"
			}
			for tag, gtd := range lookup {
				var data []byte
				if data, err = json.MarshalIndent(gtd, "", "    "); err != nil {
					return
				}
				destination := prefix + tag.String() + "/messages.gotext.json.~draft"
				if err = mountPoint.RWFS.WriteFile(destination, data, 0660); err != nil {
					return
				}
			}
			break
		}
	}
	return
}

func (f *CFeature) DeleteDraftLocales(mountPoints feature.MountPoints) {
	locales := f.Enjin.SiteLocales()
	for _, mountPoint := range mountPoints {
		if mountPoint.RWFS != nil {
			var prefix string
			if mountPoint.RWFS.Exists("locales") {
				prefix += "locales/"
			}
			for _, tag := range locales {
				for _, name := range []string{"messages", "out"} {
					target := prefix + tag.String() + "/" + name + ".gotext.json.~draft"
					_ = mountPoint.RWFS.Remove(target)
				}
			}
		}
	}
	return
}

func (f *CFeature) ReadDraftLocales(fsid, code string, mountPoints feature.MountPoints, withOut bool) (ld *LocaleData, err error) {
	ld = &LocaleData{
		FSID: fsid,
		Code: code,
		Data: make(map[string]map[language.Tag]*LocaleMessage),
	}

	unique := make(map[string]struct{})
	process := func(tag language.Tag, parsed *catalog.GoText) {
		for _, msg := range parsed.Messages {
			shasum, _ := sha.DataHash10(msg.Key)
			if _, present := ld.Data[shasum]; !present {
				ld.Data[shasum] = make(map[language.Tag]*LocaleMessage)
			}
			if _, present := ld.Data[shasum][tag]; present {
				continue
			}
			var txSelect *Select
			if msg.Translation.Select != nil {
				cases := map[string]string{}
				for k, v := range msg.Translation.Select.Cases {
					cases[k] = ConvertToPlaceholders(v.Msg, msg.Placeholders)
				}
				txSelect = &Select{
					Arg:     msg.Translation.Select.Arg,
					Feature: msg.Translation.Select.Feature,
					Cases:   cases,
				}
			}
			if _, present := unique[shasum]; !present {
				ld.Order = append(ld.Order, shasum)
				unique[shasum] = struct{}{}
			}
			ld.Data[shasum][tag] = &LocaleMessage{
				ID:      msg.ID,
				Key:     msg.Key,
				Message: msg.Message,
				Translation: &LocaleTranslation{
					String: ConvertToPlaceholders(msg.Translation.String, msg.Placeholders),
					Select: txSelect,
				},
				TranslatorComment: msg.TranslatorComment,
				Placeholders:      msg.Placeholders[:],
				Fuzzy:             msg.Fuzzy,
				Shasum:            shasum,
			}
		}
	}

	for _, mountPoint := range mountPoints {
		var prefixDir string
		dirs, _ := mountPoint.ROFS.ListDirs(".")
		for _, dir := range dirs {
			if dir == "locales" {
				prefixDir = "locales/"
				dirs, _ = mountPoint.ROFS.ListDirs("locales")
				break
			}
		}
		var localeDirs []language.Tag
		for _, dir := range dirs {
			dir = strings.TrimPrefix(dir, "locales/")
			if parsed, eee := language.Parse(dir); eee == nil && parsed != language.Und {
				localeDirs = append(localeDirs, parsed)
			}
		}

		if len(localeDirs) > 0 {

			for _, localeDir := range localeDirs {

				var foundDraft, readingOut bool
				prefix := prefixDir + localeDir.String() + "/"
				msgPath := prefix + "messages.gotext.json.~draft"
				if foundDraft = mountPoint.RWFS.Exists(msgPath); !foundDraft {
					msgPath = prefix + "messages.gotext.json"
					if readingOut = !mountPoint.RWFS.Exists(msgPath); readingOut {
						msgPath = prefix + "out.gotext.json"
					}
				}

				if data, ee := mountPoint.ROFS.ReadFile(msgPath); ee == nil {
					if parsed, _, eee := catalog.ParseGoText(data); eee == nil {
						process(localeDir, parsed)
					}
				}

				if withOut {
					if foundDraft || !readingOut {
						outPath := prefix + "out.gotext.json"
						if msgPath != outPath {
							if data, ee := mountPoint.ROFS.ReadFile(outPath); ee == nil {
								if parsed, _, eee := catalog.ParseGoText(data); eee == nil {
									process(localeDir, parsed)
								}
							}
						}
					}
				}

			}

		}

	}

	ld.AddMissingTranslations(f.Enjin.SiteDefaultLanguage(), f.Enjin.SiteLocales())
	return
}

func (f *CFeature) MakeTable(ld *LocaleData) (table []Row, err error) {

	tags := f.Enjin.SiteLocales()
	defTag := f.Enjin.SiteDefaultLanguage()

	var ok bool
	for _, shasum := range ld.Order {
		var msgs map[language.Tag]*LocaleMessage
		if msgs, ok = ld.Data[shasum]; ok {
			row := make(Row, 0)
			var defMsg *LocaleMessage
			if defMsg, ok = msgs[defTag]; ok {
				row = append(row, &Cell{Locale: defTag, Src: defMsg, Msg: defMsg})
			} else {
				row = append(row, &Cell{Locale: defTag, Src: nil, Msg: nil})
			}
			for _, tag := range tags {
				var msg *LocaleMessage
				if msg, ok = msgs[tag]; ok {
					row = append(row, &Cell{Locale: tag, Src: defMsg, Msg: msg})
				} else {
					row = append(row, &Cell{Locale: tag, Src: defMsg, Msg: nil})
				}
			}
			table = append(table, row)
		}
	}

	return
}

func (f *CFeature) HasDraftLocales(fsid, code string) (present bool) {
	var found feature.FileSystemFeature
	if found = f.FindFS(fsid); found == nil {
		return
	}
	check := "/" + code
	locales := f.Enjin.SiteLocales()
	for mount, mountPoints := range found.GetMountedPoints() {
		if mount == check {
			for _, mountPoint := range mountPoints {
				if mountPoint.RWFS != nil {
					var prefix string
					if mountPoint.RWFS.Exists("locales") {
						prefix += "locales/"
					}
					for _, tag := range locales {
						if present = mountPoint.RWFS.Exists(prefix + tag.String() + "/messages.gotext.json.~draft"); !present {
							return
						}
					}
				}
			}
		}
	}
	return
}

func (f *CFeature) IsLocaleLocked(fsid, code string) (locked bool, eid string, err error) {
	var found feature.FileSystemFeature
	if found = f.FindFS(fsid); found == nil {
		err = fmt.Errorf("filesystem not found")
		return
	}

	process := func(rwfs fs.RWFileSystem, files ...string) (locked bool, eid string) {
		for _, file := range files {
			lockfile := file + ".~lock"
			if locked = rwfs.Exists(lockfile); !locked {
				continue
			} else if data, err := rwfs.ReadFile(lockfile); err == nil {
				if v := string(data); len(v) == 10 || v == userbase.VisitorEID {
					eid = v
				} else {
					log.ErrorF("lockfile contains invalid data: %v", lockfile)
				}
			}

		}
		return
	}

	locales := f.Enjin.SiteLocales()
	check := "/" + bePath.TrimSlashes(code)
	for mount, mountPoints := range found.GetMountedPoints() {
		if mount == check {
			for _, mountPoint := range mountPoints {
				if mountPoint.RWFS != nil {
					for _, tag := range locales {
						prefix := tag.String() + "/"
						if mountPoint.RWFS.Exists("locales") {
							prefix = "locales/" + prefix
						}
						if locked, eid = process(mountPoint.RWFS,
							prefix+"messages.gotext.json",
							prefix+"out.gotext.json",
						); !locked {
							return
						}
					}
				}
			}
		}
	}
	return
}

func (f *CFeature) LockLocale(eid, fsid, code string) (err error) {
	var found feature.FileSystemFeature
	if found = f.FindFS(fsid); found == nil {
		err = fmt.Errorf("filesystem not found")
		return
	}

	process := func(rwfs fs.RWFileSystem, files ...string) (err error) {
		for _, file := range files {
			if rwfs.Exists(file) {
				if err = rwfs.WriteFile(file+".~lock", []byte(eid), 0660); err != nil {
					return
				}
			}
		}
		return
	}

	locales := f.Enjin.SiteLocales()
	check := "/" + bePath.TrimSlashes(code)
	for mount, mountPoints := range found.GetMountedPoints() {
		if mount == check {
			for _, mountPoint := range mountPoints {
				if mountPoint.RWFS != nil {
					for _, tag := range locales {
						prefix := tag.String() + "/"
						if mountPoint.RWFS.Exists("locales") {
							prefix = "locales/" + prefix
						}
						if err = process(mountPoint.RWFS,
							prefix+"messages.gotext.json",
							prefix+"out.gotext.json",
						); err != nil {
							return
						}
					}
				}
			}
		}
	}
	return
}

func (f *CFeature) UnlockLocales(fsid, code string) (err error) {
	var found feature.FileSystemFeature
	if found = f.FindFS(fsid); found == nil {
		err = fmt.Errorf("filesystem not found")
		return
	}

	process := func(rwfs fs.RWFileSystem, files ...string) (err error) {
		for _, file := range files {
			if rwfs.Exists(file + ".~lock") {
				if err = rwfs.Remove(file + ".~lock"); err != nil {
					return
				}
			}
		}
		return
	}

	locales := f.Enjin.SiteLocales()
	check := "/" + bePath.TrimSlashes(code)
	for mount, mountPoints := range found.GetMountedPoints() {
		if mount == check {
			for _, mountPoint := range mountPoints {
				if mountPoint.RWFS != nil {
					for _, tag := range locales {
						prefix := tag.String() + "/"
						if mountPoint.RWFS.Exists("locales") {
							prefix = "locales/" + prefix
						}
						if err = process(mountPoint.RWFS,
							prefix+"messages.gotext.json",
							prefix+"out.gotext.json",
						); err != nil {
							return
						}
					}
				}
			}
		}
	}
	return
}

func (f *CFeature) UpdatePathInfo(info *editor.File, r *http.Request) {
	// page-level actions (floating bottom-right button menu actions)
	printer := lang.GetPrinterFromRequest(r)

	if !info.ReadOnly {
		info.Actions = append(info.Actions, editor.MakeCreatePageAction(printer))
	}

	info.Actions = info.Actions.Sort()
	return
}

func (f *CFeature) UpdateFileInfo(info *editor.File, r *http.Request) {
	// browser row actions and editing page
	//f.CEditorFeature.UpdateFileInfo(info, r)
	printer := lang.GetPrinterFromRequest(r)
	eid := userbase.GetCurrentEID(r)

	if info.Locked {
		info.Actions = append(info.Actions, editor.MakeRetakeFileAction(printer, info.EditCodeFilePath()))
		info.Actions = info.Actions.Prune(editor.CancelActionKey, editor.CommitActionKey)
	} else if !info.Locked && info.LockedBy == eid {
		if !info.Actions.Has(editor.CancelActionKey) {
			info.Actions = append(info.Actions, editor.MakeUnlockFileAction(printer, info.EditCodeFilePath()))
		}
		if info.HasDraft {
			info.Actions = append(info.Actions, editor.MakePublishFileAction(printer, info.EditCodeFilePath()))
			info.Actions = append(info.Actions, editor.MakeDeleteDraftFileAction(printer, info.EditCodeFilePath()))
		}
	}

	info.Actions = info.Actions.Sort()
}

func (f *CFeature) ListLocales(r *http.Request) (list editor.Files) {
	//eid := userbase.GetCurrentEID(r)
	for _, mpf := range f.EditingFileSystems {
		bt := mpf.BaseTag().String()
		fsid := mpf.Tag().String()
		mountedPoints := mpf.GetMountedPoints()

		var infos []*editor.File
		for mount, mountPoints := range mountedPoints {
			rw := mountPoints.HasRWFS()
			if mount == "/" {
				infos = append(infos, &editor.File{
					FSBT:     bt,
					FSID:     fsid,
					Name:     fsid,
					ReadOnly: rw,
					MimeType: beMime.DirectoryMimeType,
				})
			} else {
				code := strings.TrimPrefix(mount, "/")
				infos = append(infos, &editor.File{
					FSBT:     bt,
					FSID:     fsid,
					Code:     code,
					Name:     code,
					ReadOnly: rw,
					MimeType: beMime.DirectoryMimeType,
				})
			}
		}

		for _, info := range infos {
			f.UpdateFileInfo(info, r)
			list = append(list, info)
		}
	}

	list = list.Sort()
	return
}

func (f *CFeature) ListLocaleFileSystems(r *http.Request) (list editor.Files) {
	eid := userbase.GetCurrentEID(r)
	for _, mpf := range f.EditingFileSystems {
		fsid := mpf.Tag().String()
		mountedPoints := mpf.GetMountedPoints()
		info := &editor.File{
			FSBT:     mpf.BaseTag().String(),
			FSID:     fsid,
			Name:     fsid,
			MimeType: beMime.DirectoryMimeType,
			ReadOnly: mountedPoints.HasRWFS(),
		}
		info.HasDraft = f.HasDraftLocales(fsid, "")
		if locked, lockedBy, ee := f.IsLocaleLocked(fsid, ""); ee == nil {
			if locked {
				info.LockedBy = lockedBy
				if info.Locked = lockedBy != eid; info.Locked {
					info.ReadOnly = true
				}
			}
		}
		f.UpdateFileInfo(info, r)
		list = append(list, info)
	}
	list = list.Sort()
	return
}

func (f *CFeature) ListLocaleFileSystemLocales(r *http.Request, fsid string) (list editor.Files) {
	eid := userbase.GetCurrentEID(r)
	unique := map[string]struct{}{}
	if found := f.FindFS(fsid); found != nil {
		for mount, mountPoints := range found.GetMountedPoints() {
			code := strings.TrimPrefix(mount, "/")
			unique[mount] = struct{}{}
			info := &editor.File{
				FSBT: found.BaseTag().String(),
				FSID: fsid,
				Code: code,
				Name: code,
				//Locale:   &language.Und, // above locales in path-space is Und
				MimeType: beMime.DirectoryMimeType,
				ReadOnly: !mountPoints.HasRWFS(),
			}
			info.HasDraft = f.HasDraftLocales(fsid, code)
			if locked, lockedBy, ee := f.IsLocaleLocked(fsid, code); ee == nil {
				if locked {
					info.LockedBy = lockedBy
					if info.Locked = lockedBy != eid; info.Locked {
						info.ReadOnly = true
					}
				}
			}
			f.UpdateFileInfo(info, r)
			list = append(list, info)
		}
	}
	list = list.Sort()
	return
}