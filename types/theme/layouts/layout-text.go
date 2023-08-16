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
	"fmt"
	textTemplate "text/template"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/maps"
)

func (l *Layout) NewTextTemplate(enjin feature.Internals, ctx context.Context) (tmpl *textTemplate.Template, err error) {
	if tmpl, err = textTemplate.New(l.name).Parse(`{{/* empty */}}`); err == nil {
		err = l.ApplyTextTemplate(enjin, tmpl, ctx)
	}
	return
}

func (l *Layout) NewTextTemplateFrom(enjin feature.Internals, parent feature.ThemeLayout, ctx context.Context) (tmpl *textTemplate.Template, err error) {
	if parent != nil {
		if tmpl, err = parent.NewTextTemplate(enjin, ctx); err == nil {
			err = l.ApplyTextTemplate(enjin, tmpl, ctx)
		}
	} else {
		tmpl, err = l.NewTextTemplate(enjin, ctx)
	}
	return
}

func (l *Layout) ApplyTextTemplate(enjin feature.Internals, tt *textTemplate.Template, ctx context.Context) (err error) {
	l.RLock()
	defer l.RUnlock()

	tt.Funcs(enjin.MakeFuncMap(ctx).AsTEXT())
	for _, name := range maps.SortedKeys(l.cache) {
		if _, err = tt.New(name).Parse(l.cache[name]); err != nil {
			err = fmt.Errorf("error parsing cached template: %v - %v", name, err)
			return
		} else {
			// log.TraceF("parsed %v into %v", name, tt.Name())
		}
	}
	return
}