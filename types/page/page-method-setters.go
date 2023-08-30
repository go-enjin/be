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
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/go-enjin/golang-org-x-text/language"
	"github.com/gofrs/uuid"

	"github.com/go-enjin/be/pkg/hash/sha"
	"github.com/go-enjin/be/pkg/pages"
	"github.com/go-enjin/be/types/page/matter"
)

func (p *CPage) SetType(pageType string) {
	p.fields.Type = pageType
	p.fields.PageMatter.Matter.SetSpecific("type", pageType)
	p.fields.Context.SetSpecific("Type", pageType)
}

func (p *CPage) SetFormat(filepath string) (path string) {
	if format, match := p.fields.Formats.MatchFormat(filepath); format != nil {
		p.fields.Format = match
		path = strings.TrimSuffix(filepath, "."+match)
	} else {
		p.fields.Format = "tmpl"
		path = strings.TrimSuffix(filepath, ".tmpl")
	}
	p.fields.Context.SetSpecific("Format", p.fields.Format)
	return
}

func (p *CPage) SetSlugUrl(filepath string) {
	p.fields.Url, p.fields.Path, p.fields.Section, p.fields.Slug = pages.GetUrlPathSectionSlug(filepath)
	p.fields.Context.SetSpecific("Url", p.fields.Url)
	p.fields.Context.SetSpecific("Path", p.fields.Path)
	p.fields.Context.SetSpecific("Section", p.fields.Section)
	p.fields.Context.SetSpecific("Slug", p.fields.Slug)
}

func (p *CPage) SetTitle(title string) {
	p.fields.Title = title
	p.fields.Context.SetSpecific("Title", title)
}

func (p *CPage) SetSummary(summary string) {
	p.fields.Summary = summary
	p.fields.Context.SetSpecific("Summary", summary)
}

func (p *CPage) SetDescription(description string) {
	p.fields.Description = description
	p.fields.Context.SetSpecific("Description", description)
}

func (p *CPage) SetLayout(layoutName string) {
	p.fields.Layout = layoutName
	p.fields.Context.SetSpecific("Layout", layoutName)
}

func (p *CPage) SetArchetype(archetype string) {
	p.fields.Archetype = archetype
	p.fields.Context.SetSpecific("Archetype", archetype)
}

func (p *CPage) SetPermalink(value string) (err error) {
	if value == "" {
		p.fields.Context.SetSpecific("LongLink", p.fields.Url)
		p.fields.Context.SetSpecific("ShortLink", p.fields.Url)
		p.fields.Context.SetSpecific("Permalink", "")
		p.fields.Context.SetSpecific("Permalinked", false)
		p.fields.Context.SetSpecific("PermalinkUrl", p.fields.Url)
		p.fields.Context.SetSpecific("PermalinkSha", "")
		p.fields.Context.SetSpecific("PermalinkLongUrl", p.fields.Url)
		return
	}

	if id, e := uuid.FromString(value); e != nil {
		err = fmt.Errorf("error parsing permalink id: %v - %v", value, e)
		return
	} else if sum, ee := sha.DataHash10(id.Bytes()); ee != nil {
		err = fmt.Errorf("error getting permalink sha: %v - %v", id, ee)
		return
	} else {
		p.fields.Permalink = id
		p.fields.PermalinkSha = sum
		p.fields.Context.SetSpecific("LongLink", "/"+p.fields.Permalink.String())
		p.fields.Context.SetSpecific("ShortLink", "/"+p.fields.PermalinkSha)
		p.fields.Context.SetSpecific("Permalink", id)
		p.fields.Context.SetSpecific("Permalinked", true)
		p.fields.Context.SetSpecific("PermalinkSha", sum)
		if p.fields.Url == "/" {
			p.fields.Context.SetSpecific("PermalinkUrl", p.fields.Url+p.fields.PermalinkSha)
		} else {
			p.fields.Context.SetSpecific("PermalinkUrl", p.fields.Url+"-"+p.fields.PermalinkSha)
		}
	}
	return
}

func (p *CPage) SetLanguage(tag language.Tag) {
	p.fields.LanguageTag = tag
	p.fields.Language = p.LanguageTag().String()
	p.fields.Context.Set("Language", p.fields.Language)
}

func (p *CPage) SetTranslates(url string) {
	p.fields.Translates = url
	p.fields.Context.SetSpecific("Translates", url)
}

func (p *CPage) SetContent(content string) {
	p.fields.Content = content
}

func (p *CPage) SetFrontMatter(frontMatter string) {
	p.fields.FrontMatter = frontMatter
}

func (p *CPage) SetFrontMatterType(frontMatterType matter.FrontMatterType) {
	p.fields.FrontMatterType = frontMatterType
}

func (p *CPage) SetCreatedAt(at time.Time) {
	p.fields.CreatedAt = at
	p.fields.Context.SetSpecific("Created", at)
}

func (p *CPage) SetUpdatedAt(at time.Time) {
	p.fields.UpdatedAt = at
	p.fields.Context.SetSpecific("Updated", at)
}

func (p *CPage) SetDeletedAt(at sql.NullTime) {
	p.fields.DeletedAt = at
	p.fields.Context.SetSpecific("Deleted", at)
}