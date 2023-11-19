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
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/net/html"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/editor"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/slices"
	"github.com/go-enjin/be/pkg/userbase"
	"github.com/go-enjin/be/types/page/matter"
	"github.com/go-enjin/golang-org-x-text/language"
)

func (f *CFeature) ParseFormToDraft(pm *matter.PageMatter, fields context.Fields, form context.Context, info *editor.File, r *http.Request) (modified *matter.PageMatter, redirect string, errs map[string]error) {
	var err error
	eid := userbase.GetCurrentEID(r)
	printer := lang.GetPrinterFromRequest(r)

	if pm == nil {
		if pm, err = f.ReadDraftPage(info); err != nil {
			log.ErrorRF(r, "error encoding form context: %v", err)
			f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error encoding form context: "%[1]s"`, err.Error()))
			redirect = f.SelfEditor().GetEditorPath() + "/" + info.EditFilePath()
			return
		}
	}

	var formMatter context.Context
	if ctx, ok := form["matter"].(context.Context); ok {
		formMatter = ctx
	} else if fm, ok := form["matter"].(map[string]interface{}); ok {
		formMatter = fm
	} else {
		// nop
	}

	strict := bluemonday.StrictPolicy()
	var parseCustom func(v interface{}) (parsed interface{})
	parseCustom = func(v interface{}) (parsed interface{}) {
		if value, ok := v.(string); ok {
			value = strict.Sanitize(value)
			parsed = html.UnescapeString(value)
		} else if list, ok := v.([]interface{}); ok {
			var items []interface{}
			for _, item := range list {
				items = append(items, parseCustom(item))
			}
			parsed = items
		} else if dictionary, ok := v.(map[string]interface{}); ok {
			cleaned := make(map[string]interface{})
			for dk, dv := range dictionary {
				cleaned[dk] = parseCustom(dv)
			}
			parsed = cleaned
		}
		return
	}

	errs = make(map[string]error)

	for k, v := range formMatter.AsDeepKeyed() {
		if field, ok := fields.Lookup(k); ok {
			if field.Input != "checkbox" {
				if vi, ee := field.Parse(field, v); ee != nil {
					errs[field.Key] = ee
				} else {
					//field.Value = vi
					_ = maps.Set(k, vi, pm.Matter)
				}
			}
		} else {
			log.WarnRF(r, "strict policy for custom field: %q", k)
			_ = maps.Set(k, parseCustom(v), pm.Matter)
		}
	}

	for _, field := range fields {
		if field.Input == "checkbox" {
			var fv string
			fv = r.FormValue(".matter." + field.Key)
			_ = maps.Set("."+field.Key, fv != "", pm.Matter)
		}
	}

	if contents, ok := form.FirstString("body"); ok {
		pm.Body = strings.ReplaceAll(contents, "\r", "")
	} else {
		log.ErrorRF(r, "error decoding form body: %v", err)
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error decoding form body: "%v"`, err))
		redirect = f.SelfEditor().GetEditorPath() + "/" + info.EditFilePath()
		return
	}

	pm.Matter.KebabKeys()
	modified = pm

	return
}

func (f *CFeature) ParseCreatePageForm(r *http.Request, pg feature.Page, ctx, form context.Context, info *editor.File, eid string, redirect *string) (dstUri, dstArchetype string, dstInfo *editor.File, dstFS feature.FileSystemFeature, dstMP *feature.CMountPoint, dstExists bool, stop bool) {
	printer := lang.GetPrinterFromRequest(r)

	t := f.Enjin.MustGetTheme()
	archetypes := t.ListArchetypes()

	var err error
	var fileLocale language.Tag
	var fsid, fileLang, filePath, fileName, fileFormat, fullPath, dstPath, archetype string
	fsid, _ = form.FirstString(editor.CreatePageActionKey + "~dst-fsid")
	filePath, _ = form.FirstString(editor.CreatePageActionKey + "~dst-path")

	if fileLang, _ = form.FirstString(editor.CreatePageActionKey + "~dst-lang"); fileLang == "" {
		f.Editor.Site().PushWarnNotice(eid, true, printer.Sprintf(`a locale is required to create a new page`))
		stop = true
		return
	} else if fileLocale, err = language.Parse(fileLang); err != nil {
		f.Editor.Site().PushWarnNotice(eid, true, printer.Sprintf(`a valid locale is required to create a new page`))
		stop = true
		return
	} else if fileName, _ = form.FirstString(editor.CreatePageActionKey + "~dst-name"); fileName == "" {
		if stop = f.Emit(feature.FileNameRequiredSignal, f.Tag().String(), r, pg, ctx, form, info, eid, redirect); stop {
			return
		}
		f.Editor.Site().PushWarnNotice(eid, true, printer.Sprintf(`a file name is required`))
		*redirect = f.SelfEditor().GetEditorPath() + "/" + info.EditParentDirectoryPath()
		stop = true
		return
	}

	if fileFormat, _ = form.FirstString(editor.CreatePageActionKey + "~dst-format"); fileFormat == "" {
		f.Editor.Site().PushWarnNotice(eid, true, printer.Sprintf(`a file format is required to create a new page`))
		*redirect = f.SelfEditor().GetEditorPath() + "/" + info.EditParentDirectoryPath()
		stop = true
		return
	} else if _, matched := t.MatchFormat("not-a-thing." + fileFormat); matched == "" {
		f.Editor.Site().PushWarnNotice(eid, true, printer.Sprintf(`a valid file format is required to create a new page`))
		*redirect = f.SelfEditor().GetEditorPath() + "/" + info.EditParentDirectoryPath()
		stop = true
		return
	} else {
		fileFormat = matched
	}

	createMode := form.String(editor.CreatePageActionKey+"~create-mode", "page")
	if createMode == "arch" {
		if archetype, _ = form.FirstString(editor.CreatePageActionKey + "~dst-archetype"); archetype != "" && !slices.Within(archetype, archetypes) {
			archetypeNames := strings.Join(archetypes, ", ")
			f.Editor.Site().PushWarnNotice(eid, true, printer.Sprintf(`a supported archetype is required to create a new page, supported archetypes are: %[1]s`, archetypeNames))
			*redirect = f.SelfEditor().GetEditorPath() + "/" + info.EditParentDirectoryPath()
			stop = true
			return
		} else {
			if archetype != "" {
				var match string
				if _, match = t.MatchFormat(archetype); match == "" {
					f.Editor.Site().PushWarnNotice(eid, true, printer.Sprintf(`archetype page format not supported`))
					*redirect = f.SelfEditor().GetEditorPath() + "/" + info.EditParentDirectoryPath()
					return
				}
				fileFormat = match
				dstArchetype = archetype
			}
		}
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