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

package html

import (
	"bytes"
	"fmt"
	htmlTemplate "html/template"
	"strings"

	"github.com/blevesearch/bleve/v2/mapping"
	"golang.org/x/net/html"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/page"
	beStrings "github.com/go-enjin/be/pkg/strings"
	"github.com/go-enjin/be/pkg/theme"
	"github.com/go-enjin/be/pkg/types/theme-types"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "pages-formats-html"

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
	f.FeatureTag = Tag
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Name() (name string) {
	name = "html"
	return
}

func (f *CFeature) Extensions() (extensions []string) {
	extensions = append(extensions, "html", "html.tmpl")
	return
}

func (f *CFeature) Label() (label string) {
	label = "HTML"
	return
}

func (f *CFeature) Prepare(ctx context.Context, content string) (out context.Context, err error) {
	return
}

func (f *CFeature) Process(ctx context.Context, t types.Theme, content string) (html htmlTemplate.HTML, redirect string, err *types.EnjinError) {
	html = htmlTemplate.HTML(content)
	return
}

func (f *CFeature) SearchDocumentMapping(tag language.Tag) (doctype string, dm *mapping.DocumentMapping) {
	doctype, _, dm = f.NewDocumentMapping(tag)
	return
}

func (f *CFeature) AddSearchDocumentMapping(tag language.Tag, indexMapping *mapping.IndexMappingImpl) {
	doctype, _, dm := f.NewDocumentMapping(tag)
	indexMapping.AddDocumentMapping(doctype, dm)
}

func (f *CFeature) IndexDocument(thing interface{}) (out interface{}, err error) {
	pg, _ := thing.(*page.Page) // FIXME: this "thing" avoids package import loops

	var rendered string

	if strings.HasSuffix(pg.Format, ".tmpl") {
		var buf bytes.Buffer
		if tt, e := htmlTemplate.New("content.html.tmpl").Funcs(theme.DefaultFuncMap()).Parse(pg.Content); e != nil {
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

	doc := NewHtmlDocument(pg.Language, pg.Url, pg.Title)
	var parsed *html.Node
	if parsed, err = html.Parse(strings.NewReader(rendered)); err != nil {
		return
	}

	skipNext := false
	addLinkNext := false
	addHeadingNext := false

	contents := ""

	var walk func(node *html.Node)
	walk = func(node *html.Node) {

		if node.Type == html.ElementNode {
			switch node.Data {
			case "a":
				addLinkNext = true
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
			case "style":
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
						// log.DebugF("adding html link: %v", data)
						doc.AddLink(data)
						contents = beStrings.AppendWithSpace(contents, data)
					} else if addHeadingNext {
						addHeadingNext = false
						// log.DebugF("adding html heading: %v", data)
						doc.AddHeading(data)
					} else {
						contents = beStrings.AppendWithSpace(contents, data)
					}
				} else {
					addLinkNext = false
					addHeadingNext = false
				}
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(parsed)

	if !beStrings.Empty(contents) {
		doc.AddContent(contents)
		// log.DebugF("adding html contents:\n%v", contents)
	}

	out = doc
	err = nil
	return
}