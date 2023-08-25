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

package theme

import (
	htmlTemplate "html/template"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
)

func (t *CTheme) NewHtmlTemplate(enjin feature.Internals, name string, ctx context.Context) (tmpl *htmlTemplate.Template, err error) {
	tmpl = htmlTemplate.New(name).Funcs(enjin.MakeFuncMap(ctx).AsHTML())
	if parent := t.GetParent(); parent != nil {
		if err = parent.Layouts().ApplyHtmlTemplates(enjin, tmpl, ctx); err != nil {
			return
		}
	}
	err = t.Layouts().ApplyHtmlTemplates(enjin, tmpl, ctx)
	return
}