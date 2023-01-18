//go:build database || all

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
	"fmt"

	"github.com/iancoleman/strcase"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/database"
	bePath "github.com/go-enjin/be/pkg/path"
)

type Table struct {
	Name string
	Path string
}

func NewTable(path, name string) *Table {
	return &Table{
		Name: strcase.ToSnake(name),
		Path: "/" + bePath.TrimSlashes(path),
	}
}

func (t *Table) Tx() (tx *gorm.DB) {
	tx = database.MustGet().Scopes(func(tx *gorm.DB) *gorm.DB {
		return tx.Table(t.Name)
	})
	return
}

func (t *Table) Migrate() (err error) {
	err = t.Tx().AutoMigrate(&Page{})
	return
}

func (t *Table) stripTablePathPrefix(path string) (cleaned string) {
	cleaned = bePath.TrimPrefix(path, t.Path)
	cleaned = "/" + bePath.TrimSlashes(cleaned)
	return
}

func (t *Table) Get(path string) (p *Page, err error) {
	cleaned := t.stripTablePathPrefix(path)
	sql := fmt.Sprintf("SELECT * FROM %v WHERE url = ?", t.Name)
	if err = t.Tx().Raw(sql, cleaned).First(&p).Error; err != nil {
		return
	}

	p.Context = context.New()
	p.Context.Set("Url", p.Url)
	p.Context.Set("Slug", p.Slug)
	p.Context.Set("Title", p.Title)
	p.Context.Set("Format", p.Format)
	p.Context.Set("Summary", p.Summary)
	p.Context.Set("Description", p.Description)
	p.Context.Set("Archetype", p.Archetype)
	p.Context.Set("Section", p.Section)
	p.Context.Set("Content", p.Content)
	p.Context.Set("Language", p.Language)

	frontMatter, _, frontMatterType := ParseFrontMatterContent(p.FrontMatter)
	switch frontMatterType {
	case TomlMatter:
		if ctx, ee := ParseToml(frontMatter); ee != nil {
			err = fmt.Errorf("error parsing page toml front matter: %v", ee)
		} else {
			p.parseContext(ctx)
		}
	case YamlMatter:
		if ctx, ee := ParseYaml(frontMatter); ee != nil {
			err = fmt.Errorf("error parsing page yaml front matter: %v", ee)
		} else {
			p.parseContext(ctx)
		}
	case JsonMatter:
		if ctx, ee := ParseJson(frontMatter); ee != nil {
			err = fmt.Errorf("error parsing page json front matter: %v", ee)
		} else {
			p.parseContext(ctx)
		}
	default:
		p.parseContext(p.Context)
	}

	// log.DebugF("%v page.Table got: %v", t.Name, p)
	return
}

func (t *Table) Put(p *Page) (err error) {
	p.Url = t.stripTablePathPrefix(p.Url)
	tx := t.Tx().Clauses(clause.OnConflict{UpdateAll: true})
	if err = tx.Create(p).Error; err != nil {
		return
	}
	// log.DebugF("%v page.Table put: %v", t.Name, p)
	return
}