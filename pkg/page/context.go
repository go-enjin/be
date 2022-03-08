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

package page

import (
	"strings"

	"github.com/iancoleman/strcase"

	"github.com/go-enjin/be/pkg/context"
	bePath "github.com/go-enjin/be/pkg/path"
)

func CamelizeContextKeys(ctx context.Context) {
	var remove []string
	for k, v := range ctx {
		c := strcase.ToCamel(k)
		if k != c {
			remove = append(remove, k)
			ctx.Set(c, v)
		}
	}
	ctx.DeleteKeys(remove...)
}

func (p *Page) parseContext(ctx context.Context) {
	CamelizeContextKeys(ctx)

	ctx.DeleteKeys("Path", "Content", "Section")

	p.Slug = ctx.String("Slug", p.Slug)
	p.Url = ctx.String("Url", p.Url)
	if p.Url == "" || p.Url[0] != '/' {
		p.Url = "/" + p.Url
	}
	p.Path, p.Section, p.Slug = bePath.GetSectionSlug(p.Url)
	p.Archetype = p.Section
	ctx.Set("Url", p.Url)
	ctx.Set("Slug", p.Slug)

	p.Title = ctx.String("Title", p.Title)
	ctx.Set("Title", p.Title)

	raw := ctx.String("Format", p.Format.String())
	format := Format(strings.ToLower(raw))
	if format.String() != "nil" {
		p.Format = format
	} else {
		p.Format = Html
	}
	ctx.Set("Format", p.Format.String())

	p.Summary = ctx.String("Summary", p.Summary)
	ctx.Set("Summary", p.Summary)

	p.Description = ctx.String("Description", p.Description)
	ctx.Set("Description", p.Description)

	p.Layout = ctx.String("Layout", p.Layout)
	ctx.Set("Layout", p.Layout)

	p.Archetype = ctx.String("Archetype", p.Archetype)
	ctx.Set("Archetype", p.Archetype)

	p.Language = ctx.String("Language", p.Language)
	ctx.Set("Language", p.Language)

	// context content is not "source" content, do not populate "from" context,
	// only set it so that it's current
	ctx.Set("Content", p.Content)

	// section cannot be set from front-matter
	ctx.Set("Section", p.Section)

	// path cannot be set from front-matter
	ctx.Set("Path", p.Path)

	p.Context.Apply(ctx)
}