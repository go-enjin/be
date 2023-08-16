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

package feature

import (
	"github.com/go-enjin/golang-org-x-text/language"
	catalog2 "github.com/go-enjin/golang-org-x-text/message/catalog"

	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/lang/catalog"
	"github.com/go-enjin/be/pkg/page"
)

type SiteEnjin interface {
	SiteTag() (key string)
	SiteName() (name string)
	SiteTagLine() (tagLine string)
	SiteLocales() (locales []language.Tag)
	SiteLangCatalog() (c *catalog.Catalog)
	SiteLanguageMode() (mode lang.Mode)
	SiteLanguageCatalog() (c catalog2.Catalog)
	SiteDefaultLanguage() (tag language.Tag)
	SiteSupportsLanguage(tag language.Tag) (supported bool)
	SiteLanguageDisplayName(tag language.Tag) (name string, ok bool)

	FindTranslations(url string) (pages []*page.Page)
	FindPage(tag language.Tag, url string) (p *page.Page)
	FindPages(prefix string) (pages []*page.Page)
}

type SiteInfo struct {
	Tag         string
	Name        string
	TagLine     string
	Locales     []language.Tag
	LangMode    lang.Mode
	DefaultLang language.Tag
}

func MakeSiteInfo(e SiteEnjin) (info SiteInfo) {
	info = SiteInfo{
		Tag:         e.SiteTag(),
		Name:        e.SiteName(),
		TagLine:     e.SiteTagLine(),
		Locales:     e.SiteLocales(),
		LangMode:    e.SiteLanguageMode(),
		DefaultLang: e.SiteDefaultLanguage(),
	}
	return
}