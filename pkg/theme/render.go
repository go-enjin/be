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
	"bytes"
	"fmt"
	"html/template"
	"time"

	"github.com/go-enjin/golang-org-x-text/language"
	"github.com/go-enjin/golang-org-x-text/language/display"
	"github.com/go-enjin/golang-org-x-text/message"

	"github.com/go-enjin/be/pkg/types/site"
	"github.com/go-enjin/be/pkg/types/theme-types"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/page"
)

func (t *Theme) RenderTemplateContent(ctx context.Context, tmplContent string) (html string, err error) {
	var tt *template.Template
	if tt, err = t.NewHtmlTemplateWithContext("content.tmpl", ctx); err == nil {
		if tt, err = tt.Parse(tmplContent); err == nil {
			var w bytes.Buffer
			if err = tt.Execute(&w, ctx); err == nil {
				html = string(w.Bytes())
				return
			} else {
				err = fmt.Errorf("error executing template content: %v", err)
			}
		} else {
			err = fmt.Errorf("error parsing template content: %v", err)
		}
	} else {
		err = fmt.Errorf("error making new theme template: %v", err)
	}
	return
}

func (t *Theme) NewFuncMapWithContext(ctx context.Context) (fm template.FuncMap) {
	enjin, _ := ctx.Get("SiteEnjin").(site.Enjin)

	fm = template.FuncMap{}
	for k, v := range t.FuncMap {
		fm[k] = v
	}

	// translate page paths
	fm["__"] = func(argv ...string) (translated string, err error) {
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
								log.ErrorF("__%v error: page not found, fallback not found, returning fallback", argv)
								translated = fallbackPath
								return
							}
						}
					} else {
						log.ErrorF("__%v error: page not found, fallback not given, returning target", argv)
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

	// translate page content
	fm["_"] = func(format string, argv ...interface{}) (translated string) {
		if printer, ok := ctx.Get("LangPrinter").(*message.Printer); ok {
			translated = printer.Sprintf(format, argv...)
			if fmt.Sprintf(format, argv...) != translated {
				log.DebugF("template translated: \"%v\" -> \"%v\"", format, translated)
			}
		} else {
			translated = fmt.Sprintf(format, argv...)
		}
		return
	}

	fm["displayLangTag"] = func(tag language.Tag) (name string) {
		name = display.Tags(tag).Name(tag)
		return
	}
	return
}

func (t *Theme) NewHtmlTemplateWithContext(name string, ctx context.Context) (tmpl *template.Template, err error) {
	if parent := t.GetParent(); parent != nil {
		if tmpl, err = parent.NewHtmlTemplateWithContext(name, ctx); err != nil {
			return
		}
		// log.DebugF("starting %v template from theme %v", name, t.Config.Name)
	} else {
		tmpl = template.New(name).Funcs(t.NewFuncMapWithContext(ctx))
		log.DebugF("starting %v template from theme %v", name, t.Config.Name)
	}

	var layoutsTmpl *template.Template
	if layoutsTmpl, err = t.Layouts.NewTemplate("", ctx); err != nil {
		return
	}
	err = AddParseTree(layoutsTmpl, tmpl)
	return
}

func (t *Theme) NewHtmlTemplate(name string) (tmpl *template.Template, err error) {
	tmpl, err = t.NewHtmlTemplateWithContext(name, context.New())
	return
}

func (t *Theme) FindLayout(named string) (layout *Layout, name string, ok bool) {
	if named == "" {
		named = "_default"
	}
	name = named

	layout = t.Layouts.GetLayout(name)
	if ok = layout != nil; ok {
		log.DebugF("found layout in %v (%v) context: %v", t.Config.Name, t.Config.Parent, name)
	}
	return
}

func (t *Theme) TemplateFromContext(view string, ctx context.Context) (tt *template.Template, err error) {
	var ok bool
	var ctxLayout, parentLayout *Layout
	layoutName := ctx.String("Layout", "_default")
	if ctxLayout, layoutName, ok = t.FindLayout(layoutName); !ok {
		if parent := t.GetParent(); parent != nil {
			if ctxLayout, _, ok = parent.FindLayout(layoutName); !ok {
				err = fmt.Errorf("%v layout not found in %v (%v)", layoutName, t.Name, parent.Name)
				return
			}
		} else {
			err = fmt.Errorf("%v layout not found in %v", layoutName, t.Name)
			return
		}
	} else {
		if parent := t.GetParent(); parent != nil {
			parentLayout, _, _ = parent.FindLayout(layoutName)
		}
	}

	var baseLookups []string
	section := ctx.String("Section", "")
	archetype := ctx.String("Archetype", "")
	pageFormat := ctx.String("Format", "")
	if archetype != "" {
		if section != "" {
			baseLookups = append(baseLookups, fmt.Sprintf("%s/%s", archetype, section))
		}
		baseLookups = append(baseLookups, fmt.Sprintf("%s/%s", archetype, view))
	} else if section != "" {
		baseLookups = append(
			baseLookups,
			fmt.Sprintf("%s/%s", layoutName, section),
			fmt.Sprintf("%s/%s-baseof", layoutName, section),
		)
	}
	baseLookups = append(
		baseLookups,
		fmt.Sprintf("%s/%s", layoutName, view),
		fmt.Sprintf("%s/%s-baseof", layoutName, view),
		fmt.Sprintf("%s/baseof", layoutName),
	)

	if layoutName != "_default" {
		if ctxLayout != nil {
			if section != "" {
				baseLookups = append(baseLookups, fmt.Sprintf("%v/%s-baseof", layoutName, section))
			}
			baseLookups = append(
				baseLookups,
				fmt.Sprintf("%v/%s-baseof", layoutName, view),
				"_default/baseof",
			)
		}
	}

	var lookups []string
	for _, name := range baseLookups {
		lookups = append(lookups, name+".tmpl")
		if pageFormat != "" {
			lookups = append(lookups, name+"."+pageFormat+".tmpl")
		}
	}

	var tmpl *template.Template
	if ctxTmpl, e := ctxLayout.NewTemplateFrom(parentLayout, ctx); e != nil {
		err = fmt.Errorf("error creating new %v layout template: %v", layoutName, e)
		return
	} else {
		tmpl = ctxTmpl
	}

	if parent := t.GetParent(); parent != nil {
		if partials, _, ok := parent.FindLayout("partials"); ok {
			if ee := partials.Apply(tmpl, ctx); ee != nil {
				err = fmt.Errorf("error applying parent partials to %v: %v", layoutName, ee)
				return
			}
		}
	}

	if partials, _, ok := t.FindLayout("partials"); ok {
		if ee := partials.Apply(tmpl, ctx); ee != nil {
			err = fmt.Errorf("error applying partials to %v: %v", layoutName, ee)
			return
		}
	}

	if tt = LookupTemplate(tmpl, lookups...); tt == nil {
		err = fmt.Errorf("%v theme template not found for: archetype=%v, section=%v, layout=%v, pageFormat=%v, lookups=%v", t.Config.Name, archetype, section, layoutName, pageFormat, lookups)
	} else {
		log.DebugF("lookup success: %v", tt.Name())
	}

	return
}

func (t *Theme) Render(view string, ctx context.Context) (data []byte, err error) {
	now := time.Now()
	ctx.Set("CurrentYear", now.Year())
	var tt *template.Template
	if tt, err = t.TemplateFromContext(view, ctx); err == nil {
		log.DebugF("%v theme template used: (%v) %v - \"%v\"", t.Config.Name, view, tt.Name(), ctx.String("Url", "nil"))
		var wr bytes.Buffer
		if err = tt.Execute(&wr, ctx); err != nil {
			return
		}
		data = wr.Bytes()
	}
	return
}

func (t *Theme) renderErrorPage(title, summary, output string) (html template.HTML) {
	html = types.NewEnjinError(title, summary, output).Html()
	return
}

func (t *Theme) RenderPage(ctx context.Context, p *page.Page) (data []byte, err error) {
	ctx.Apply(p.Context.Copy())
	ctx.Set("Theme", t.GetConfig())

	if output, e := t.RenderTemplateContent(ctx, p.Content); e == nil {
		if format := page.GetFormat(p.Format); format != nil {
			if html, ee := format.Process(ctx, t, output); ee != nil {
				log.ErrorF("error processing %v page format: %v - %v", p.Format, ee.Title, ee.Summary)
				ctx["Content"] = ee.Html()
			} else {
				ctx["Content"] = html
				log.DebugF("page format success: %v", format.Name())
			}
		} else {
			ctx["Content"] = t.renderErrorPage("Unsupported Page Format", fmt.Sprintf(`Unknown page format specified: "%v"`, p.Format), p.String())
		}
	} else {
		ctx["Content"] = t.renderErrorPage("Template Render Error", e.Error(), p.String())
	}

	return t.Render("single", ctx)
}