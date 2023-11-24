//go:build !exclude_theme_renderer

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

package renderer

import (
	"bytes"
	"fmt"
	htmlTemplate "html/template"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/templates"
)

func (f *CFeature) NewHtmlTemplateWith(t feature.Theme, name string, ctx context.Context) (tmpl *htmlTemplate.Template, err error) {

	var makeTemplate func(t feature.Theme, name string, ctx context.Context) (tmpl *htmlTemplate.Template, err error)
	makeTemplate = func(t feature.Theme, name string, ctx context.Context) (tmpl *htmlTemplate.Template, err error) {
		if parent := t.GetParent(); parent != nil {
			tmpl, err = makeTemplate(parent, name, ctx)
			return
		}
		tmpl, err = t.NewHtmlTemplate(f.Enjin, name, ctx)
		return
	}

	var layoutsTmpl *htmlTemplate.Template
	if layoutsTmpl, err = t.Layouts().NewHtmlTemplate(f.Enjin, "", ctx); err != nil {
		return
	}

	if tmpl, err = makeTemplate(t, name, ctx); err != nil {
		return
	}

	fm := f.Enjin.MakeFuncMap(ctx).AsHTML()
	tmpl = tmpl.Funcs(fm)

	err = templates.AddHtmlParseTree(layoutsTmpl, tmpl)
	return
}

func (f *CFeature) RenderHtmlTemplateContent(t feature.Theme, ctx context.Context, tmplContent string) (rendered string, err error) {
	var tt *htmlTemplate.Template
	if tt, err = f.NewHtmlTemplateWith(t, "content.html.tmpl", ctx); err == nil {
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

func (f *CFeature) NewHtmlTemplateFromContext(t feature.Theme, view string, ctx context.Context) (tt *htmlTemplate.Template, err error) {

	var ctxLayout, parentLayout feature.ThemeLayout
	var layoutName, pagetype, archetype, pageFormat string

	if ctxLayout, parentLayout, layoutName, pagetype, archetype, pageFormat, err = prepareNewTemplateVars(t, ctx); err != nil {
		return
	}

	lookups := makeLookupList(pagetype, pageFormat, archetype, layoutName, view)

	var tmpl *htmlTemplate.Template

	if parent := t.GetParent(); parent != nil {
		if defaults, _, ok := parent.FindLayout(globals.DefaultThemeLayoutName); ok {
			if tmpl, err = defaults.NewHtmlTemplateFrom(f.Enjin, parentLayout, ctx); err != nil {
				err = fmt.Errorf("error creating new parent defaults to %v: %v", layoutName, err)
				return
			}
		}
	}

	if tmpl == nil {
		if tmpl, err = ctxLayout.NewHtmlTemplateFrom(f.Enjin, parentLayout, ctx); err != nil {
			err = fmt.Errorf("error creating new %v layout template: %v", layoutName, err)
			return
		}
	} else if err = ctxLayout.ApplyHtmlTemplate(f.Enjin, tmpl, ctx); err != nil {
		err = fmt.Errorf("error applying parent defaults to %v: %v", layoutName, err)
		return
	}

	if parent := t.GetParent(); parent != nil {
		if partials, _, ok := parent.FindLayout(globals.PartialThemeLayoutName); ok {
			if ee := partials.ApplyHtmlTemplate(f.Enjin, tmpl, ctx); ee != nil {
				err = fmt.Errorf("error applying parent partials to %v: %v", layoutName, ee)
				return
			}
		}
	}

	if partials, _, ok := t.FindLayout(globals.PartialThemeLayoutName); ok {
		if ee := partials.ApplyHtmlTemplate(f.Enjin, tmpl, ctx); ee != nil {
			err = fmt.Errorf("error applying partials to %v: %v", layoutName, ee)
			return
		}
	}

	if tt = templates.LookupHtmlTemplate(tmpl, lookups...); tt == nil {
		err = fmt.Errorf(
			"%v theme template not found for: archetype=%v, type=%v, layout=%v, pageFormat=%v, lookups=%v",
			t.Name(), archetype, pagetype, layoutName, pageFormat, lookups,
		)
	} else {
		log.TraceF("lookup success: %v", tt.Name())
	}

	return
}