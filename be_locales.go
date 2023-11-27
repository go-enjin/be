// Copyright (c) 2022  The Go-Enjin Authors
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

package be

import (
	"github.com/go-enjin/golang-org-x-text/language"
	"github.com/go-enjin/golang-org-x-text/language/display"
	"github.com/go-enjin/golang-org-x-text/message/catalog"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/signals"

	"github.com/go-enjin/be/pkg/lang"
	pkgLangCatalog "github.com/go-enjin/be/pkg/lang/catalog"
	"github.com/go-enjin/be/pkg/log"
)

func (e *Enjin) ReloadLocales() {
	e.Emit(signals.PreEnjinReloadLocales, feature.EnjinTag.String(), interface{}(e).(feature.Internals))

	ctlg := pkgLangCatalog.New()
	// include locale features
	for _, lp := range e.eb.fLocalesProviders {
		//log.DebugF("adding %v locales", lp.Tag())
		lp.AddLocales(ctlg)
	}
	e.mutex.Lock()
	e.catalog = ctlg
	e.locales = ctlg.LocaleTagsWithDefault(e.eb.defaultLang)
	e.mutex.Unlock()
	e.Emit(signals.PostEnjinReloadLocales, feature.EnjinTag.String(), interface{}(e).(feature.Internals))
}

func (e *Enjin) SiteLocales() (locales lang.Tags) {
	if len(e.eb.localeTags) == 0 {
		e.mutex.RLock()
		defer e.mutex.RUnlock()
		locales = e.locales
		return
	}
	locales = append(locales, e.eb.localeTags...)
	return
}

func (e *Enjin) SiteLanguageMode() (mode lang.Mode) {
	//e.mutex.RLock()
	//defer e.mutex.RUnlock()
	mode = e.eb.langMode
	return
}

func (e *Enjin) SiteLanguageCatalog() (c catalog.Catalog) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	if v, err := e.catalog.MakeGoTextCatalog(); err == nil {
		c = v
	} else {
		log.ErrorF("error making gotext catalog: %v", err)
		c = catalog.NewBuilder() // always return a valid catalog
	}
	return
}

func (e *Enjin) SiteDefaultLanguage() (tag language.Tag) {
	//e.mutex.RLock()
	//defer e.mutex.RUnlock()
	tag = e.eb.defaultLang
	return
}

func (e *Enjin) SiteSupportsLanguage(tag language.Tag) (supported bool) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	for _, known := range e.SiteLocales() {
		if supported = language.Compare(tag, known); supported {
			break
		}
	}
	return
}

func (e *Enjin) SiteLanguageDisplayName(tag language.Tag) (name string, ok bool) {
	//e.mutex.RLock()
	//defer e.mutex.RUnlock()
	if len(e.eb.localeNames) > 0 {
		name, ok = e.eb.localeNames[tag]
	} else {
		name = display.Self.Name(tag)
		ok = name != ""
	}
	return
}
