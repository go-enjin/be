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

package site

import (
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/golang-org-x-text/language"
	"github.com/go-enjin/golang-org-x-text/message/catalog"

	"github.com/go-enjin/be/pkg/page"
)

type Enjin interface {
	SiteTag() (key string)
	SiteName() (name string)
	SiteTagLine() (tagLine string)
	SiteLocales() (locales []language.Tag)
	SiteLangCatalog() (c *lang.Catalog)
	SiteLanguageMode() (mode string)
	SiteLanguageCatalog() (c catalog.Catalog)
	SiteDefaultLanguage() (tag language.Tag)

	FindPage(tag language.Tag, url string) (p *page.Page)
}

func Info(e Enjin) (info map[string]interface{}) {
	info = map[string]interface{}{
		"Tag":         e.SiteTag(),
		"Name":        e.SiteName(),
		"TagLine":     e.SiteTagLine(),
		"Locales":     e.SiteLocales(),
		"LangMode":    e.SiteLanguageMode(),
		"DefaultLang": e.SiteDefaultLanguage(),
	}
	return
}