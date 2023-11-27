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
	textTemplate "text/template"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/templates"
)

func (f *CFeature) NewTextTemplateWith(t feature.Theme, name string, ctx context.Context) (tmpl *textTemplate.Template, err error) {

	var makeTemplate func(t feature.Theme, name string, ctx context.Context) (tmpl *textTemplate.Template, err error)
	makeTemplate = func(t feature.Theme, name string, ctx context.Context) (tmpl *textTemplate.Template, err error) {
		if parent := t.GetParent(); parent != nil {
			tmpl, err = makeTemplate(parent, name, ctx)
			return
		}
		tmpl, err = t.NewTextTemplate(f.Enjin, name, ctx)
		return
	}

	var layoutsTmpl *textTemplate.Template
	if layoutsTmpl, err = t.Layouts().NewTextTemplate(f.Enjin, "", ctx); err != nil {
		return
	}

	if tmpl, err = makeTemplate(t, name, ctx); err != nil {
		return
	}

	fm := f.Enjin.MakeFuncMap(ctx).AsTEXT()
	tmpl = tmpl.Funcs(fm)

	err = templates.AddTextParseTree(layoutsTmpl, tmpl)
	return
}

func (f *CFeature) RenderTextTemplateContent(t feature.Theme, ctx context.Context, tmplContent string) (rendered string, err error) {
	var tt *textTemplate.Template
	if tt, err = f.NewTextTemplateWith(t, "content.tmpl", ctx); err == nil {
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

func (f *CFeature) NewTextTemplateFromContext(t feature.Theme, view string, ctx context.Context) (tt *textTemplate.Template, err error) {

	var ctxLayout, parentLayout feature.ThemeLayout
	var layoutName, pagetype, archetype, pageFormat string

	if ctxLayout, parentLayout, layoutName, pagetype, archetype, pageFormat, err = prepareNewTemplateVars(t, ctx); err != nil {
		return
	}

	lookups := makeLookupList(pagetype, pageFormat, archetype, layoutName, view)

	var tmpl *textTemplate.Template
	if ctxTmpl, e := ctxLayout.NewTextTemplateFrom(f.Enjin, parentLayout, ctx); e != nil {
		err = fmt.Errorf("error creating new %v layout template: %v", layoutName, e)
		return
	} else {
		tmpl = ctxTmpl
	}

	if parent := t.GetParent(); parent != nil {
		if partials, _, ok := parent.FindLayout(globals.PartialThemeLayoutName); ok {
			if ee := partials.ApplyTextTemplate(f.Enjin, tmpl, ctx); ee != nil {
				err = fmt.Errorf("error applying parent partials to %v: %v", layoutName, ee)
				return
			}
		}
	}

	if partials, _, ok := t.FindLayout(globals.PartialThemeLayoutName); ok {
		if ee := partials.ApplyTextTemplate(f.Enjin, tmpl, ctx); ee != nil {
			err = fmt.Errorf("error applying partials to %v: %v", layoutName, ee)
			return
		}
	}

	if tt = templates.LookupTextTemplate(tmpl, lookups...); tt == nil {
		err = fmt.Errorf(
			"%v theme template not found for: archetype=%v, type=%v, layout=%v, pageFormat=%v, lookups=%v",
			t.Name(), archetype, pagetype, layoutName, pageFormat, lookups,
		)
	} else {
		log.TraceF("lookup success: %v", tt.Name())
	}

	return
}
