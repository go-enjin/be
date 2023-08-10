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
	htmlTemplate "html/template"
	"strings"
	textTemplate "text/template"

	"github.com/go-enjin/be/pkg/request/argv"
	"github.com/go-enjin/be/pkg/types/theme-types"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/page"
)

func (t *Theme) RenderHtmlTemplateContent(ctx context.Context, tmplContent string) (rendered string, err error) {
	var tt *htmlTemplate.Template
	if tt, err = t.NewHtmlTemplateWithContext("content.tmpl", ctx); err == nil {
		if tt, err = tt.Funcs(DefaultFuncMap()).Parse(tmplContent); err == nil {
			var w bytes.Buffer
			if err = tt.Execute(&w, ctx); err == nil {
				rendered = string(w.Bytes())
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

func (t *Theme) RenderTextTemplateContent(ctx context.Context, tmplContent string) (rendered string, err error) {
	var tt *textTemplate.Template
	if tt, err = t.NewTextTemplateWithContext("content.tmpl", ctx); err == nil {
		if tt, err = tt.Parse(tmplContent); err == nil {
			var w bytes.Buffer
			if err = tt.Execute(&w, ctx); err == nil {
				rendered = string(w.Bytes())
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

func (t *Theme) NewTextTemplateWithContext(name string, ctx context.Context) (tmpl *textTemplate.Template, err error) {
	if parent := t.GetParent(); parent != nil {
		if tmpl, err = parent.NewTextTemplateWithContext(name, ctx); err != nil {
			return
		}
	} else {
		tmpl = textTemplate.New(name).Funcs(t.NewTextFuncMapWithContext(ctx))
		log.TraceF("starting %v (text) template from theme %v", name, t.Config.Name)
	}
	return
}

func (t *Theme) NewHtmlTemplateWithContext(name string, ctx context.Context) (tmpl *htmlTemplate.Template, err error) {
	if parent := t.GetParent(); parent != nil {
		if tmpl, err = parent.NewHtmlTemplateWithContext(name, ctx); err != nil {
			return
		}
	} else {
		tmpl = htmlTemplate.New(name).Funcs(t.NewHtmlFuncMapWithContext(ctx))
		log.TraceF("starting %v (html) template from theme %v", name, t.Config.Name)
	}
	var layoutsTmpl *htmlTemplate.Template
	if layoutsTmpl, err = t.Layouts.NewTemplate("", ctx); err != nil {
		return
	}
	err = AddParseTree(layoutsTmpl, tmpl)
	return
}

func (t *Theme) FindLayout(named string) (layout *Layout, name string, ok bool) {
	if named == "" {
		named = "_default"
	}
	name = named

	layout = t.Layouts.GetLayout(name)
	if ok = layout != nil; ok {
		if t.Config.Parent == "" {
			log.TraceF("found layout in %v (theme) context: %v", t.Config.Name, name)
		} else {
			log.TraceF("found layout in %v (%v) context: %v", t.Config.Name, t.Config.Parent, name)
		}
	}
	return
}

func (t *Theme) TemplateFromContext(view string, ctx context.Context) (tt *htmlTemplate.Template, err error) {
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
	} else if parent := t.GetParent(); parent != nil {
		parentLayout, _, _ = parent.FindLayout(layoutName)
	}

	//  - layouts/_default/blog-single.fmt.tmpl
	//  - layouts/_default/blog.fmt.tmpl
	//  - layouts/_default/blog.tmpl

	/*
		layouts/posts/single-baseof.html.html
		layouts/posts/baseof.html.html
		layouts/posts/single-baseof.html
		layouts/posts/baseof.html
		layouts/_default/single-baseof.html.html
		layouts/_default/baseof.html.html
		layouts/_default/single-baseof.html
		layouts/_default/baseof.html
	*/

	var lookups []string

	pagetype := ctx.String("Type", "page")
	archetype := ctx.String("Archetype", "")
	pageFormat := ctx.String("Format", "")
	pageFormat = strings.TrimSuffix(pageFormat, ".tmpl")

	addLookup := func(layoutName, archetype, view, format, extn string) {
		name := layoutName + "/"
		if archetype != "" {
			name += archetype
			if view != "" {
				name += "-"
			}
		}
		if view != "" {
			name += view
		}
		if format != "" {
			name += "." + format
		}
		name += "." + extn
		lookups = append(lookups, name)
	}

	/*
		layouts/{layoutName}/{archetype,pagetype}-{view}.{fmt}.{html,tmpl}
		layouts/{layoutName}/{archetype,pagetype}.{fmt}.{html,tmpl}
		layouts/{layoutName}/{view}.{fmt}.{html,tmpl}
		layouts/{layoutName}/baseof.{fmt}.{html,tmpl}
		layouts/{layoutName}/{archetype,pagetype}-{view}.{html,tmpl}
		layouts/{layoutName}/{archetype,pagetype}.{html,tmpl}
		layouts/{layoutName}/{view}.{html,tmpl}
		layouts/{layoutName}/baseof.{html,tmpl}
		layouts/_default/{archetype,pagetype}-{view}.{fmt}.{html,tmpl}
		layouts/_default/{archetype,pagetype}.{fmt}.{html,tmpl}
		layouts/_default/{view}.{fmt}.{html,tmpl}
		layouts/_default/baseof.{fmt}.{html,tmpl}
		layouts/_default/{archetype,pagetype}-{view}.{html,tmpl}
		layouts/_default/{archetype,pagetype}.{html,tmpl}
		layouts/_default/{view}.{html,tmpl}
		layouts/_default/baseof.{html,tmpl}
	*/

	if pageFormat != "" {
		if archetype != "" {
			addLookup(layoutName, archetype, view, pageFormat, "tmpl")
			addLookup(layoutName, archetype, view, pageFormat, "html")
			addLookup(layoutName, archetype, "", pageFormat, "tmpl")
			addLookup(layoutName, archetype, "", pageFormat, "html")
		}
		addLookup(layoutName, pagetype, view, pageFormat, "tmpl")
		addLookup(layoutName, pagetype, view, pageFormat, "html")
		addLookup(layoutName, pagetype, "", pageFormat, "tmpl")
		addLookup(layoutName, pagetype, "", pageFormat, "html")
		addLookup(layoutName, "", view, pageFormat, "tmpl")
		addLookup(layoutName, "", view, pageFormat, "html")
		addLookup(layoutName, "", "baseof", pageFormat, "tmpl")
		addLookup(layoutName, "", "baseof", pageFormat, "html")
	}
	if archetype != "" {
		addLookup(layoutName, archetype, view, "", "tmpl")
		addLookup(layoutName, archetype, view, "", "html")
		addLookup(layoutName, archetype, "", "", "tmpl")
		addLookup(layoutName, archetype, "", "", "html")
	}
	addLookup(layoutName, pagetype, view, "", "tmpl")
	addLookup(layoutName, pagetype, view, "", "html")
	addLookup(layoutName, pagetype, "", "", "tmpl")
	addLookup(layoutName, pagetype, "", "", "html")
	addLookup(layoutName, "", view, "", "tmpl")
	addLookup(layoutName, "", view, "", "html")
	addLookup(layoutName, "", "baseof", "", "tmpl")
	addLookup(layoutName, "", "baseof", "", "html")

	if layoutName != "_default" {
		if pageFormat != "" {
			if archetype != "" {
				addLookup("_default", archetype, view, pageFormat, "tmpl")
				addLookup("_default", archetype, view, pageFormat, "html")
				addLookup("_default", archetype, "", pageFormat, "tmpl")
				addLookup("_default", archetype, "", pageFormat, "html")
			}
			addLookup("_default", pagetype, view, pageFormat, "tmpl")
			addLookup("_default", pagetype, view, pageFormat, "html")
			addLookup("_default", pagetype, "", pageFormat, "tmpl")
			addLookup("_default", pagetype, "", pageFormat, "html")
			addLookup("_default", "", view, pageFormat, "tmpl")
			addLookup("_default", "", view, pageFormat, "html")
			addLookup("_default", "", "baseof", pageFormat, "tmpl")
			addLookup("_default", "", "baseof", pageFormat, "html")
		}
		if archetype != "" {
			addLookup("_default", archetype, view, "", "tmpl")
			addLookup("_default", archetype, view, "", "html")
			addLookup("_default", archetype, "", "", "tmpl")
			addLookup("_default", archetype, "", "", "html")
		}
		addLookup("_default", pagetype, view, "", "tmpl")
		addLookup("_default", pagetype, view, "", "html")
		addLookup("_default", pagetype, "", "", "tmpl")
		addLookup("_default", pagetype, "", "", "html")
		addLookup("_default", "", view, "", "tmpl")
		addLookup("_default", "", view, "", "html")
		addLookup("_default", "", "baseof", "", "tmpl")
		addLookup("_default", "", "baseof", "", "html")
	}

	var tmpl *htmlTemplate.Template
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
		err = fmt.Errorf(
			"%v theme template not found for: archetype=%v, type=%v, layout=%v, pageFormat=%v, lookups=%v",
			t.Config.Name, archetype, pagetype, layoutName, pageFormat, lookups,
		)
	} else {
		log.TraceF("lookup success: %v", tt.Name())
	}

	return
}

func (t *Theme) Render(view string, ctx context.Context) (data []byte, err error) {
	var tt *htmlTemplate.Template
	if tt, err = t.TemplateFromContext(view, ctx); err == nil {
		log.TraceF("%v theme template used: (%v) %v - \"%v\"", t.Config.Name, view, tt.Name(), ctx.String("Url", "nil"))
		var wr bytes.Buffer
		if err = tt.Execute(&wr, ctx); err != nil {
			return
		}
		data = wr.Bytes()
	}
	return
}

func (t *Theme) renderErrorPage(title, summary, output string) (html htmlTemplate.HTML) {
	html = types.NewEnjinError(title, summary, output).Html()
	return
}

func (t *Theme) RenderPage(ctx context.Context, p *page.Page) (data []byte, redirect string, err error) {
	ctx.Apply(p.Context.Copy())
	ctx.Set("Theme", t.GetConfig())

	var output string

	if p.Format == "html.tmpl" {
		if output, err = t.RenderHtmlTemplateContent(ctx, p.Content); err != nil {
			ctx["Content"] = t.renderErrorPage("Template Render Error", err.Error(), p.String())
		}
	} else if strings.HasSuffix(p.Format, ".tmpl") {
		// TODO: find a more safe way to pre-render .njn.tmpl files
		if output, err = t.RenderTextTemplateContent(ctx, p.Content); err != nil {
			ctx["Content"] = t.renderErrorPage("Template Render Error", err.Error(), p.String())
		}
	} else {
		output = p.Content
	}

	if err == nil {
		if format := t.GetFormat(p.Format); format != nil {
			if html, redir, ee := format.Process(ctx, t, output); ee != nil {
				log.ErrorF("error processing %v page format: %v - %v", p.Format, ee.Title, ee.Summary)
				ctx["Content"] = ee.Html()
			} else if redir != "" {
				redirect = redir
				return
			} else {
				ctx["Content"] = html
				log.TraceF("page format success: %v", format.Name())
			}
		} else {
			ctx["Content"] = t.renderErrorPage("Unsupported Page Format", fmt.Sprintf(`Unknown page format specified: "%v"`, p.Format), p.String())
		}
	}

	if !p.Context.Bool(argv.RequestArgvIgnoredKey, false) {
		if redirect = ctx.String(argv.RequestRedirectKey, ""); redirect != "" {
			return
		} else if consumed := ctx.Bool(argv.RequestArgvConsumedKey, false); !consumed {
			if reqArgv, ok := ctx.Get(string(argv.RequestArgvKey)).(*argv.RequestArgv); ok && reqArgv != nil && reqArgv.MustConsume() {
				redirect = p.Url
				return
			}
		}
	}

	data, err = t.Render("single", ctx)
	return
}