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
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/uuid"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	beStrings "github.com/go-enjin/be/pkg/strings"
	"github.com/go-enjin/be/types/page/matter"
)

var (
	_ feature.Page = (*CPage)(nil)
)

type cPageData struct {
	Type   string `json:"type"`
	Format string `json:"format"`

	Url  string `json:"url"`
	Slug string `json:"slug"`
	Path string `json:"path"`

	Title       string `json:"title"`
	Description string `json:"description"`

	Layout    string `json:"layout"`
	Section   string `json:"section"`
	Archetype string `json:"archetype"`

	Permalink    uuid.UUID `json:"permalink"`
	PermalinkSha string    `json:"permalink-sha"`

	Language    string       `json:"language"`
	Translates  string       `json:"translates"`
	LanguageTag language.Tag `json:"language-tag"`

	Shasum          string                 `json:"shasum"`
	Content         string                 `json:"content"`
	FrontMatter     string                 `json:"frontMatter"`
	FrontMatterType matter.FrontMatterType `json:"front-matter-type"`
	PageMatter      *matter.PageMatter     `json:"page-matter"`

	CreatedAt time.Time    `json:"created"`
	UpdatedAt time.Time    `json:"updated"`
	DeletedAt sql.NullTime `json:"deleted"`

	Context context.Context `json:"Context"`

	Formats feature.PageFormatProvider `json:"-"`
}

type CPage struct {
	fields cPageData

	mutable bool
	copied  int

	sync.RWMutex
}

func New(origin string, path, raw string, created, updated int64, formats feature.PageFormatProvider, enjin context.Context) (p feature.Page, err error) {
	var pm *matter.PageMatter
	if pm, err = matter.ParsePageMatter(origin, path, time.Unix(created, 0), time.Unix(updated, 0), []byte(raw)); err != nil {
		return
	}
	p, err = NewFromPageMatter(pm, formats, enjin)
	return
}

func NewFromPageMatter(pm *matter.PageMatter, formats feature.PageFormatProvider, enjin context.Context) (p *CPage, err error) {
	pg := &CPage{
		fields: cPageData{
			PageMatter:  pm,
			Content:     pm.Body,
			Formats:     formats,
			Context:     enjin.Copy(),
			Shasum:      pm.Shasum,
			Language:    pm.Locale.String(),
			LanguageTag: pm.Locale,
			Permalink:   uuid.Nil,
			CreatedAt:   pm.Created,
			UpdatedAt:   pm.Updated,
		},
	}

	path := pg.SetFormat(pm.Path)
	pg.SetSlugUrl(path)
	pg.fields.Title = beStrings.TitleCase(strings.Join(strings.Split(pg.fields.Slug, "-"), " "))

	if err = pg.initFrontMatter(); err != nil {
		return
	}

	if format := formats.GetFormat(pg.fields.Format); format != nil {
		if ctx, e := format.Prepare(pg.fields.Context, pg.fields.Content); e != nil {
			err = e
			return
		} else if ctx != nil {
			if v, ok := ctx.Get("Url").(string); ok {
				if v != pg.fields.Url {
					pg.SetSlugUrl(v)
				}
			}
			if v, ok := ctx.Get("Title").(string); ok {
				pg.fields.Title = v
			}
			if v, ok := ctx.Get("Description").(string); ok {
				pg.fields.Description = v
			}
			pg.fields.Context.Apply(ctx)
		}
	}

	p = pg
	return
}

func NewMatterFromPage(p feature.Page) (pm *matter.PageMatter, err error) {
	pmCtx := context.Context{}
	for key, value := range p.PageMatter().Matter {
		if v, ok := p.Context()[key]; ok {
			pmCtx[key] = v
		} else {
			pmCtx[key] = value
		}
	}
	if p.Path() == "" {
		p.SetSlugUrl("/")
	}
	if !strings.HasPrefix(p.PageMatter().Path, p.Path()) {
		log.ErrorF("detected important inconsistency, page Path is not a prefix of PageMatter.Path: %#+v", p)
	}
	if _, exists := pmCtx["Url"]; !exists {
		pmCtx["Url"] = p.Url()
	}
	if _, exists := pmCtx["Path"]; !exists {
		pmCtx["Path"] = p.Path()
	}
	if _, exists := pmCtx["Section"]; !exists {
		pmCtx["Section"] = p.Section()
	}
	if _, exists := pmCtx["Slug"]; !exists {
		pmCtx["Slug"] = p.Slug()
	}
	if _, exists := pmCtx["Title"]; !exists {
		pmCtx["Title"] = p.Title()
	}
	if _, exists := pmCtx["Language"]; !exists {
		pmCtx["Language"] = p.LanguageTag().String()
	}
	if _, exists := pmCtx["Created"]; !exists {
		pmCtx["Created"] = p.CreatedAt()
	}
	if _, exists := pmCtx["Updated"]; !exists {
		pmCtx["Updated"] = p.CreatedAt()
	}
	pmCtx.Delete("Shasum")

	stanza := matter.MakeStanza(p.FrontMatterType(), pmCtx)
	data := []byte(stanza + "\n" + p.Content())

	pm, err = matter.ParsePageMatter(p.PageMatter().Origin, p.Path(), p.CreatedAt(), p.UpdatedAt(), data)
	return
}

func NewPageFromStub(ps *feature.PageStub, formats feature.PageFormatProvider) (p feature.Page, err error) {
	var data []byte
	if data, err = ps.FS.ReadFile(ps.Source); err != nil {
		err = fmt.Errorf("error reading %v mount file: %v - %v", ps.FS.Name(), ps.Source, err)
		return
	}

	//path := beStrings.TrimPrefixes(ps.Source, ps.Fallback.String())
	var epoch, created, updated int64

	if epoch, err = ps.FS.FileCreated(ps.Source); err == nil {
		created = epoch
	} else {
		log.ErrorF("error getting page created epoch: %v", err)
	}

	if epoch, err = ps.FS.LastModified(ps.Source); err == nil {
		updated = epoch
	} else {
		log.ErrorF("error getting page last modified epoch: %v", err)
	}

	if p, err = New(ps.Origin, ps.Source, string(data), created, updated, formats, ps.EnjinCtx); err == nil {
		if language.Compare(p.LanguageTag(), language.Und) {
			p.SetLanguage(ps.Fallback)
		}
		if !strings.HasPrefix(p.Url(), "!") {
			p.SetSlugUrl(filepath.Clean(ps.Point + p.Url()))
		}
		// log.DebugF("made page from %v stub: [%v] %v (%v)", s.FS.Name(), p.Language, s.Source, p.Url)
	} else {
		err = fmt.Errorf("error: new %v mount page %v - %v", ps.FS.Name(), ps.Source, err)
	}
	return
}

func (p *CPage) Copy() (copy feature.Page) {
	if p.copied > 0 {
		p.copied += 1
		return p
	}
	pg := &CPage{
		fields: cPageData{
			Type:         p.fields.Type,
			Url:          p.fields.Url,
			Slug:         p.fields.Slug,
			Path:         p.fields.Path,
			Title:        p.fields.Title,
			Format:       p.fields.Format,
			Description:  p.fields.Description,
			Layout:       p.fields.Layout,
			Section:      p.fields.Section,
			Archetype:    p.fields.Archetype,
			FrontMatter:  p.fields.FrontMatter,
			Language:     p.fields.Language,
			LanguageTag:  p.fields.LanguageTag,
			Translates:   p.fields.Translates,
			Permalink:    p.fields.Permalink,
			PermalinkSha: p.fields.PermalinkSha,
			Content:      p.fields.Content,
			Formats:      p.fields.Formats,
			Context:      context.New(),
			CreatedAt:    p.fields.CreatedAt,
			UpdatedAt:    p.fields.UpdatedAt,
			DeletedAt:    p.fields.DeletedAt,
			PageMatter:   p.fields.PageMatter.Copy(),
		},
		copied:  1,
		mutable: p.mutable,
	}
	// log.WarnDF(1, "copied page: %v", p.Url)
	_ = pg.initFrontMatter()
	copy = pg
	return
}