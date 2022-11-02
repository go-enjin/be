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

	"github.com/go-enjin/golang-org-x-text/language"
	"github.com/gofrs/uuid"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/hash/sha"
	"github.com/go-enjin/be/pkg/log"
	bePath "github.com/go-enjin/be/pkg/path"
)

func (p *Page) parseContext(ctx context.Context) {
	ctx.CamelizeKeys()

	ctx.DeleteKeys("Path", "Content", "Section")

	p.Url = ctx.String("Url", p.Url)
	if p.Url == "" || p.Url[0] != '/' {
		p.Url = "/" + p.Url
	}
	p.Path, p.Section, p.Slug = bePath.GetSectionSlug(p.Url)
	p.Archetype = p.Section
	ctx.Set("Url", p.Url)
	ctx.Set("Slug", p.Slug)

	p.SetLanguage(language.Und)
	if ctxLang := ctx.String("Language", ""); ctxLang != "" {
		if tag, err := language.Parse(ctxLang); err == nil {
			p.SetLanguage(tag)
		}
	}

	if ctxTranslates := ctx.String("Translates", ""); ctxTranslates != "" {
		p.Translates = ctxTranslates
	}

	if ctxPermalinkId := ctx.String("Permalink", ""); ctxPermalinkId != "" {
		if id, e := uuid.FromString(ctxPermalinkId); e != nil {
			log.ErrorF("error parsing permalink id: %v - %v", ctxPermalinkId, e)
		} else if sum, ee := sha.DataHash10([]byte(ctxPermalinkId)); ee != nil {
			log.ErrorF("error getting permalink sha: %v - %v", id, ee)
		} else {
			p.Permalink = id
			p.PermalinkSha = sum
			ctx.SetSpecific("Permalink", id)
			ctx.SetSpecific("PermalinkSha", sum)
			ctx.SetSpecific("Permalinked", true)
		}
	}

	p.Title = ctx.String("Title", p.Title)
	ctx.Set("Title", p.Title)

	raw := ctx.String("Format", p.Format)
	p.Format = strings.ToLower(raw)
	ctx.Set("Format", p.Format)

	p.Summary = ctx.String("Summary", p.Summary)
	ctx.Set("Summary", p.Summary)

	p.Description = ctx.String("Description", p.Description)
	ctx.Set("Description", p.Description)

	p.Layout = ctx.String("Layout", p.Layout)
	ctx.Set("Layout", p.Layout)

	p.Archetype = ctx.String("Archetype", p.Archetype)
	ctx.Set("Archetype", p.Archetype)

	// context content is not "source" content, do not populate "from" context,
	// only set it so that it's current
	ctx.Set("Content", p.Content)

	// section cannot be set from front-matter
	ctx.Set("Section", p.Section)

	// path cannot be set from front-matter
	ctx.Set("Path", p.Path)

	p.Initial.Apply(ctx)
	p.Context.Apply(p.Initial)
}