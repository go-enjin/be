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
	"fmt"
	"strings"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
)

func prepareNewTemplateVars(t feature.Theme, ctx context.Context) (ctxLayout, parentLayout feature.ThemeLayout, layoutName, pagetype, archetype, pageFormat string, err error) {
	var ok bool
	layoutName = ctx.String("Layout", globals.DefaultThemeLayoutName)
	if ctxLayout, layoutName, ok = t.FindLayout(layoutName); !ok {
		if parent := t.GetParent(); parent != nil {
			if ctxLayout, _, ok = parent.FindLayout(layoutName); !ok {
				err = fmt.Errorf("%v layout not found in %v (%v)", layoutName, t.Name(), parent.Name())
				return
			}
		} else {
			err = fmt.Errorf("%v layout not found in %v", layoutName, t.Name())
			return
		}
	} else if parent := t.GetParent(); parent != nil {
		parentLayout, _, _ = parent.FindLayout(layoutName)
	}

	pagetype = ctx.String("Type", "page")
	archetype = ctx.String("Archetype", "")
	pageFormat = ctx.String("Format", "")
	pageFormat = strings.TrimSuffix(pageFormat, ".tmpl")
	return
}

func makeLookupList(pagetype, pageFormat, archetype, layoutName, view string) (lookups []string) {

	//  - layouts/~default/blog-single.fmt.tmpl
	//  - layouts/~default/blog.fmt.tmpl
	//  - layouts/~default/blog.tmpl

	/*
		layouts/posts/single-baseof.html.html
		layouts/posts/baseof.html.html
		layouts/posts/single-baseof.html
		layouts/posts/baseof.html
		layouts/~default/single-baseof.html.html
		layouts/~default/baseof.html.html
		layouts/~default/single-baseof.html
		layouts/~default/baseof.html
	*/

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
		layouts/~default/{archetype,pagetype}-{view}.{fmt}.{html,tmpl}
		layouts/~default/{archetype,pagetype}.{fmt}.{html,tmpl}
		layouts/~default/{view}.{fmt}.{html,tmpl}
		layouts/~default/baseof.{fmt}.{html,tmpl}
		layouts/~default/{archetype,pagetype}-{view}.{html,tmpl}
		layouts/~default/{archetype,pagetype}.{html,tmpl}
		layouts/~default/{view}.{html,tmpl}
		layouts/~default/baseof.{html,tmpl}
	*/

	if archetype != "" {
		if pageFormat != "" {
			addLookup(layoutName, archetype, view, pageFormat, "tmpl")
			addLookup(layoutName, archetype, view, pageFormat, "html")
			addLookup(layoutName, archetype, "", pageFormat, "tmpl")
			addLookup(layoutName, archetype, "", pageFormat, "html")
		}
		addLookup(layoutName, archetype, view, "", "tmpl")
		addLookup(layoutName, archetype, view, "", "html")
		addLookup(layoutName, archetype, "", "", "tmpl")
		addLookup(layoutName, archetype, "", "", "html")
	}

	if pageFormat != "" {
		addLookup(layoutName, pagetype, view, pageFormat, "tmpl")
		addLookup(layoutName, pagetype, view, pageFormat, "html")
		addLookup(layoutName, pagetype, "", pageFormat, "tmpl")
		addLookup(layoutName, pagetype, "", pageFormat, "html")
		addLookup(layoutName, "", view, pageFormat, "tmpl")
		addLookup(layoutName, "", view, pageFormat, "html")
		addLookup(layoutName, "", "baseof", pageFormat, "tmpl")
		addLookup(layoutName, "", "baseof", pageFormat, "html")
	}

	addLookup(layoutName, pagetype, view, "", "tmpl")
	addLookup(layoutName, pagetype, view, "", "html")
	addLookup(layoutName, pagetype, "", "", "tmpl")
	addLookup(layoutName, pagetype, "", "", "html")
	addLookup(layoutName, "", view, "", "tmpl")
	addLookup(layoutName, "", view, "", "html")
	addLookup(layoutName, "", "baseof", "", "tmpl")
	addLookup(layoutName, "", "baseof", "", "html")

	if layoutName != globals.DefaultThemeLayoutName {
		if pageFormat != "" {
			if archetype != "" {
				addLookup(globals.DefaultThemeLayoutName, archetype, view, pageFormat, "tmpl")
				addLookup(globals.DefaultThemeLayoutName, archetype, view, pageFormat, "html")
				addLookup(globals.DefaultThemeLayoutName, archetype, "", pageFormat, "tmpl")
				addLookup(globals.DefaultThemeLayoutName, archetype, "", pageFormat, "html")
			}
			addLookup(globals.DefaultThemeLayoutName, pagetype, view, pageFormat, "tmpl")
			addLookup(globals.DefaultThemeLayoutName, pagetype, view, pageFormat, "html")
			addLookup(globals.DefaultThemeLayoutName, pagetype, "", pageFormat, "tmpl")
			addLookup(globals.DefaultThemeLayoutName, pagetype, "", pageFormat, "html")
			addLookup(globals.DefaultThemeLayoutName, "", view, pageFormat, "tmpl")
			addLookup(globals.DefaultThemeLayoutName, "", view, pageFormat, "html")
			addLookup(globals.DefaultThemeLayoutName, "", "baseof", pageFormat, "tmpl")
			addLookup(globals.DefaultThemeLayoutName, "", "baseof", pageFormat, "html")
		}
		if archetype != "" {
			addLookup(globals.DefaultThemeLayoutName, archetype, view, "", "tmpl")
			addLookup(globals.DefaultThemeLayoutName, archetype, view, "", "html")
			addLookup(globals.DefaultThemeLayoutName, archetype, "", "", "tmpl")
			addLookup(globals.DefaultThemeLayoutName, archetype, "", "", "html")
		}
		addLookup(globals.DefaultThemeLayoutName, pagetype, view, "", "tmpl")
		addLookup(globals.DefaultThemeLayoutName, pagetype, view, "", "html")
		addLookup(globals.DefaultThemeLayoutName, pagetype, "", "", "tmpl")
		addLookup(globals.DefaultThemeLayoutName, pagetype, "", "", "html")
		addLookup(globals.DefaultThemeLayoutName, "", view, "", "tmpl")
		addLookup(globals.DefaultThemeLayoutName, "", view, "", "html")
		addLookup(globals.DefaultThemeLayoutName, "", "baseof", "", "tmpl")
		addLookup(globals.DefaultThemeLayoutName, "", "baseof", "", "html")
	}

	return
}
