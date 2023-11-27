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

package layouts

import (
	htmlTemplate "html/template"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
)

func (l *Layouts) NewHtmlTemplate(enjin feature.Internals, name string, ctx context.Context) (tmpl *htmlTemplate.Template, err error) {

	if tmpl, err = htmlTemplate.New(name).Parse(`{{/* empty */}}`); err == nil {
		err = l.ApplyHtmlTemplates(enjin, tmpl, ctx)
	}

	return
}

func (l *Layouts) ApplyHtmlTemplates(enjin feature.Internals, tmpl *htmlTemplate.Template, ctx context.Context) (err error) {

	if partials := l.GetLayout(globals.PartialThemeLayoutName); partials != nil {
		if err = partials.ApplyHtmlTemplate(enjin, tmpl, ctx); err != nil {
			return
		}
	}

	if _default := l.GetLayout(globals.DefaultThemeLayoutName); _default != nil {
		if err = _default.ApplyHtmlTemplate(enjin, tmpl, ctx); err != nil {
			return
		}
	}

	for _, layoutName := range maps.SortedKeys(l.cache) {
		switch layoutName {
		case globals.PartialThemeLayoutName, globals.DefaultThemeLayoutName:
			continue
		}

		if layout, ok := l.cache[layoutName]; ok {
			if err = layout.ApplyHtmlTemplate(enjin, tmpl, ctx); err != nil {
				return
			}
		} else {
			log.ErrorF("inconsistent cache key: %v", layoutName)
		}
	}

	return
}
