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
	"os"
	"strings"

	"github.com/go-enjin/be/pkg/editor"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/types/page/matter"
	"github.com/go-enjin/golang-org-x-text/language"
)

func (f *CFeature) ReadDraftPage(info *editor.File) (pm *matter.PageMatter, err error) {
	var data []byte
	if info.HasDraft {
		if data, err = f.SelfEditor().ReadDraft(info); err != nil {
			return
		}
	} else if data, err = f.SelfEditor().ReadFile(info); err != nil {
		return
	}
	pm, err = matter.ParsePageMatter(info.FSID, info.FilePath(), info.Created, info.Updated, data)
	return
}

func (f *CFeature) WriteDraftPage(info *editor.File, pm *matter.PageMatter) (err error) {
	var data []byte
	if data, err = pm.Bytes(); err != nil {
		return
	}
	err = f.SelfEditor().WriteDraft(info, data)
	return
}

func (f *CFeature) PublishDraftPage(info *editor.File) (err error) {

	if f.SelfEditor().DraftExists(info) {

		var pm *matter.PageMatter
		if _, pm, err = f.InfoRenderCheck(info); err != nil {
			return
		}

		f.RemoveIndexing(info)

		pm.Matter = pm.Matter.PruneEmpty()
		if err = f.WritePage(info, pm); err != nil {
			return
		}

		f.AddIndexing(info)
	}
	return
}

func (f *CFeature) ReadPageMatter(info *editor.File) (pm *matter.PageMatter, err error) {
	filePath := info.FilePath()
	for _, mpf := range f.EditingFileSystems {
		mpfTag := mpf.Tag().String()
		if mpfTag == info.FSID {
			for _, mountPoints := range mpf.GetMountedPoints() {
				for _, mountPoint := range mountPoints {
					if !strings.HasPrefix("/"+info.FilePath(), mountPoint.Mount) {
						continue
					} else if mountPoint.ROFS.Exists(filePath) {
						if pm, err = mountPoint.ROFS.ReadPageMatter(filePath); err == nil {
							pm.Stub = feature.NewPageStub(mpfTag, f.Enjin.Context(), mountPoint.ROFS, mountPoint.Mount, filePath, pm.Shasum, *info.Locale)
							pm.Locale = *info.Locale
							return
						}
					}
				}
			}
		}
	}
	err = os.ErrNotExist
	return
}

func (f *CFeature) WritePage(info *editor.File, pm *matter.PageMatter) (err error) {
	pm.Matter.Delete("~")
	var data []byte
	if data, err = pm.Bytes(); err != nil {
		return
	}
	err = f.SelfEditor().WriteFile(info, data)
	return
}

func (f *CFeature) RemovePage(info *editor.File, pm *matter.PageMatter) (err error) {

	f.RemoveIndexing(info)

	if err = f.SelfEditor().RemoveFile(info); err != nil {
		return
	}

	return
}

func (f *CFeature) UpdatePathInfo(info *editor.File, r *http.Request) {
	// page-level actions (floating bottom-right button menu actions)
	printer := lang.GetPrinterFromRequest(r)

	if !info.ReadOnly {
		info.Actions = append(info.Actions, editor.MakeCreatePageAction(printer))
	}

	return
}

func (f *CFeature) UpdateFileInfo(info *editor.File, r *http.Request) {
	// browser row actions and editing page
	f.CEditorFeature.UpdateFileInfo(info, r)
	printer := lang.GetPrinterFromRequest(r)

	if info.HasDraft {
		info.Actions = append(info.Actions, editor.MakePreviewDraftAction(printer))
	} else if !info.Locked {
		if f.HasIndexing(info) {
			info.Actions = append(info.Actions, editor.MakeDeIndexPageAction(printer, info.FilePath()))
		} else {
			info.Actions = append(info.Actions, editor.MakeIndexPageAction(printer, info.FilePath()))
		}
	}

	info.Actions = info.Actions.Sort()
}

func (f *CFeature) UpdateFileInfoForEditing(info *editor.File, r *http.Request) {
	// only on the editing page
	f.CEditorFeature.UpdateFileInfoForEditing(info, r)
	printer := lang.GetPrinterFromRequest(r)

	if untranslated := f.GetUntranslatedLocales(info); len(untranslated) > 0 {
		info.Actions = append(info.Actions, editor.MakeTranslateAction(printer, info.File))
	}

	info.Actions = info.Actions.Sort()
	return
}

func (f *CFeature) GetTranslatedLocales(info *editor.File) (translations map[language.Tag]string) {
	translations = map[language.Tag]string{}
	if url := info.Url(); url != "" {
		txs := f.Enjin.FindTranslations(url)
		if dtag := f.Enjin.SiteDefaultLanguage(); dtag == *info.Locale {
			// is default locale, find translations
			for _, p := range txs {
				if tag := p.LanguageTag(); tag != dtag {
					pm := p.PageMatter()
					translations[tag] = pm.Origin + "/" + pm.Locale.String() + pm.Path
				}
			}
		} else {
			// is translation, find default locale
			for _, p := range txs {
				if tag := p.LanguageTag(); tag == dtag {
					pm := p.PageMatter()
					translations[tag] = pm.Origin + "/" + pm.Locale.String() + pm.Path
					break
				}
			}
		}
	}
	return
}

func (f *CFeature) GetUntranslatedLocales(info *editor.File) (locales []language.Tag) {

	if url := info.Url(); url != "" {
		txs := f.Enjin.FindTranslationUrls(url)
		for _, locale := range f.Enjin.SiteLocales() {
			if _, present := txs[locale]; !present {
				locales = append(locales, locale)
			}
		}
	}

	return
}