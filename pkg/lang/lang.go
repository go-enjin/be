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
	"net/http"
	"strings"

	"github.com/go-enjin/golang-org-x-text/language"
	"github.com/go-enjin/golang-org-x-text/message"
	"github.com/go-enjin/golang-org-x-text/message/catalog"

	"github.com/go-enjin/be/pkg/log"
)

type ContextKey string

const LanguageTag ContextKey = "language-tag"
const LanguagePrinter ContextKey = "language-printer"
const LanguageDefault ContextKey = "language-default"

func ParseLangPath(p string) (tag language.Tag, modified string, ok bool) {
	modified = p
	var path string
	if p == "" {
		return
	} else if path = p; path[0] == '/' {
		path = path[1:]
	}
	if parts := strings.Split(path, "/"); len(parts) >= 2 {
		if t, err := language.Parse(parts[0]); err == nil {
			tag = t
			modified = "/" + strings.Join(parts[1:], "/")
			ok = true
		}
	}
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
		log.ErrorDF(1, "request missing language printer, defaulting to Undefined")
		printer = message.NewPrinter(language.Und)
	}
	return
}