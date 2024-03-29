//go:build !exclude_pages_formats && !exclude_pages_format_bb

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

package bb

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/frustra/bbcode"
	"golang.org/x/net/html"

	clStrings "github.com/go-corelibs/strings"
	"github.com/go-corelibs/x-text/language"
	"github.com/go-corelibs/x-text/message"
	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
)

const Tag feature.Tag = "pages-formats-bb"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type Feature interface {
	feature.Feature
	feature.PageFormat
}

type MakeFeature interface {
	Make() Feature
}

type CFeature struct {
	feature.CFeature

	compiler bbcode.Compiler
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.CFeature.Construct(f)
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.compiler = bbcode.NewCompiler(false, false)
}

func (f *CFeature) AutoCloseTags(enabled bool) MakeFeature {
	f.compiler.AutoCloseTags = enabled
	return f
}

func (f *CFeature) IgnoreUnmatchedClosingTags(ignored bool) MakeFeature {
	f.compiler.IgnoreUnmatchedClosingTags = ignored
	return f
}

func (f *CFeature) SortOutputAttributes(sorted bool) MakeFeature {
	f.compiler.SortOutputAttributes = sorted
	return f
}

func (f *CFeature) AddTag(name string, fn bbcode.TagCompilerFunc) MakeFeature {
	f.compiler.SetTag(name, fn)
	return f
}

func (f *CFeature) RemoveTag(name string) MakeFeature {
	f.compiler.SetTag(name, nil)
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Name() (name string) {
	name = "bb"
	return
}

func (f *CFeature) Extensions() (extensions []string) {
	extensions = append(extensions, "bb", "bb.tmpl")
	return
}

func (f *CFeature) Label() (label string) {
	label = "Bulletin-Board"
	return
}

func (f *CFeature) Prepare(ctx context.Context, content string) (out context.Context, err error) {
	return
}

func (f *CFeature) Process(ctx context.Context, content string) (html template.HTML, redirect string, err error) {
	html = template.HTML(f.Enjin.TranslateShortcodes(f.compiler.Compile(content), ctx))
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

func (f *CFeature) IndexDocument(pg feature.Page) (out interface{}, err error) {

	r, _ := http.NewRequest("GET", pg.Url(), nil)
	r = message.SetTag(r, pg.LanguageTag())
	for _, ptp := range feature.FilterTyped[feature.PageTypeProcessor](f.Enjin.Features().List()) {
		if v, _, processed, e := ptp.ProcessRequestPageType(r, pg); e != nil {
			log.ErrorF("error processing page type for html format indexing: %v - %v", pg.Url(), e)
		} else if processed {
			pg = v
		}
	}

	var rendered string
	if strings.HasSuffix(pg.Format(), ".tmpl") {
		renderer := f.Enjin.GetThemeRenderer(pg.Context())
		if rendered, err = renderer.RenderHtmlTemplateContent(f.Enjin.MustGetTheme(), pg.Context(), pg.Content()); err != nil {
			err = fmt.Errorf("error rendering .html.tmpl content: %v", err)
			return
		}
	} else {
		rendered = pg.Content()
	}

	doc := NewBulletinBoardDocument(pg.Language(), pg.Url(), pg.Title())
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
				data := clStrings.PruneTmplActions(node.Data)
				data = strings.ReplaceAll(data, "permalink", "")
				data = strings.ReplaceAll(data, "top", "")
				if !clStrings.Empty(data) {
					if addLinkNext {
						addLinkNext = false
						// log.DebugF("adding html link: %v", data)
						doc.AddLink(data)
						contents = clStrings.AppendWithSpace(contents, data)
					} else if addHeadingNext {
						addHeadingNext = false
						// log.DebugF("adding html heading: %v", data)
						doc.AddHeading(data)
					} else {
						contents = clStrings.AppendWithSpace(contents, data)
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

	if !clStrings.Empty(contents) {
		doc.AddContent(contents)
		// log.DebugF("adding html contents:\n%v", contents)
	}

	out = doc
	err = nil
	return
}
