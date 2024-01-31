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
	"net/http"

	"github.com/go-corelibs/x-text/language"
	"github.com/go-corelibs/x-text/message"
	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/editor"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
)

func (f *CFeature) ParseCreateMenuForm(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string, redirect *string) (dstUri string, dstInfo *editor.File, dstFS feature.FileSystemFeature, dstMP *feature.CMountPoint, dstExists bool, stop bool) {
	printer := message.GetPrinter(r)

	var err error
	var fileLocale language.Tag
	var fsid, fileLang, filePath, fileName, fileFormat, fullPath, dstPath string
	fsid, _ = form.FirstString(editor.CreateMenuActionKey + "~dst-fsid")
	filePath, _ = form.FirstString(editor.CreateMenuActionKey + "~dst-path")

	if fileLang, _ = form.FirstString(editor.CreateMenuActionKey + "~dst-lang"); fileLang == "" {
		f.Editor.Site().PushWarnNotice(eid, true, printer.Sprintf(`a locale is required to create a new page`))
		stop = true
		return
	} else if fileLocale, err = language.Parse(fileLang); err != nil {
		f.Editor.Site().PushWarnNotice(eid, true, printer.Sprintf(`a valid locale is required to create a new page`))
		stop = true
		return
	} else if fileName, _ = form.FirstString(editor.CreateMenuActionKey + "~dst-name"); fileName == "" {
		if stop = f.Emit(feature.FileNameRequiredSignal, f.Tag().String(), r, pg, ctx, form, info, eid, redirect); stop {
			return
		}
		f.Editor.Site().PushWarnNotice(eid, true, printer.Sprintf(`a file name is required`))
		*redirect = f.SelfEditor().GetEditorPath() + "/" + info.EditParentDirectoryPath()
		stop = true
		return
	} else {
		fileFormat = "json"
	}

	fsid = forms.StrictCleanKebabValue(fsid)
	fileName = forms.StrictCleanKebabValue(fileName)
	if filePath = forms.KebabRelativePath(filePath); filePath != "" {
		fullPath = filePath + "/" + fileName + "." + fileFormat
	} else {
		fullPath = fileName + "." + fileFormat
	}

	dstPath = fileLocale.String() + "/" + fullPath
	dstUri = fsid + "://" + dstPath
	dstInfo = editor.ParseFile(fsid, dstPath)

	for _, efs := range f.EditingFileSystems {
		if efs.Tag().String() == dstInfo.FSID {
			dstFS = efs
			for _, mps := range efs.GetMountedPoints() {
				for _, mp := range mps {
					// TODO: figure out mount point prefix
					if dstExists = mp.ROFS.Exists(dstInfo.FilePath()); mp.RWFS != nil {
						dstMP = mp
						break
					}
				}
				if dstExists || dstMP != nil {
					break
				}
			}
		}
	}
	return
}
