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

package page

import (
	"fmt"
	"time"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/types/page/matter"
)

func (p *CPage) initFrontMatter() (err error) {
	p.fields.Context.SetSpecific("Url", p.fields.Url)
	p.fields.Context.SetSpecific("Path", p.fields.Path)
	p.fields.Context.SetSpecific("Section", p.fields.Section)
	p.fields.Context.SetSpecific("Slug", p.fields.Slug)
	p.fields.Context.SetSpecific("Title", p.fields.Title)
	p.fields.Context.SetSpecific("Shasum", p.fields.Shasum)
	p.fields.Context.SetSpecific("Created", p.fields.CreatedAt)
	p.fields.Context.SetSpecific("Updated", p.fields.UpdatedAt)
	p.fields.Context.SetSpecific("Deleted", p.fields.DeletedAt)

	if p.fields.PageMatter != nil {
		p.parseContext(p.fields.PageMatter.Matter)
		return
	}

	switch p.fields.FrontMatterType {
	case matter.TomlMatter:
		if ctx, ee := context.ParseToml(p.fields.FrontMatter); ee != nil {
			err = fmt.Errorf("error parsing page toml front matter: %v", ee)
		} else {
			p.parseContext(ctx)
		}
	case matter.YamlMatter:
		if ctx, ee := context.ParseYaml(p.fields.FrontMatter); ee != nil {
			err = fmt.Errorf("error parsing page yaml front matter: %v", ee)
		} else {
			p.parseContext(ctx)
		}
	case matter.JsonMatter:
		if ctx, ee := context.ParseJson(p.fields.FrontMatter); ee != nil {
			err = fmt.Errorf("error parsing page json front matter: %v", ee)
		} else {
			p.parseContext(ctx)
		}
	default:
		p.parseContext(context.New())
	}
	return
}

func (p *CPage) parseContext(ctx context.Context) {
	ctx.CamelizeKeys()
	ctx.DeleteKeys("Path", "Section", "Slug", "Content")

	p.fields.Type = ctx.String("Type", "page")

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
		p.fields.Translates = ctxTranslates
	}

	p.SetSlugUrl(ctx.String("Url", p.fields.Url))

	if ctxPermalinkId := ctx.String("Permalink", ""); ctxPermalinkId != "" {
		if err := p.SetPermalink(ctxPermalinkId); err != nil {
			log.ErrorF("error setting permalink: %v - %v", err)
			_ = p.SetPermalink("")
		}
	} else {
		_ = p.SetPermalink("")
	}

	p.fields.Title = ctx.String("Title", p.fields.Title)
	ctx.Set("Title", p.fields.Title)

	p.fields.Summary = ctx.String("Summary", p.fields.Summary)
	ctx.Set("Summary", p.fields.Summary)

	p.fields.Description = ctx.String("Description", p.fields.Description)
	ctx.Set("Description", p.fields.Description)

	p.fields.Layout = ctx.String("Layout", p.fields.Layout)
	ctx.Set("Layout", p.fields.Layout)

	p.fields.Archetype = ctx.String("Archetype", p.fields.Archetype)
	ctx.Set("Archetype", p.fields.Archetype)
	if format := ctx.String("Format", ""); format != "" {
		p.fields.Format = format
	}

	if created := ctx.String("Created", ""); created != "" {
		// 2020-05-01T12:55:05-04:00
		if parsed, err := time.Parse(time.RFC3339, created); err == nil {
			p.fields.CreatedAt = parsed
		} else if parsed, err = time.Parse(time.RFC1123Z, created); err == nil {
			p.fields.CreatedAt = parsed
		} else if parsed, err = time.Parse(time.RFC1123, created); err == nil {
			p.fields.CreatedAt = parsed
		} else {
			log.ErrorF("error parsing unsupported time/date format: %v", created)
		}
	}

	if updated := ctx.String("Updated", ""); updated != "" {
		// 2020-05-01T12:55:05-04:00
		if parsed, err := time.Parse(time.RFC3339, updated); err == nil {
			p.fields.UpdatedAt = parsed
		} else if parsed, err = time.Parse(time.RFC1123Z, updated); err == nil {
			p.fields.UpdatedAt = parsed
		} else if parsed, err = time.Parse(time.RFC1123, updated); err == nil {
			p.fields.UpdatedAt = parsed
		} else {
			log.ErrorF("error parsing unsupported time/date format: %v", updated)
		}
	}

	if deleted := ctx.String("Deleted", ""); deleted != "" {
		// 2020-05-01T12:55:05-04:00
		if parsed, err := time.Parse(time.RFC3339, deleted); err == nil {
			p.fields.DeletedAt.Time = parsed
			p.fields.DeletedAt.Valid = true
		} else if parsed, err = time.Parse(time.RFC1123Z, deleted); err == nil {
			p.fields.DeletedAt.Time = parsed
			p.fields.DeletedAt.Valid = true
		} else if parsed, err = time.Parse(time.RFC1123, deleted); err == nil {
			p.fields.DeletedAt.Time = parsed
			p.fields.DeletedAt.Valid = true
		} else {
			log.ErrorF("error parsing unsupported time/date format: %v", deleted)
		}
	}

	p.fields.Context.Apply(ctx)
}