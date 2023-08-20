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

package lang

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-enjin/golang-org-x-text/language"
	"github.com/go-enjin/golang-org-x-text/message"
	"github.com/go-enjin/golang-org-x-text/message/catalog"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/log"
	bePath "github.com/go-enjin/be/pkg/path"
)

const (
	LanguageTag     beContext.RequestKey = "language-tag"
	LanguagePrinter beContext.RequestKey = "language-printer"
	LanguageDefault beContext.RequestKey = "language-default"
	PrinterKey      string               = "LangPrinter"
)

func ParseLangPath(p string) (tag language.Tag, modified string, ok bool) {
	modified = p
	lead := "/"
	var path string
	if p == "" {
		return
	} else if path = p; path[0] == '/' {
		path = path[1:]
	} else if path[0] == '!' {
		lead = "!"
		path = path[1:]
	}
	if parts := strings.Split(path, "/"); len(parts) >= 2 {
		if t, err := language.Parse(parts[0]); err == nil {
			tag = t
			modified = lead + strings.Join(parts[1:], "/")
			ok = true
		}
	}
	return
}

func SetTag(r *http.Request, tag language.Tag) (modified *http.Request) {
	modified = r.Clone(context.WithValue(r.Context(), LanguageTag, tag))
	return
}

func GetTag(r *http.Request) (tag language.Tag) {
	if p, ok := r.Context().Value(LanguageTag).(language.Tag); ok {
		tag = p
	} else if p, ok := r.Context().Value(LanguageDefault).(language.Tag); ok {
		tag = p
	} else {
		tag = language.Und
		log.ErrorDF(1, "request missing language tag and default language, defaulting to Undefined")
	}
	return
}

func NewCatalogPrinter(lang string, c catalog.Catalog) (tag language.Tag, printer *message.Printer) {
	var err error
	if tag, err = language.Parse(lang); err != nil {
		tag = language.Und
		printer = message.NewPrinter(language.Und, message.Catalog(c))
		log.ErrorDF(1, "error parsing language: %v - %v, defaulting to Undefined", lang, err)
	} else {
		printer = message.NewPrinter(tag, message.Catalog(c))
	}
	return
}

func GetPrinterFromRequest(r *http.Request) (printer *message.Printer) {
	if p, ok := r.Context().Value(LanguagePrinter).(*message.Printer); ok {
		printer = p
	} else if tag, ok := r.Context().Value(LanguageDefault).(language.Tag); ok {
		printer = message.NewPrinter(tag)
	} else {
		log.TraceDF(1, "request missing language printer, defaulting to Undefined")
		printer = message.NewPrinter(language.Und)
	}
	return
}

func TagInTags(needle language.Tag, haystack ...language.Tag) (found bool) {
	for _, tag := range haystack {
		if found = language.Compare(needle, tag); found {
			return
		}
	}
	return
}

func TagInTagSlices(needle language.Tag, haystacks ...[]language.Tag) (found bool) {
	for _, haystack := range haystacks {
		for _, tag := range haystack {
			if found = language.Compare(needle, tag); found {
				return
			}
		}
	}
	return
}

var rxTranslatorInlineComments = regexp.MustCompile(`(?ms)\((\s*_\s+.+?\s*)/\*.+?\*/\s*\)`)
var rxTranslatorPipelineComments = regexp.MustCompile(`(?ms)\{\{(-??\s*_\s+.+?\s*)/\*.+?\*/(\s*-??)}}`)

func StripTranslatorComments(raw string) (clean string) {
	clean = rxTranslatorInlineComments.ReplaceAllString(raw, `(${1})`)
	clean = rxTranslatorPipelineComments.ReplaceAllString(clean, `{{${1}${2}}}`)
	return
}

func NonPageRequested(r *http.Request) (is bool) {
	path := forms.TrimQueryParams(r.URL.Path)
	is = bePath.Ext(path) != ""
	return
}