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
	"github.com/go-enjin/golang-org-x-text/message/catalog"

	"github.com/go-enjin/be/pkg/lang"
	pkgLangCatalog "github.com/go-enjin/be/pkg/lang/catalog"
	"github.com/go-enjin/be/pkg/log"
)

func (e *Enjin) initLocales() {
	e.catalog = pkgLangCatalog.NewCatalog()
	for _, f := range e.eb.localeFiles {
		e.catalog.AddLocalesFromFS(e.eb.defaultLang, f)
	}
}

func (e *Enjin) SiteLocales() (locales []language.Tag) {
	if len(e.eb.localeTags) == 0 {
		locales = e.catalog.LocaleTagsWithDefault(e.eb.defaultLang)
		return
	}
	locales = append(locales, e.eb.localeTags...)
	return
}

func (e *Enjin) SiteLanguageMode() (mode lang.Mode) {
	mode = e.eb.langMode
	return
}

func (e *Enjin) SiteLangCatalog() (c *pkgLangCatalog.Catalog) {
	c = e.catalog
	return
}

func (e *Enjin) SiteLanguageCatalog() (c catalog.Catalog) {
	if v, err := e.catalog.MakeGoTextCatalog(); err == nil {
		c = v
	} else {
		log.ErrorF("error making gotext catalog: %v", err)
		c = catalog.NewBuilder() // always return a valid catalog
	}
	return
}

func (e *Enjin) SiteDefaultLanguage() (tag language.Tag) {
	tag = e.eb.defaultLang
	return
}

func (e *Enjin) SiteSupportsLanguage(tag language.Tag) (supported bool) {
	for _, known := range e.SiteLocales() {
		if supported = language.Compare(tag, known); supported {
			break
		}
	}
	return
}

func (e *Enjin) SiteLanguageDisplayName(tag language.Tag) (name string, ok bool) {
	if len(e.eb.localeNames) > 0 {
		name, ok = e.eb.localeNames[tag]
	}
	return
}