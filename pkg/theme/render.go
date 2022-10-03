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
)

func (t *Theme) RenderTemplateContent(ctx context.Context, tmplContent string) (html string, err error) {
	var tt *template.Template
	if tt, err = t.NewHtmlTemplate("content").Parse(tmplContent); err == nil {
		var w bytes.Buffer
		if err = tt.Execute(&w, ctx); err == nil {
			html = string(w.Bytes())
			return
		}
	}
	return
}

func (t *Theme) NewHtmlTemplate(name string) (tt *template.Template) {
	tt = template.New(name).Funcs(t.FuncMap)
	t.Layouts.AddPartialsToTemplate(tt)
	return
}

func (t *Theme) FindLayout(named string) (layout *Layout, name string, err error) {
	if named == "" {
		named = "_default"
	}
	name = named
	if layout = t.Layouts.getLayout(name); layout == nil {
		err = fmt.Errorf("%v theme layout not found, expected: \"%v\"", t.Config.Name, name)
	} else {
		log.DebugF("using layout from context: %v", name)
	}
	return
}

func (t *Theme) TemplateFromContext(view string, ctx context.Context) (tt *template.Template, err error) {
	var layout *Layout
	var layoutName string
	if layout, layoutName, err = t.FindLayout(ctx.String("Layout", "_default")); err != nil {
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
		if defaultLayout := t.Layouts.getLayout("_default"); defaultLayout != nil {
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
	if tt = layout.Lookup(lookups...); tt == nil {
		err = fmt.Errorf("%v theme template not found for: archetype=%v, section=%v, layout=%v, pageFormat=%v", t.Config.Name, archetype, section, layoutName, pageFormat)
		return
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

func (t *Theme) renderErrorPage(heading string, output string) (html template.HTML) {
	html = "<header><h1>" + template.HTML(heading) + "</h1></header>\n"
	html += "<section>\n"
	html += "<pre>\n"
	html += template.HTML(template.HTMLEscapeString(output))
	html += "\n</pre>\n"
	html += "</section>"
	return
}

func (t *Theme) RenderPage(ctx context.Context, p *page.Page) (data []byte, err error) {
	ctx.Apply(p.Context.Copy())
	ctx.Set("Theme", t.Config)

	if output, e := t.RenderTemplateContent(ctx, p.Content); e == nil {
		if p.Format == "<unsupported>" {
			ctx["Content"] = t.renderErrorPage("Unsupported Page Format", output)
		} else if format := page.GetFormat(p.Format); format != nil {
			if html, e := format.Process(ctx, t, output); e != nil {
				err = fmt.Errorf("error processing %v page format: %v", p.Format, e)
				return
			} else {
				ctx["Content"] = html
			}
		} else {
			ctx["Content"] = t.renderErrorPage("Unknown Page Format: "+p.Format, output)
		}
	} else {
		err = fmt.Errorf("error rendering template: %v", e)
		return
	}

	return t.Render("single", ctx)
}