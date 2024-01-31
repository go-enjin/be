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
	"github.com/go-corelibs/x-text/language"
	"github.com/go-corelibs/x-text/message"

	"github.com/go-enjin/be/pkg/lang"
)

const (
	EnjinTag        Tag = "enjin"
	EnjinLocalesTag Tag = "enjin-locales"
)

type EnjinTextFn func(printer *message.Printer) (text EnjinText)

type EnjinText struct {
	Name            string
	TagLine         string
	CopyrightName   string
	CopyrightYear   string
	CopyrightNotice string
}

type EnjinInfo struct {
	Tag string
	EnjinText
	Locales     []language.Tag
	LangMode    lang.Mode
	DefaultLang language.Tag
}

func MakeEnjinInfo(e EnjinBase) (info EnjinInfo) {
	info = EnjinInfo{
		Tag: e.SiteTag(),
		EnjinText: EnjinText{
			Name:            e.SiteName(),
			TagLine:         e.SiteTagLine(),
			CopyrightName:   e.SiteCopyrightName(),
			CopyrightYear:   e.SiteCopyrightYear(),
			CopyrightNotice: e.SiteCopyrightNotice(),
		},
		Locales:     e.SiteLocales(),
		LangMode:    e.SiteLanguageMode(),
		DefaultLang: e.SiteDefaultLanguage(),
	}
	return
}
