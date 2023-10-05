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

package feature

import (
	"database/sql"
	"time"

	"github.com/gofrs/uuid"

	"github.com/go-enjin/golang-org-x-text/language"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/types/page/matter"
)

type Page interface {
	Type() (pageType string)
	Format() (pageFormat string)
	Url() (url string)
	Slug() (slug string)
	Path() (path string)

	Title() (title string)
	Description() (description string)

	Layout() (layoutName string)
	Section() (section string)
	Archetype() (archetype string)

	Permalink() (id uuid.UUID)
	PermalinkSha() (shasum string)

	Language() (code string)
	LanguageTag() (tag language.Tag)
	Translates() (url string)

	Shasum() (shasum string)
	Content() (content string)
	FrontMatter() (frontMatter string)
	FrontMatterType() (frontMatterType matter.FrontMatterType)
	PageMatter() (pageMatter *matter.PageMatter)

	CreatedAt() (at time.Time)
	UpdatedAt() (at time.Time)
	DeletedAt() (at sql.NullTime)

	Context() (ctx beContext.Context)

	String() (jsonPage string)

	Match(path string) (found string, ok bool)
	MatchQL(query string) (ok bool, err error)
	MatchPrefix(prefix string) (found string, ok bool)
	Redirections() (redirects []string)
	IsRedirection(path string) (ok bool)
	IsTranslation(path string) (ok bool)
	HasTranslation() (ok bool)

	SetType(pageType string)
	SetFormat(pageFormat string) (path string)
	SetSlugUrl(url string)

	SetTitle(title string)
	SetDescription(description string)

	SetLayout(layoutName string)
	SetArchetype(archetype string)

	SetPermalink(value string) (err error)

	SetLanguage(tag language.Tag)
	SetTranslates(url string)

	SetContent(content string)

	SetFrontMatter(frontMatter string)
	SetFrontMatterType(frontMatterType matter.FrontMatterType)

	SetCreatedAt(at time.Time)
	SetUpdatedAt(at time.Time)
	SetDeletedAt(at sql.NullTime)

	Copy() (copy Page)
}