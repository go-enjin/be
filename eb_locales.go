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
	"embed"
	"strings"

	"github.com/go-enjin/be/pkg/feature"
	embed2 "github.com/go-enjin/be/pkg/fs/embed"
	"github.com/go-enjin/be/pkg/fs/local"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/golang-org-x-text/language"
)

func (eb *EnjinBuilder) SiteLanguageMode(mode string) feature.Builder {
	check := strings.ToLower(mode)
	switch check {
	case "domain":
		log.FatalDF(1, "domain language mode not implemented yet")
	case "path", "query":
		eb.langMode = mode
	default:
		log.FatalDF(1, "invalid site language mode: %v", mode)
	}
	return eb
}

func (eb *EnjinBuilder) SiteDefaultLanguage(tag language.Tag) feature.Builder {
	eb.defaultLang = tag
	return eb
}

func (eb *EnjinBuilder) AddLocalesLocalFS(path string) feature.Builder {
	if f, err := local.New(path); err == nil {
		eb.localeFiles = append(eb.localeFiles, f)
	}
	return eb
}

func (eb *EnjinBuilder) AddLocalesEmbedFS(path string, efs embed.FS) feature.Builder {
	if f, err := embed2.New(path, efs); err == nil {
		eb.localeFiles = append(eb.localeFiles, f)
	}
	return eb
}