//go:build page_funcmaps || pages || all

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

package lang

import (
	"fmt"
	"net/http"

	"github.com/urfave/cli/v2"

	cllang "github.com/go-corelibs/lang"
	"github.com/go-corelibs/x-text/language"
	"github.com/go-corelibs/x-text/language/display"
	"github.com/go-corelibs/x-text/message"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "pages-funcmaps-lang"

type Feature interface {
	feature.Feature
	feature.FuncMapProvider
}

type MakeFeature interface {
	Make() Feature
}

type CFeature struct {
	feature.CFeature
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.CFeature.Construct(f)
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	return
}

func (f *CFeature) Make() (feat Feature) {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	return
}

func (f *CFeature) MakeFuncMap(ctx beContext.Context) (fm feature.FuncMap) {
	fm = feature.FuncMap{
		"cmpLang": CmpLang,
	}
	if f.Enjin != nil {
		fm["_"] = f.makeUnderscore(ctx)
		fm["__"] = f.makeUnderscoreUnderscore(ctx)
		fm["___"] = f.makeUnderscoreUnderscoreUnderscore(ctx)
		fm["_tag"] = f.makeUnderscoreTag(ctx)
		fm["__tag"] = f.makeUnderscoreUnderscoreTag(ctx)
		fm["_txs"] = f.makeTranslations(ctx)
	}
	return
}

func (f *CFeature) makeTranslations(ctx beContext.Context) interface{} {
	cache := map[string]feature.Pages{}
	return func(url string) (translations feature.Pages) {
		if _, cached := cache[url]; !cached {
			cache[url] = f.Enjin.FindTranslations(url)
		}
		translations = cache[url]
		return
	}
}

func (f *CFeature) makeUnderscore(ctx beContext.Context) interface{} {
	cache := map[string]string{}
	return func(format string, argv ...interface{}) (translated string) {
		var ok bool
		untranslated := fmt.Sprintf(format, argv...)
		if translated, ok = cache[untranslated]; ok {
			//log.DebugF("template underscore cached hit: \"%v\" -> \"%v\"", format, translated)
			return
		}
		if printer, ok := ctx.Get(lang.PrinterKey).(*message.Printer); ok && printer != nil {
			translated = printer.Sprintf(format, argv...)
			cache[untranslated] = translated
			if untranslated != translated {
				log.TraceF("template underscore translated: \"%v\" -> \"%v\"", format, translated)
			} else {
				log.TraceF("template underscore defaulting: \"%v\" -> \"%v\"", format, translated)
			}
		} else {
			log.TraceF("template underscore language printer not found, using fmt.Sprintf")
			translated = fmt.Sprintf(format, argv...)
		}
		return
	}
}

func (f *CFeature) makeUnderscoreUnderscore(ctx beContext.Context) interface{} {
	return func(argvInput ...interface{}) (translated string, err error) {
		r, _ := ctx.Get("R").(*http.Request)
		targetLang, _ := ctx.Get("ReqLangTag").(language.Tag)

		var argv []string
		for _, input := range argvInput {
			argv = append(argv, fmt.Sprintf("%v", input))
		}
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

		if !f.Enjin.SiteSupportsLanguage(targetLang) {
			log.DebugF("unsupported site language requested: %v, reverting to default", targetLang)
			targetLang = f.Enjin.SiteDefaultLanguage()
		}

		var targetPage feature.Page
		if targetPage = f.Enjin.FindPage(r, targetLang, targetPath); targetPage == nil {
			if found := f.Enjin.FindTranslations(targetPath); len(found) > 0 {
				for _, pg := range found {
					if pg.IsTranslation(targetPath) {
						if language.Compare(pg.LanguageTag(), targetLang) {
							targetPage = pg
							break
						}
					} else {
						targetPage = f.Enjin.FindPage(r, targetLang, pg.Translates())
						break
					}
				}
			}

			if targetPage == nil {
				if targetPage = f.Enjin.FindPage(r, language.Und, targetPath); targetPage == nil {
					if fallbackPath != "" {
						if targetPage = f.Enjin.FindPage(r, targetLang, fallbackPath); targetPage == nil {
							if targetPage = f.Enjin.FindPage(r, language.Und, fallbackPath); targetPage == nil {
								log.TraceF("__%v error: page not found, fallback not found, returning fallback", argv)
								translated = fallbackPath
								return
							}
						}
					} else {
						log.TraceF("__%v error: page not found, fallback not given, returning target", argv)
						translated = targetPath
						return
					}
				}
			}
		}

		if targetPath != targetPage.Url() {
			targetPath = targetPage.Url()
		}

		// log.WarnF("__: [%v] tp=%v ([%v] %v) - %#v", targetLang, targetPath, targetPage.LanguageTag, targetPage.Url, argv)
		translated = f.Enjin.SiteLanguageMode().ToUrl(f.Enjin.SiteDefaultLanguage(), targetLang, targetPath)
		// log.WarnF("__: [%v] tx=%v ([%v] %v) - %#v", targetLang, translated, targetPage.LanguageTag, targetPage.Url, argv)
		return
	}
}

func (f *CFeature) makeUnderscoreUnderscoreUnderscore(ctx beContext.Context) interface{} {
	return func(argv ...string) (translated string, err error) {
		targetLang, _ := ctx.Get("ReqLangTag").(language.Tag)
		var targetPath string

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
		default:
			err = fmt.Errorf("called with too many arguments")
			return
		}

		if targetPath == "" || targetPath[0] != '/' {
			translated = targetPath
			return
		}

		if !f.Enjin.SiteSupportsLanguage(targetLang) {
			log.DebugF("unsupported site language requested: %v, reverting to default", targetLang)
			targetLang = f.Enjin.SiteDefaultLanguage()
		}

		// log.WarnF("__: [%v] tp=%v ([%v] %v) - %#v", targetLang, targetPath, targetPage.LanguageTag, targetPage.Url, argv)
		translated = f.Enjin.SiteLanguageMode().ToUrl(f.Enjin.SiteDefaultLanguage(), targetLang, targetPath)
		// log.WarnF("__: [%v] tx=%v ([%v] %v) - %#v", targetLang, translated, targetPage.LanguageTag, targetPage.Url, argv)
		return
	}
}

func (f *CFeature) makeUnderscoreTag(ctx beContext.Context) interface{} {
	return func(tagOrString interface{}) (name string, err error) {
		var tag language.Tag
		if tag, err = cllang.ParseTag(tagOrString); err != nil {
			return
		}
		var ok bool
		if name, ok = f.Enjin.SiteLanguageDisplayName(tag); !ok {
			name = display.Tags(tag).Name(tag)
		}
		return
	}
}

func (f *CFeature) makeUnderscoreUnderscoreTag(ctx beContext.Context) interface{} {
	return func(txInput, tagInput interface{}) (name string, err error) {
		var tx, tag language.Tag
		if tx, err = cllang.ParseTag(txInput); err != nil {
			return
		} else if tag, err = cllang.ParseTag(tagInput); err != nil {
			return
		}
		name = display.Tags(tx).Name(tag)
		return
	}
}

func CmpLang(a interface{}, other ...interface{}) (equal bool, err error) {
	var aTag language.Tag
	var oTags []language.Tag

	if aTag, err = cllang.ParseTag(a); err != nil {
		return
	}

	for _, o := range other {
		var oTag language.Tag
		if oTag, err = cllang.ParseTag(o); err != nil {
			return
		}
		oTags = append(oTags, oTag)
	}

	equal = language.Compare(aTag, oTags...)
	return
}
