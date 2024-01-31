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

package feature

import (
	"net/http"

	"github.com/go-corelibs/x-text/language"

	"github.com/go-enjin/be/pkg/context"
)

type PageContextFilterFn = func(ctx context.Context, r *http.Request) (modCtx context.Context)

type PageContextModifier interface {
	Feature
	FilterPageContext(themeCtx, pageCtx context.Context, r *http.Request) (themeOut context.Context)
}

type PageContextUpdater interface {
	Feature
	UpdatePageContext(pageCtx context.Context, r *http.Request) (additions context.Context)
}

type EnjinContextProvider interface {
	Feature
	EnjinContext(r *http.Request) (ctx context.Context)
}

type PageRestrictionHandler interface {
	Feature
	RestrictServePage(ctx context.Context, w http.ResponseWriter, r *http.Request) (modCtx context.Context, modReq *http.Request, allow bool)
}

type DataRestrictionHandler interface {
	Feature
	RestrictServeData(data []byte, mime string, w http.ResponseWriter, r *http.Request) (modReq *http.Request, allow bool)
}

type FileProvider interface {
	Feature
	FindFile(url string) (data []byte, mime string, err error)
}

type PageContextParsersProvider interface {
	Feature
	PageContextParsers() (fields context.Parsers)
}

type PageContextFieldsProvider interface {
	Feature
	MakePageContextFields(r *http.Request) (fields context.Fields)
}

type PageProvider interface {
	Feature
	FindRedirection(url string) (p Page)
	FindTranslations(url string) (pages []Page)
	FindTranslationUrls(url string) (pages map[language.Tag]string)
	FindPage(r *http.Request, tag language.Tag, url string) (p Page)
	LookupPrefixed(prefix string) (pages []Page)
}

type PageTypeProcessor interface {
	Feature
	PageTypeNames() (names []string)
	ProcessRequestPageType(r *http.Request, p Page) (pg Page, redirect string, processed bool, err error)
}

type PageShortcodeProcessor interface {
	Feature
	TranslateShortcodes(content string, ctx context.Context) (modified string)
}

type TemplatePartialsProvider interface {
	Feature
	RegisterTemplatePartial(block, position, name, tmpl string) (err error)
	ListTemplatePartials(block, position string) (names []string)
	GetTemplatePartial(block, position, name string) (tmpl string, ok bool)
}
