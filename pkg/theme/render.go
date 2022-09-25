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
	"encoding/json"
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

func (t *Theme) RenderOrgModeContent(content string) (html string, ok bool) {
	var err error
	input := strings.NewReader(content)
	if html, err = org.New().Parse(input, "./").Write(org.NewHTMLWriter()); err != nil {
		log.ErrorF("error rendering org-mode content: %v", err)
		html = "<h2>Internal Theming Error</h2>\n"
		html += fmt.Sprintf("<p>Error rendering org-mode content: \"%v\"</p>", err)
		html += "<pre>\n"
		html += content
		html += "\n</pre>\n"
	}
	ok = err == nil
	return
}

func (t *Theme) RenderMarkdownContent(content string) (html string, ok bool) {
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
	ok = true
	html = string(sanitizedBytes)
	return
}

func (t *Theme) RenderTemplateContent(ctx context.Context, tmplContent string) (html string, ok bool) {
	var err error
	var tt *template.Template
	if tt, err = t.NewHtmlTemplate("content").Parse(tmplContent); err == nil {
		var w bytes.Buffer
		if err = tt.Execute(&w, ctx); err == nil {
			ok = true
			html = string(w.Bytes())
			return
		} else {
			log.ErrorF("error rendering template: %v", err)
		}
	} else {
		log.ErrorF("error parsing template: %v", err)
	}
	html += "<h2>Internal Theming Error</h2>\n"
	html += fmt.Sprintf("<p>Error rendering template content: \"%v\"</p>", err)
	html += "<pre>\n"
	html += tmplContent
	html += "\n</pre>\n"
	return
}

func (t *Theme) FindAllPartials() (partials []*Layout) {
	if p, ok := t.Layouts["partials"]; ok {
		partials = append(partials, p)
	}
	return
}

func (t *Theme) AddAllPartialsToHtmlTemplate(tt *template.Template) {
	partials := t.FindAllPartials()
	for _, partial := range partials {
		for _, tmpl := range partial.Tmpl.Templates() {
			_, _ = tt.AddParseTree(tmpl.Name(), tmpl.Tree)
		}
	}
}

func (t *Theme) NewHtmlTemplate(name string) (tt *template.Template) {
	tt = template.New(name).Funcs(t.FuncMap)
	t.AddAllPartialsToHtmlTemplate(tt)
	return
}

func (t *Theme) ListAllFiles(path string) (paths []string) {
	if filenames, err := t.FileSystem.ListAllFiles(path); err == nil {
		paths = append(paths, filenames...)
	} else {
		log.ErrorF("error listing all (%v) %v theme files: %v", path, t.Name, err)
	}
	return
}

func (t *Theme) LayoutFromContext(ctx context.Context) (layout *Layout, name string, err error) {
	return t.FindLayout(ctx.String("Layout", "_default"))
}

func (t *Theme) FindLayout(named string) (layout *Layout, name string, err error) {
	if named == "" {
		named = "_default"
	}
	name = named
	var ok bool
	if layout, ok = t.Layouts[name]; !ok {
		err = fmt.Errorf("%v theme layout not found, expected: \"%v\"", t.Name, name)
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
		if _, ok := t.Layouts["_default"]; ok {
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
		lookups = append(lookups, name+".html.tmpl", name+".tmpl", name+".html")
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
		if text, ok := t.RenderTemplateContent(ctx, p.Content); ok {
			ctx["Content"], _ = t.RenderOrgModeContent(text)
		} else {
			ctx["Content"] = text
		}
	case page.Markdown:
		if text, ok := t.RenderTemplateContent(ctx, p.Content); ok {
			ctx["Content"], _ = t.RenderMarkdownContent(text)
		} else {
			ctx["Content"] = text
		}
	case page.Template:
		ctx["Content"], _ = t.RenderTemplateContent(ctx, p.Content)
	case page.HtmlTmpl:
		if text, ok := t.RenderTemplateContent(ctx, p.Content); ok {
			ctx["Content"] = template.HTML(text)
		} else {
			ctx["Content"] = text
		}
	case page.Semantic:
		if text, ok := t.RenderTemplateContent(ctx, p.Content); ok {
			ctx["Content"], _ = t.RenderSemanticContent(ctx, text)
		} else {
			ctx["Content"] = text
		}
	default:
		ctx["Content"] = p.Content
	}
	return t.Render("single", ctx)
}

func (t *Theme) RenderSemanticContent(ctx context.Context, content string) (html template.HTML, ok bool) {
	var err error
	if html, err = t.parseSemanticContent(ctx, content); err != nil {
		log.ErrorF("error parsing semantic content: %v", err)
		html = "<h1>Internal Theming Error</h1>\n"
		html += template.HTML(fmt.Sprintf("<p>Error parsing semantic content: \"%v\"</p>\n", err))
		html += "<pre>\n"
		html += template.HTML(template.HTMLEscapeString(content))
		html += "\n</pre>\n"
	}
	ok = err == nil
	return
}

func (t *Theme) parseSemanticContent(ctx context.Context, content string) (html template.HTML, err error) {
	var data interface{}
	if err = json.Unmarshal([]byte(content), &data); err != nil {
		return "", err
	}
	renderer := newNjnRenderer(ctx, t)
	html, err = renderer.render(ctx, data)
	return
}