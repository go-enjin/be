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

package theme

import (
	"fmt"
	"html/template"
	"sync"
	textTemplate "text/template"

	"github.com/go-enjin/golang-org-x-text/language"
	"github.com/go-enjin/golang-org-x-text/language/display"
	"github.com/go-enjin/golang-org-x-text/message"
	"github.com/iancoleman/strcase"
	"github.com/leekchan/gtf"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/page"
	"github.com/go-enjin/be/pkg/theme/funcs"
	"github.com/go-enjin/be/pkg/types/site"
)

var (
	_knownFuncMap   = make(map[string]interface{})
	_knownFuncMutex = sync.RWMutex{}
)

func RegisterFuncMap(key string, fn interface{}) {
	_knownFuncMutex.Lock()
	defer _knownFuncMutex.Unlock()
	_knownFuncMap[key] = fn
}

func GetFuncMap(key string) (fn interface{}, ok bool) {
	_knownFuncMutex.RLock()
	defer _knownFuncMutex.RUnlock()
	fn, ok = _knownFuncMap[key]
	return
}

func AddRegisteredHtmlFuncsToMap(fm *template.FuncMap) {
	_knownFuncMutex.RLock()
	defer _knownFuncMutex.RUnlock()
	for key, fn := range _knownFuncMap {
		(*fm)[key] = fn
	}
	return
}

func AddRegisteredTextFuncsToMap(fm *textTemplate.FuncMap) {
	_knownFuncMutex.RLock()
	defer _knownFuncMutex.RUnlock()
	for key, fn := range _knownFuncMap {
		(*fm)[key] = fn
	}
	return
}

func DefaultFuncMap() (funcMap template.FuncMap) {
	funcMap = template.FuncMap{
		"toCamel":              strcase.ToCamel,
		"toLowerCamel":         strcase.ToLowerCamel,
		"toDelimited":          strcase.ToDelimited,
		"toScreamingDelimited": strcase.ToScreamingDelimited,
		"toKebab":              strcase.ToKebab,
		"toScreamingKebab":     strcase.ToScreamingKebab,
		"toSnake":              strcase.ToSnake,
		"toScreamingSnake":     strcase.ToScreamingSnake,

		"asHTML":     funcs.AsHTML,
		"asHTMLAttr": funcs.AsHTMLAttr,
		"asCSS":      funcs.AsCSS,
		"asJS":       funcs.AsJS,

		"fsHash":   funcs.FsHash,
		"fsUrl":    funcs.FsUrl,
		"fsMime":   funcs.FsMime,
		"fsExists": funcs.FsExists,

		"add":      funcs.Add,
		"sub":      funcs.Sub,
		"mul":      funcs.Mul,
		"div":      funcs.Div,
		"mod":      funcs.Mod,
		"addFloat": funcs.AddFloat,
		"subFloat": funcs.SubFloat,
		"mulFloat": funcs.MulFloat,
		"divFloat": funcs.DivFloat,

		"mergeClassNames": funcs.MergeClassNames,

		"unescapeHTML":     funcs.UnescapeHtml,
		"escapeJsonString": funcs.EscapeJsonString,
		"escapeHTML":       funcs.EscapeHtml,
		"escapeUrlPath":    funcs.EscapeUrlPath,

		"element":           funcs.Element,
		"elementOpen":       funcs.ElementOpen,
		"elementClose":      funcs.ElementClose,
		"elementAttributes": funcs.ElementAttributes,

		"Nonce": funcs.Nonce,

		"isUrl":    funcs.IsUrl,
		"isPath":   funcs.IsPath,
		"parseUrl": funcs.ParseUrl,

		"sortedKeys": funcs.SortedKeys,

		"cmpDateFmt": funcs.CompareDateFormats,

		"DebugF": funcs.LogDebug,
		"WarnF":  funcs.LogWarn,
		"ErrorF": funcs.LogError,

		"CmpLang": funcs.CmpLang,
	}
	for k, v := range gtf.GtfFuncMap {
		funcMap[k] = v
	}
	AddRegisteredHtmlFuncsToMap(&funcMap)
	return
}

func (t *Theme) NewTextFuncMapWithContext(ctx context.Context) (fm textTemplate.FuncMap) {

	fm = textTemplate.FuncMap{}
	for k, v := range t.FuncMap {
		fm[k] = v
	}
	AddRegisteredTextFuncsToMap(&fm)

	fm["_"] = t.makeUnderscore(ctx)            // translate page content
	fm["__"] = t.makeUnderscoreUnderscore(ctx) // translate page paths
	fm["_tag"] = t.makeUnderscoreTag(ctx)      // render language tag in native language
	return
}

func (t *Theme) NewHtmlFuncMapWithContext(ctx context.Context) (fm template.FuncMap) {

	fm = template.FuncMap{}
	for k, v := range t.FuncMap {
		fm[k] = v
	}
	AddRegisteredHtmlFuncsToMap(&fm)

	fm["_"] = t.makeUnderscore(ctx)            // translate page content
	fm["__"] = t.makeUnderscoreUnderscore(ctx) // translate page paths
	fm["_tag"] = t.makeUnderscoreTag(ctx)      // render language tag in native language
	return
}

func (t *Theme) makeUnderscoreTag(ctx context.Context) func(tag language.Tag) (name string) {
	return func(tag language.Tag) (name string) {
		enjin, _ := ctx.Get("SiteEnjin").(site.Enjin)
		var ok bool
		if name, ok = enjin.SiteLanguageDisplayName(tag); !ok {
			name = display.Tags(tag).Name(tag)
		}
		return
	}
}

func (t *Theme) makeUnderscore(ctx context.Context) func(format string, argv ...interface{}) (translated string) {
	return func(format string, argv ...interface{}) (translated string) {
		printer, _ := ctx.Get("LangPrinter").(*message.Printer)
		if printer != nil {
			translated = printer.Sprintf(format, argv...)
			if fmt.Sprintf(format, argv...) != translated {
				log.DebugF("template translated: \"%v\" -> \"%v\"", format, translated)
			}
		} else {
			translated = fmt.Sprintf(format, argv...)
		}
		return
	}
}

func (t *Theme) makeUnderscoreUnderscore(ctx context.Context) func(argv ...string) (translated string, err error) {
	return func(argv ...string) (translated string, err error) {
		enjin, _ := ctx.Get("SiteEnjin").(site.Enjin)
		targetLang, _ := ctx.Get("ReqLangTag").(language.Tag)
		var targetPath, fallbackPath string

		switch len(argv) {
		case 0:
			err = fmt.Errorf("called with no arguments")
			return
		case 1:
			targetPath = argv[0]
		case 2:
			if targetLang, err = language.Parse(argv[0]); err != nil {
				err = fmt.Errorf("called with invalid language: %v", argv[0])
				return
			}
			targetPath = argv[1]
		case 3:
			if targetLang, err = language.Parse(argv[0]); err != nil {
				err = fmt.Errorf("called with invalid language: %v", argv[0])
				return
			}
			fallbackPath = argv[1]
			targetPath = argv[2]
		default:
			err = fmt.Errorf("called with too many arguments")
			return
		}

		if targetPath == "" || targetPath[0] != '/' {
			translated = targetPath
			return
		}

		if !enjin.SiteSupportsLanguage(targetLang) {
			log.ErrorF("unsupported site language requested: %v", targetLang)
			targetLang = enjin.SiteDefaultLanguage()
		}

		var targetPage *page.Page
		if targetPage = enjin.FindPage(targetLang, targetPath); targetPage == nil {
			if found := enjin.FindTranslations(targetPath); len(found) > 0 {
				for _, pg := range found {
					if pg.IsTranslation(targetPath) {
						if language.Compare(pg.LanguageTag, targetLang) {
							targetPage = pg
							break
						}
					} else {
						targetPage = enjin.FindPage(targetLang, pg.Translates)
						break
					}
				}
			}

			if targetPage == nil {
				if targetPage = enjin.FindPage(language.Und, targetPath); targetPage == nil {
					if fallbackPath != "" {
						if targetPage = enjin.FindPage(targetLang, fallbackPath); targetPage == nil {
							if targetPage = enjin.FindPage(language.Und, fallbackPath); targetPage == nil {
								log.DebugF("__%v error: page not found, fallback not found, returning fallback", argv)
								translated = fallbackPath
								return
							}
						}
					} else {
						log.DebugF("__%v error: page not found, fallback not given, returning target", argv)
						translated = targetPath
						return
					}
				}
			}
		}

		if targetPath != targetPage.Url {
			targetPath = targetPage.Url
		}

		// log.WarnF("__: [%v] tp=%v ([%v] %v) - %#v", targetLang, targetPath, targetPage.LanguageTag, targetPage.Url, argv)
		translated = enjin.SiteLanguageMode().ToUrl(enjin.SiteDefaultLanguage(), targetLang, targetPath)
		// log.WarnF("__: [%v] tx=%v ([%v] %v) - %#v", targetLang, translated, targetPage.LanguageTag, targetPage.Url, argv)
		return
	}
}