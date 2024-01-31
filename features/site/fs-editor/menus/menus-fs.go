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
	"github.com/go-enjin/be/pkg/editor"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/menu"
)

func (f *CFeature) UpdatePathInfo(info *editor.File, r *http.Request) {
	// page-level actions (floating bottom-right button menu actions)
	printer := message.GetPrinter(r)

	if !info.ReadOnly {
		info.Actions = append(info.Actions, editor.MakeCreateMenuAction(printer))
	}

	return
}

func (f *CFeature) UpdateFileInfo(info *editor.File, r *http.Request) {
	f.CEditorFeature.UpdateFileInfo(info, r)
	t := f.Enjin.MustGetTheme()
	supported := t.GetConfig().Supports.Menus
	printer := message.GetPrinter(r)
	if info.Path != "" {
		info.Indicators = append(info.Indicators, &editor.Indicator{
			Icon:    "danger fa-solid fa-circle-xmark",
			Message: printer.Sprintf(`%[1]s ignores all menus in sub-directories`, t.Name()),
		})
	} else if supported.Has(info.BaseName()) {
		info.Indicators = append(info.Indicators, &editor.Indicator{
			Icon:    "important fa-solid fa-circle-check",
			Message: printer.Sprintf(`%[1]s renders this menu`, t.Name()),
		})
	} else {
		info.Indicators = append(info.Indicators, &editor.Indicator{
			Icon:    "caution fa-solid fa-circle-xmark",
			Message: printer.Sprintf(`%[1]s ignores this menu`, t.Name()),
		})
	}
	if info.HasDraft {
		info.Actions = append(info.Actions, editor.MakePreviewDraftAction(printer))
	}
	info.Actions = info.Actions.Sort()
	return
}

func (f *CFeature) UpdateFileInfoForEditing(info *editor.File, r *http.Request) {
	// only on the editing page
	f.CEditorFeature.UpdateFileInfoForEditing(info, r)
	printer := message.GetPrinter(r)

	if untranslated := f.GetUntranslatedLocales(info); len(untranslated) > 0 {
		info.Actions = append(info.Actions, editor.MakeTranslateAction(printer, info.File))
	}

	info.Actions = info.Actions.Sort()
	return
}

func (f *CFeature) GetAllMenus() (allMenus map[language.Tag]map[string]menu.Menu) {
	allMenus = map[language.Tag]map[string]menu.Menu{}
	for _, mp := range f.Enjin.GetMenuProviders() {
		for locale, m := range mp.GetAllMenus() {
			maps.MakeTypedKey(locale, allMenus)
			for name, mm := range m {
				allMenus[locale][name] = mm
			}
		}
	}
	return
}

func (f *CFeature) GetTranslatedLocales(info *editor.File) (txs map[language.Tag]string) {
	txs = map[language.Tag]string{}
	_, menuPath, _ := lang.ParseLangPath(info.FilePath())
	defaultLocale := f.Enjin.SiteDefaultLanguage()
	isDefaultLocale := defaultLocale == *info.Locale
	for _, ef := range f.EditingFileSystems {
		for _, mountedPoint := range ef.GetMountedPoints() {
			for _, mountPoint := range mountedPoint {
				if files, err := mountPoint.ROFS.ListAllFiles("."); err == nil {
					for _, file := range files {
						if _, _, ok := editor.ParseEditorWorkFile(file); ok {
							continue
						}
						if tag, modified, ok := lang.ParseLangPath(file); ok {
							if isDefaultLocale {
								// is default locale, find other translations
								if modified == menuPath && tag != *info.Locale {
									txs[tag] = ef.Tag().String() + "/" + file
								}
							} else {
								// is translation, find default locale
								if modified == menuPath && tag == defaultLocale {
									txs[tag] = ef.Tag().String() + "/" + file
								}
							}
						}
					}
				}
			}
		}
	}
	return
}

func (f *CFeature) GetUntranslatedLocales(info *editor.File) (locales []language.Tag) {
	_, menuPath, _ := lang.ParseLangPath(info.FilePath())
	translated := map[language.Tag]struct{}{}
	for _, ef := range f.EditingFileSystems {
		for _, mountedPoint := range ef.GetMountedPoints() {
			for _, mountPoint := range mountedPoint {
				if files, err := mountPoint.ROFS.ListAllFiles("."); err == nil {
					for _, file := range files {
						if _, _, ok := editor.ParseEditorWorkFile(file); ok {
							continue
						}
						if tag, modified, ok := lang.ParseLangPath(file); ok {
							if modified == menuPath {
								translated[tag] = struct{}{}
							}
						}
					}
				}
			}
		}
	}
	for _, tag := range f.Enjin.SiteLocales() {
		if _, present := translated[tag]; !present {
			locales = append(locales, tag)
		}
	}

	return
}
