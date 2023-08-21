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
)

func prepareNewTemplateVars(t feature.Theme, ctx context.Context) (ctxLayout, parentLayout feature.ThemeLayout, layoutName, pagetype, archetype, pageFormat string, err error) {
	var ok bool
	layoutName = ctx.String("Layout", "_default")
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

	return
}