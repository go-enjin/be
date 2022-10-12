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

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/page"
	"github.com/go-enjin/be/pkg/theme/types"
)

func (t *Theme) RenderTemplateContent(ctx context.Context, tmplContent string) (html string, err error) {
	var tt *template.Template
	if tt, err = t.NewHtmlTemplate("content.tmpl"); err == nil {
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

func (t *Theme) NewHtmlTemplate(name string) (tmpl *template.Template, err error) {
	if parent := t.GetParent(); parent != nil {
		if tmpl, err = parent.NewHtmlTemplate(name); err != nil {
			return
		}
	} else {
		tmpl = template.New(name).Funcs(t.FuncMap)
		// log.DebugF("starting %v template from theme %v", name, t.Config.Name)
	}

	var layoutsTmpl *template.Template
	if layoutsTmpl, err = t.Layouts.NewTemplate(""); err != nil {
		return
	}
	err = AddParseTree(layoutsTmpl, tmpl)
	return
}

func (t *Theme) FindLayout(named string) (layout *Layout, name string, err error) {
	if named == "" {
		named = "_default"
	}
	name = named
	if layout = t.Layouts.GetLayout(name); layout == nil {
		if parent := t.GetParent(); parent != nil {
			if layout = parent.Layouts.GetLayout(name); layout == nil {
				err = fmt.Errorf("%v (%v) theme layout not found, expected: \"%v\"", t.Config.Name, parent.Config.Name, name)
			} else {
				log.DebugF("found layout in %v (%v) context: %v", t.Config.Name, parent.Config.Name, name)
			}
		} else {
			err = fmt.Errorf("%v theme layout not found, expected: \"%v\"", t.Config.Name, name)
		}
	} else {
		log.DebugF("found layout in %v (%v) context: %v", t.Config.Name, t.Config.Parent, name)
	}
	return
}

func (t *Theme) TemplateFromContext(view string, ctx context.Context) (tt *template.Template, err error) {
	var ctxLayout *Layout
	var layoutName string
	if ctxLayout, layoutName, err = t.FindLayout(ctx.String("Layout", "_default")); err != nil {
		return
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
				baseLookups = append(baseLookups, fmt.Sprintf("_default/%s-baseof", section))
			}
			baseLookups = append(
				baseLookups,
				fmt.Sprintf("_default/%s-baseof", view),
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

	if tmpl, e := ctxLayout.NewTemplate(); e == nil {
		if partials, _, ee := t.FindLayout("partials"); ee == nil {
			if ee = partials.Apply(tmpl); ee != nil {
				err = fmt.Errorf("error applying partials to %v", layoutName)
				return
			}
		}
		if tt = LookupTemplate(tmpl, lookups...); tt == nil {
			e = fmt.Errorf("%v theme template not found for: archetype=%v, section=%v, layout=%v, pageFormat=%v, lookups=%v", t.Config.Name, archetype, section, layoutName, pageFormat, lookups)
			log.ErrorF("checked %d templates", len(tmpl.Templates()))
			for _, ttt := range tmpl.Templates() {
				log.ErrorF("checked: %v", ttt.Name())
			}
		} else {
			log.DebugF("lookup success: %v", tt.Name())
		}
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