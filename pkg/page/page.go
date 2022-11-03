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
	"encoding/json"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/iancoleman/strcase"
	"gorm.io/gorm"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/lang"
	beStrings "github.com/go-enjin/be/pkg/strings"

	"github.com/go-enjin/be/pkg/context"
	bePath "github.com/go-enjin/be/pkg/path"
)

type Page struct {
	gorm.Model

	Url         string `json:"url" gorm:"index"`
	Slug        string `json:"slug"`
	Path        string `json:"path"`
	Title       string `json:"title" gorm:"index"`
	Format      string `json:"format" gorm:"type:string"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
	Layout      string `json:"layout"`
	Section     string `json:"section"`
	Archetype   string `json:"archetype"`
	FrontMatter string `json:"frontMatter"`
	Language    string `json:"language"`
	Translates  string `json:"translates"`
	Content     string `json:"content"`

	Initial context.Context `json:"-" gorm:"-"`
	Context context.Context `json:"context" gorm:"-"`

	LanguageTag language.Tag `json:"-"`

	Permalink    uuid.UUID `json:"permalink"`
	PermalinkSha string    `json:"-"`
}

func New(path, raw string, created, updated int64) (p *Page, err error) {
	p = new(Page)
	p.Initial = context.New()
	p.Context = context.New()

	p.Permalink = uuid.Nil

	path = forms.TrimQueryParams(path)
	p.SetSlugUrl(path)

	p.CreatedAt = time.Unix(created, 0)
	p.Initial.SetSpecific("Created", p.CreatedAt)

	p.UpdatedAt = time.Unix(updated, 0)
	p.Initial.SetSpecific("Updated", p.UpdatedAt)

	// log.DebugF("new page for path: %v - %v - %v", path, slug, p.Url)
	p.Title = beStrings.TitleCase(strings.Join(strings.Split(p.Slug, "-"), " "))
	p.Initial.Set("Url", p.Url)
	p.Initial.Set("Slug", p.Slug)
	p.Initial.Set("Title", p.Title)

	raw = lang.StripTranslatorComments(raw)
	if !p.parseYaml(raw) {
		if !p.parseToml(raw) {
			if !p.parseJson(raw) {
				p.Content = raw
				p.parseContext(p.Initial)
			}
		}
	}
	return
}

func (p *Page) String() string {
	ctx, _ := json.MarshalIndent(p.Context, "", "    ")
	return string(ctx) + "\n" + p.Content
}

func (p *Page) Copy() (copy *Page) {
	copy = &Page{
		Url:          p.Url,
		Slug:         p.Slug,
		Path:         p.Path,
		Title:        p.Title,
		Format:       p.Format,
		Summary:      p.Summary,
		Description:  p.Description,
		Layout:       p.Layout,
		Section:      p.Section,
		Archetype:    p.Archetype,
		FrontMatter:  p.FrontMatter,
		Language:     p.Language,
		LanguageTag:  p.LanguageTag,
		Translates:   p.Translates,
		Permalink:    p.Permalink,
		PermalinkSha: p.PermalinkSha,
		Content:      p.Content,
	}
	copy.ID = p.ID
	copy.CreatedAt = p.CreatedAt
	copy.UpdatedAt = p.UpdatedAt
	copy.DeletedAt = p.DeletedAt
	copy.Initial = p.Initial.Copy()
	copy.Context = p.Initial.Copy()
	return
}

func (p *Page) SetLanguage(tag language.Tag) {
	p.LanguageTag = tag
	p.Language = p.LanguageTag.String()
	p.Context.Set("Language", p.Language)
}

func (p *Page) SetSlugUrl(path string) {
	trimmedPath := bePath.TrimSlashes(path)

	var slug, urlPath string
	if f, e := MatchFormatExtension(trimmedPath); f != nil {
		p.Format = f.Name()
		urlPath = strings.TrimSuffix(trimmedPath, "."+e)
		slug = filepath.Base(urlPath)
	} else {
		p.Format = "html"
		urlPath = trimmedPath
		slug = filepath.Base(slug)
	}

	dirPath := bePath.Dir(urlPath)
	if dirPath != "." {
		urlPath = strings.ToLower(dirPath)
	} else {
		urlPath = ""
	}

	p.Slug = strcase.ToKebab(slug)

	if urlPath != "" {
		p.Url = "/" + urlPath + "/" + p.Slug
	} else {
		p.Url = "/" + p.Slug
	}
}