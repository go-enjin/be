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
	"time"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/log"
)

func (p *Page) parseContext(ctx context.Context) {
	ctx.CamelizeKeys()
	ctx.DeleteKeys("Path", "Section", "Slug", "Content")

	p.Type = ctx.String("Type", "page")

	if ctxLang := ctx.String("Language", ""); ctxLang != "" {
		if tag, err := language.Parse(ctxLang); err == nil {
			p.SetLanguage(tag)
		} else {
			p.SetLanguage(language.Und)
		}
	} else {
		p.SetLanguage(language.Und)
	}

	if ctxTranslates := ctx.String("Translates", ""); ctxTranslates != "" {
		p.Translates = ctxTranslates
	}

	p.SetSlugUrl(ctx.String("Url", p.Url))

	if ctxPermalinkId := ctx.String("Permalink", ""); ctxPermalinkId != "" {
		if err := p.SetPermalink(ctxPermalinkId); err != nil {
			log.ErrorF("error setting permalink: %v - %v", err)
			_ = p.SetPermalink("")
		}
	} else {
		_ = p.SetPermalink("")
	}

	p.Title = ctx.String("Title", p.Title)
	ctx.Set("Title", p.Title)

	p.Summary = ctx.String("Summary", p.Summary)
	ctx.Set("Summary", p.Summary)

	p.Description = ctx.String("Description", p.Description)
	ctx.Set("Description", p.Description)

	p.Layout = ctx.String("Layout", p.Layout)
	ctx.Set("Layout", p.Layout)

	p.Archetype = ctx.String("Archetype", p.Archetype)
	ctx.Set("Archetype", p.Archetype)
	if format := ctx.String("Format", ""); format != "" {
		p.Format = format
	}

	if created := ctx.String("Created", ""); created != "" {
		// 2020-05-01T12:55:05-04:00
		if parsed, err := time.Parse(time.RFC3339, created); err == nil {
			p.CreatedAt = parsed
		} else if parsed, err = time.Parse(time.RFC1123Z, created); err == nil {
			p.CreatedAt = parsed
		} else if parsed, err = time.Parse(time.RFC1123, created); err == nil {
			p.CreatedAt = parsed
		} else {
			log.ErrorF("error parsing unsupported time/date format: %v", created)
		}
	}

	if updated := ctx.String("Updated", ""); updated != "" {
		// 2020-05-01T12:55:05-04:00
		if parsed, err := time.Parse(time.RFC3339, updated); err == nil {
			p.UpdatedAt = parsed
		} else if parsed, err = time.Parse(time.RFC1123Z, updated); err == nil {
			p.UpdatedAt = parsed
		} else if parsed, err = time.Parse(time.RFC1123, updated); err == nil {
			p.UpdatedAt = parsed
		} else {
			log.ErrorF("error parsing unsupported time/date format: %v", updated)
		}
	}

	p.Context.Apply(ctx)
}