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

	"github.com/iancoleman/strcase"
	"gorm.io/gorm"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net"
	bePath "github.com/go-enjin/be/pkg/path"
	beStrings "github.com/go-enjin/be/pkg/strings"
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
	Language    string `json:"lang"`
	Content     string `json:"content"`

	Context context.Context `json:"context" gorm:"-"`
}

func New() *Page {
	p := new(Page)
	p.Context = context.New()
	return p
}

func newPageForPath(path string) (p *Page, err error) {
	p = New()
	path = bePath.TrimSlashes(path)
	if extn := bePath.Ext(path); extn != "" {
		name := strings.ToLower(extn)
		if format := GetFormat(name); format != nil {
			p.Format = name
		} else {
			p.Format = "<unsupported>"
		}
	}
	p.Slug = strcase.ToKebab(bePath.Base(path))
	if path == "/" {
		p.Url = "/"
	} else if len(strings.Split(path, "/")) >= 2 {
		p.Url = bePath.Dir(path) + "/" + p.Slug
	} else {
		p.Url = "/" + p.Slug
	}
	p.Title = beStrings.TitleCase(strings.Join(strings.Split(p.Slug, "-"), " "))
	p.Context.Set("Url", p.Url)
	p.Context.Set("Slug", p.Slug)
	p.Context.Set("Title", p.Title)
	return
}

func NewFromFile(path, filePath string) (p *Page, err error) {
	path = net.TrimQueryParams(path)
	if p, err = newPageForPath(path); err != nil {
		return
	}
	var data []byte
	if data, err = bePath.ReadFile(filePath); err != nil {
		return
	}
	raw := string(data)
	if !p.parseYaml(raw) {
		if !p.parseToml(raw) {
			if !p.parseJson(raw) {
				p.Content = raw
				p.parseContext(p.Context)
			}
		}
	}
	log.DebugF("new page from file: %v\n%v", filePath, p.Context)
	return
}

func NewFromString(path, raw string) (p *Page, err error) {
	path = net.TrimQueryParams(path)
	if p, err = newPageForPath(path); err != nil {
		return
	}
	if !p.parseYaml(raw) {
		if !p.parseToml(raw) {
			if !p.parseJson(raw) {
				p.Content = raw
				p.parseContext(p.Context)
			}
		}
	}
	return
}

func (p *Page) String() string {
	return p.Url
}