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

package md

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	textTemplate "text/template"

	"github.com/blevesearch/bleve/v2/mapping"
	"golang.org/x/net/html"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/page"
	"github.com/go-enjin/be/pkg/search"
	beStrings "github.com/go-enjin/be/pkg/strings"
	"github.com/go-enjin/be/pkg/theme"
	"github.com/go-enjin/be/pkg/types/theme-types"
)

var (
	_ Feature      = (*CFeature)(nil)
	_ MakeFeature  = (*CFeature)(nil)
	_ types.Format = (*CFeature)(nil)
)

type Feature interface {
	feature.Feature
	types.Format
}

type MakeFeature interface {
	Make() Feature
}

type CFeature struct {
	feature.CFeature
}

func New() MakeFeature {
	f := new(CFeature)
	f.Init(f)
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
}

func (f *CFeature) Tag() (tag feature.Tag) {
	tag = "PageFormatMarkdown"
	return
}

func (f *CFeature) Name() (name string) {
	name = "md"
	return
}

func (f *CFeature) Extensions() (extensions []string) {
	extensions = append(extensions, "md", "md.tmpl")
	return
}

func (f *CFeature) Label() (label string) {
	label = "Markdown"
	return
}

func (f *CFeature) Process(ctx context.Context, t types.Theme, content string) (html template.HTML, redirect string, err *types.EnjinError) {
	html = template.HTML(RenderMarkdown(content))
	return
}

func (f *CFeature) AddSearchDocumentMapping(tag language.Tag, indexMapping *mapping.IndexMappingImpl) {
	doctype, _, dm := f.NewDocumentMapping(tag)
	indexMapping.AddDocumentMapping(doctype, dm)
}

func (f *CFeature) IndexDocument(p interface{}) (doc search.Document, err error) {
	pg, _ := p.(*page.Page)

	var rendered string

	if strings.HasSuffix(pg.Format, ".tmpl") {
		var buf bytes.Buffer
		if tt, e := textTemplate.New("content.md.text").Funcs(theme.DefaultFuncMap()).Parse(pg.Content); e != nil {
			err = fmt.Errorf("error parsing template: %v", e)
			return
		} else if e = tt.Execute(&buf, pg.Context); e != nil {
			err = fmt.Errorf("error executing template: %v", e)
			return
		} else {
			rendered = buf.String()
		}
	} else {
		rendered = pg.Content
	}

	rendered = RenderMarkdown(rendered)

	d := NewMarkdownDocument(pg.Language, pg.Url, pg.Title)
	var parsed *html.Node
	if parsed, err = html.Parse(strings.NewReader(rendered)); err != nil {
		return
	}

	skipNext := false
	addLinkNext := false
	addHeadingNext := false
	addFootnoteNext := false

	contents := ""

	var walk func(node *html.Node)
	walk = func(node *html.Node) {

		if node.Type == html.ElementNode {
			switch node.Data {
			case "a":
				addLinkNext = true
				for _, attr := range node.Attr {
					if attr.Key == "href" {
						if strings.HasPrefix(attr.Val, "#fn:") {
							addLinkNext = false
							break
						}
					}
				}
			case "li":
				for _, attr := range node.Attr {
					if attr.Key == "id" {
						if strings.HasPrefix(attr.Val, "fn:") {
							addFootnoteNext = true
							break
						}
					}
				}
			case "h1", "h2", "h3", "h4", "h5", "h6":
				addHeadingNext = true
			case "div":
				for _, attr := range node.Attr {
					if attr.Key == "class" {
						switch attr.Val {
						case "h1", "h2", "h3", "h4", "h5", "h6":
							addHeadingNext = true
							break
						}
					}
					if addHeadingNext {
						break
					}
				}
			case "style", "sup", "sub":
				skipNext = true
			}
		} else if node.Type == html.TextNode {
			if skipNext {
				skipNext = false
				// log.DebugF("skipping text: %v - %v", node.Type, node.Data)
			} else {
				data := beStrings.StripTmplTags(node.Data)
				data = strings.ReplaceAll(data, "permalink", "")
				data = strings.ReplaceAll(data, "top", "")
				if !beStrings.Empty(data) {
					if addLinkNext {
						addLinkNext = false
						// log.DebugF("adding markdown link: %v", data)
						d.AddLink(data)
						contents = beStrings.AppendWithSpace(contents, data)
					} else if addHeadingNext {
						addHeadingNext = false
						// log.DebugF("adding markdown heading: %v", data)
						d.AddHeading(data)
					} else if addFootnoteNext {
						addFootnoteNext = false
						// log.DebugF("adding markdown footnote: %v", data)
						d.AddFootnote(data)
					} else {
						contents = beStrings.AppendWithSpace(contents, data)
					}
				} else {
					addLinkNext = false
					addHeadingNext = false
					addFootnoteNext = false
				}
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(parsed)

	if !beStrings.Empty(contents) {
		d.AddContent(contents)
		// log.DebugF("adding markdown contents:\n%v", contents)
	}

	doc = d
	err = nil
	return
}