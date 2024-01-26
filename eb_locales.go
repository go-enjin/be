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

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/lang"
)

func (eb *EnjinBuilder) SiteLanguageMode(mode lang.Mode) feature.Builder {
	eb.langMode = mode
	return eb
}

func (eb *EnjinBuilder) SiteDefaultLanguage(tag language.Tag) feature.Builder {
	eb.defaultLang = tag
	return eb
}

func (eb *EnjinBuilder) SiteSupportedLanguages(tags ...language.Tag) feature.Builder {
	eb.localeTags = append(eb.localeTags, tags...)
	return eb
}

func (eb *EnjinBuilder) SiteLanguageDisplayNames(names map[language.Tag]string) feature.Builder {
	if eb.localeNames == nil {
		eb.localeNames = make(map[language.Tag]string)
	}
	for tag, name := range names {
		eb.localeNames[tag] = name
	}
	return eb
}
