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
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/iancoleman/strcase"
	"gorm.io/gorm"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/hash/sha"
	"github.com/go-enjin/be/pkg/page/matter"
	bePath "github.com/go-enjin/be/pkg/path"
	beStrings "github.com/go-enjin/be/pkg/strings"
	types "github.com/go-enjin/be/pkg/types/theme-types"
)

type Page struct {
	Type   string `json:"type" gorm:"type"`
	Format string `json:"format" gorm:"type:string"`

	Url  string `json:"url" gorm:"index"`
	Slug string `json:"slug"`
	Path string `json:"path"`

	Title       string `json:"title" gorm:"index"`
	Summary     string `json:"summary"`
	Description string `json:"description"`

	Layout    string `json:"layout"`
	Section   string `json:"section"`
	Archetype string `json:"archetype"`

	Permalink    uuid.UUID `json:"permalink"`
	PermalinkSha string    `json:"-" gorm:"-"`

	Language    string       `json:"language"`
	Translates  string       `json:"translates"`
	LanguageTag language.Tag `json:"-" gorm:"-"`

	Shasum          string `json:"shasum"`
	Content         string `json:"content"`
	FrontMatter     string `json:"frontMatter"`
	FrontMatterType matter.FrontMatterType
	PageMatter      *matter.PageMatter

	Context context.Context `json:"-" gorm:"-"`

	Formats types.FormatProvider
	copied  int

	gorm.Model
}

func NewFromFile(path, file string, formats types.FormatProvider, enjin context.Context) (p *Page, err error) {
	if !bePath.IsFile(file) {
		err = fmt.Errorf("not a file: %v", file)
		return
	}
	var contents []byte
	if contents, err = bePath.ReadFile(file); err != nil {
		return
	}
	var created, updated int64
	if spec, e := bePath.Stat(file); e == nil {
		if spec.HasBirthTime() {
			created = spec.BirthTime().Unix()
		}
		updated = spec.ModTime().Unix()
	}
	p, err = New(path, string(contents), created, updated, formats, enjin)
	return
}

func New(path, raw string, created, updated int64, formats types.FormatProvider, enjin context.Context) (p *Page, err error) {
	var pm *matter.PageMatter
	if pm, err = matter.ParsePageMatter(path, time.Unix(created, 0), time.Unix(updated, 0), []byte(raw)); err != nil {
		return
	}
	p, err = NewFromPageMatter(pm, formats, enjin)
	return
}

func NewFromPageMatter(pm *matter.PageMatter, formats types.FormatProvider, enjin context.Context) (p *Page, err error) {
	p = new(Page)
	p.PageMatter = pm
	p.Content = pm.Body
	p.Formats = formats
	p.Context = enjin.Copy()

	p.Shasum = pm.Shasum
	p.Permalink = uuid.Nil

	path := p.SetFormat(pm.Path)
	p.SetSlugUrl(path)

	p.CreatedAt = pm.Created
	p.UpdatedAt = pm.Updated

	p.Title = beStrings.TitleCase(strings.Join(strings.Split(p.Slug, "-"), " "))

	// TODO: figure out how to do front-matter templating again

	// tt := textTemplate.New("front-matter").Funcs(funcmaps.TextFuncMap())
	// if tt, err = tt.Parse(pm.FrontMatter); err != nil {
	// 	err = fmt.Errorf("error parsing front-matter text tmpl: %v", err)
	// 	return
	// }
	// var buf bytes.Buffer
	// if err = tt.Execute(&buf, p.Context); err != nil {
	// 	err = fmt.Errorf("error parsing front-matter text tmpl: %v", err)
	// 	return
	// }
	// p.FrontMatter = buf.String()

	if err = p.initFrontMatter(); err != nil {
		return
	}

	if format := formats.GetFormat(p.Format); format != nil {
		if ctx, e := format.Prepare(p.Context, p.Content); e != nil {
			err = e
			return
		} else if ctx != nil {
			if v, ok := ctx.Get("Url").(string); ok {
				if v != p.Url {
					p.SetSlugUrl(v)
				}
			}
			if v, ok := ctx.Get("Title").(string); ok {
				p.Title = v
			}
			if v, ok := ctx.Get("Description").(string); ok {
				p.Description = v
			}
			p.Context.Apply(ctx)
		}
	}

	return
}

func (p *Page) String() string {
	ctx, _ := json.MarshalIndent(p.Context, "", "    ")
	return "{{" + string(ctx) + "}}" + "\n" + p.Content
}

func (p *Page) initFrontMatter() (err error) {
	p.Context.SetSpecific("Url", p.Url)
	p.Context.SetSpecific("Slug", p.Slug)
	p.Context.SetSpecific("Title", p.Title)
	p.Context.SetSpecific("Shasum", p.Shasum)
	p.Context.SetSpecific("Created", p.CreatedAt)
	p.Context.SetSpecific("Updated", p.UpdatedAt)

	if p.PageMatter != nil {
		p.parseContext(p.PageMatter.Matter)
		return
	}

	switch p.FrontMatterType {
	case matter.TomlMatter:
		if ctx, ee := matter.ParseToml(p.FrontMatter); ee != nil {
			err = fmt.Errorf("error parsing page toml front matter: %v", ee)
		} else {
			p.parseContext(ctx)
		}
	case matter.YamlMatter:
		if ctx, ee := matter.ParseYaml(p.FrontMatter); ee != nil {
			err = fmt.Errorf("error parsing page yaml front matter: %v", ee)
		} else {
			p.parseContext(ctx)
		}
	case matter.JsonMatter:
		if ctx, ee := matter.ParseJson(p.FrontMatter); ee != nil {
			err = fmt.Errorf("error parsing page json front matter: %v", ee)
		} else {
			p.parseContext(ctx)
		}
	default:
		p.parseContext(context.New())
	}
	return
}

func (p *Page) Copy() (copy *Page) {
	if p.copied > 0 {
		p.copied += 1
		return p
	}
	copy = &Page{
		Type:         p.Type,
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
		Formats:      p.Formats,
		Context:      context.New(),
		copied:       1,
	}
	copy.ID = p.ID
	copy.CreatedAt = p.CreatedAt
	copy.UpdatedAt = p.UpdatedAt
	copy.DeletedAt = p.DeletedAt
	// copy.Context = p.Context.Copy()
	// log.WarnDF(1, "copied page: %v", p.Url)
	_ = copy.initFrontMatter()
	return
}

func (p *Page) SetLanguage(tag language.Tag) {
	p.LanguageTag = tag
	p.Language = p.LanguageTag.String()
	p.Context.Set("Language", p.Language)
}

func (p *Page) SetFormat(filepath string) (path string) {
	if format, match := p.Formats.MatchFormat(filepath); format != nil {
		p.Format = match
		path = strings.TrimSuffix(filepath, "."+match)
	} else {
		p.Format = "tmpl"
		path = strings.TrimSuffix(filepath, ".tmpl")
	}
	p.Context.SetSpecific("Format", p.Format)
	return
}

func (p *Page) SetSlugUrl(filepath string) {
	p.Url, p.Section, p.Slug = p.getUrlPathSectionSlug(filepath)
	p.Context.SetSpecific("Url", p.Url)
	p.Context.SetSpecific("Section", p.Section)
	p.Context.SetSpecific("Slug", p.Slug)
}

func (p *Page) SetPermalink(value string) (err error) {
	if value == "" {
		p.Context.SetSpecific("LongLink", p.Url)
		p.Context.SetSpecific("ShortLink", p.Url)
		p.Context.SetSpecific("Permalink", "")
		p.Context.SetSpecific("Permalinked", false)
		p.Context.SetSpecific("PermalinkUrl", p.Url)
		p.Context.SetSpecific("PermalinkSha", "")
		p.Context.SetSpecific("PermalinkLongUrl", p.Url)
		return
	}

	if id, e := uuid.FromString(value); e != nil {
		err = fmt.Errorf("error parsing permalink id: %v - %v", value, e)
		return
	} else if sum, ee := sha.DataHash10(id.Bytes()); ee != nil {
		err = fmt.Errorf("error getting permalink sha: %v - %v", id, ee)
		return
	} else {
		p.Permalink = id
		p.PermalinkSha = sum
		p.Context.SetSpecific("LongLink", "/"+p.Permalink.String())
		p.Context.SetSpecific("ShortLink", "/"+p.PermalinkSha)
		p.Context.SetSpecific("Permalink", id)
		p.Context.SetSpecific("Permalinked", true)
		p.Context.SetSpecific("PermalinkSha", sum)
		if p.Url == "/" {
			p.Context.SetSpecific("PermalinkUrl", p.Url+p.PermalinkSha)
		} else {
			p.Context.SetSpecific("PermalinkUrl", p.Url+"-"+p.PermalinkSha)
		}
	}
	return
}

func (p *Page) getUrlPathSectionSlug(url string) (path, section, slug string) {
	var notPath bool
	if notPath = strings.HasPrefix(url, "!"); notPath {
		url = url[1:]
	}
	path = bePath.TrimSlashes(url)
	path = strings.ToLower(path)
	slug = strcase.ToKebab(filepath.Base(path))
	if parts := strings.Split(path, "/"); len(parts) > 0 {
		section = parts[0]
	}
	if notPath {
		path = "!" + path
	} else {
		path = "/" + path
	}
	path = strings.ReplaceAll(path, "//", "/")
	return
}