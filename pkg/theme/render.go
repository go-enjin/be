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
	"strings"
	"time"

	"github.com/gomarkdown/markdown"
	mdHtml "github.com/gomarkdown/markdown/html"
	mdParser "github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"
	"github.com/niklasfasching/go-org/org"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/page"
)

func (t *Theme) RenderOrgModeContent(content string) string {
	input := strings.NewReader(content)
	if html, err := org.New().Parse(input, "./").Write(org.NewHTMLWriter()); err != nil {
		log.ErrorF("error rendering org-mode content: %v", err)
	} else {
		content = html
	}
	return content
}

func (t *Theme) RenderMarkdownContent(content string) string {
	normalizedNewlines := markdown.NormalizeNewlines([]byte(content))
	extensions := mdParser.CommonExtensions |
		mdParser.AutoHeadingIDs |
		mdParser.NoIntraEmphasis |
		mdParser.Strikethrough |
		mdParser.Attributes
	pageParser := mdParser.NewWithExtensions(extensions)
	mdHtmlFlags := mdHtml.CommonFlags | mdHtml.HrefTargetBlank | mdHtml.FootnoteReturnLinks
	opts := mdHtml.RendererOptions{Flags: mdHtmlFlags}
	pageRenderer := mdHtml.NewRenderer(opts)
	parsedBytes := markdown.ToHTML(normalizedNewlines, pageParser, pageRenderer)
	sanitizedBytes := bluemonday.UGCPolicy().SanitizeBytes(parsedBytes)
	return string(sanitizedBytes)
}

func (t *Theme) RenderTemplateContent(ctx context.Context, content string) string {
	if tt, err := template.New("content").Funcs(t.FuncMap).Parse(content); err == nil {
		if partials, ok := t.Layouts["partials"]; ok {
			for _, tmpl := range partials.Tmpl.Templates() {
				_, _ = tt.AddParseTree(tmpl.Name(), tmpl.Tree)
			}
		}
		var w bytes.Buffer
		if err := tt.Execute(&w, ctx); err == nil {
			return string(w.Bytes())
		} else {
			log.ErrorF("error rendering template: %v", err)
		}
	} else {
		log.ErrorF("error parsing template: %v", err)
	}

	return content
}

func (t *Theme) DefaultLayout() (layout *Layout, ok bool) {
	layout, ok = t.Layouts["_default"]
	return
}

func (t *Theme) LayoutFromContext(ctx context.Context) (layout *Layout, name string, err error) {
	name = ctx.String("Layout", "_default")
	var ok bool
	if layout, ok = t.Layouts[name]; !ok {
		if layout, ok = t.Layouts["_default"]; ok {
			name = "_default"
		} else {
			err = fmt.Errorf("%v theme layout not found, expected: \"%v\"", t.Name, name)
			return
		}
	}
	log.DebugF("using layout from context: %v", name)
	return
}

func (t *Theme) TemplateFromContext(view string, ctx context.Context) (tt *template.Template, err error) {
	var layout *Layout
	var layoutName string
	if layout, layoutName, err = t.LayoutFromContext(ctx); err != nil {
		return
	}
	var baseLookups []string
	section := ctx.String("Section", "")
	archetype := ctx.String("Archetype", "")
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
		if _, ok := t.DefaultLayout(); ok {
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
		lookups = append(lookups, name+".html.tmpl", name+".html")
	}
	if tt = layout.Lookup(lookups...); tt == nil {
		log.ErrorF("%v theme lookups checked: %v", t.Name, lookups)
		log.ErrorF("%v context used: %v", t.Name, ctx.AsLogString())
		err = fmt.Errorf("no %v theme template found: archetype=%v, section=%v, layout=%v", t.Name, archetype, section, layoutName)
		return
	}
	return
}

func (t *Theme) Render(view string, ctx context.Context) (data []byte, err error) {
	now := time.Now()
	ctx.Set("CurrentYear", now.Year())
	var tt *template.Template
	if tt, err = t.TemplateFromContext(view, ctx); err == nil {
		log.DebugF("%v theme template used: (%v) %v - \"%v\"", t.Name, view, tt.Name(), ctx.String("Url", "nil"))
		var wr bytes.Buffer
		if err = tt.Execute(&wr, ctx); err != nil {
			return
		}
		data = wr.Bytes()
	}
	return
}

func (t *Theme) RenderPage(ctx context.Context, p *page.Page) (data []byte, err error) {
	ctx.Apply(p.Context.Copy())
	switch p.Format {
	case page.OrgMode:
		ctx["Content"] = t.RenderOrgModeContent(t.RenderTemplateContent(ctx, p.Content))
	case page.Markdown:
		ctx["Content"] = t.RenderMarkdownContent(t.RenderTemplateContent(ctx, p.Content))
	case page.Template:
		ctx["Content"] = t.RenderTemplateContent(ctx, p.Content)
	case page.HtmlTmpl:
		ctx["Content"] = template.HTML(t.RenderTemplateContent(ctx, p.Content))
	default:
		ctx["Content"] = template.HTML(p.Content)
	}
	return t.Render("single", ctx)
}