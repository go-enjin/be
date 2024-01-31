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
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"

	"github.com/go-corelibs/x-text/language"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/types/page/matter"
)

func (p *CPage) Type() (pageType string) {
	pageType = p.fields.Type
	return
}

func (p *CPage) Format() (pageFormat string) {
	pageFormat = p.fields.Format
	return
}

func (p *CPage) Url() (url string) {
	url = p.fields.Url
	return
}

func (p *CPage) Slug() (slug string) {
	slug = p.fields.Slug
	return
}

func (p *CPage) Path() (path string) {
	path = p.fields.Path
	return
}

func (p *CPage) Title() (title string) {
	title = p.fields.Title
	return
}

func (p *CPage) Description() (description string) {
	description = p.fields.Description
	return
}

func (p *CPage) Layout() (layoutName string) {
	layoutName = p.fields.Layout
	return
}

func (p *CPage) Section() (section string) {
	section = p.fields.Section
	return
}

func (p *CPage) Archetype() (archetype string) {
	archetype = p.fields.Archetype
	return
}

func (p *CPage) Permalink() (id uuid.UUID) {
	id = p.fields.Permalink
	return
}

func (p *CPage) PermalinkSha() (shasum string) {
	shasum = p.fields.PermalinkSha
	return
}

func (p *CPage) Language() (code string) {
	code = p.fields.Language
	return
}

func (p *CPage) LanguageTag() (tag language.Tag) {
	tag = p.fields.LanguageTag
	return
}

func (p *CPage) Translates() (url string) {
	url = p.fields.Translates
	return
}

func (p *CPage) Shasum() (shasum string) {
	shasum = p.fields.Shasum
	return
}

func (p *CPage) Content() (content string) {
	content = p.fields.Content
	return
}

func (p *CPage) FrontMatter() (frontMatter string) {
	frontMatter = p.fields.FrontMatter
	return
}

func (p *CPage) FrontMatterType() (frontMatterType matter.FrontMatterType) {
	frontMatterType = p.fields.FrontMatterType
	return
}

func (p *CPage) PageMatter() (pageMatter *matter.PageMatter) {
	pageMatter = p.fields.PageMatter
	return
}

func (p *CPage) CreatedAt() (at time.Time) {
	at = p.fields.CreatedAt
	return
}

func (p *CPage) UpdatedAt() (at time.Time) {
	at = p.fields.UpdatedAt
	return
}

func (p *CPage) DeletedAt() (at sql.NullTime) {
	at = p.fields.DeletedAt
	return
}

func (p *CPage) Context() (ctx context.Context) {
	ctx = p.fields.Context
	return
}

func (p *CPage) String() (jsonPage string) {
	ctx, _ := json.MarshalIndent(p.Context(), "", "    ")
	return "{{" + string(ctx) + "}}" + "\n" + p.fields.Content
}
