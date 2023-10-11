//go:build !exclude_pages_formats && !exclude_pages_format_org

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

package org

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/blevesearch/bleve/v2/mapping"
	"golang.org/x/net/html"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

var Tag feature.Tag = "pages-formats-org"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type Feature interface {
	feature.Feature
	feature.PageFormat
}

type MakeFeature interface {
	// SetDefault updates org.Configuration.DefaultSettings
	SetDefault(key, value string) MakeFeature

	// SetDefaults replaces org.Configuration.DefaultSettings
	SetDefaults(defaults map[string]string) MakeFeature

	Make() Feature
}

type CFeature struct {
	feature.CFeature

	settings map[string]string
	replaced map[string]string
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.settings = make(map[string]string)
}

func (f *CFeature) SetDefault(key, value string) MakeFeature {
	f.settings[key] = value
	return f
}

func (f *CFeature) SetDefaults(settings map[string]string) MakeFeature {
	f.replaced = settings
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Setup(enjin feature.Internals) {
	f.CFeature.Setup(enjin)
}

func (f *CFeature) Name() (name string) {
	name = "org"
	return
}

func (f *CFeature) Extensions() (extensions []string) {
	extensions = append(extensions, "org", "org.tmpl")
	return
}

func (f *CFeature) Label() (label string) {
	label = "Org-Mode"
	return
}

func (f *CFeature) Prepare(ctx context.Context, content string) (out context.Context, err error) {
	return
}

func (f *CFeature) Process(ctx context.Context, content string) (html template.HTML, redirect string, err error) {
	if text, e := f.RenderOrgMode(content); e != nil {
		err = e
		return
	} else {
		html = template.HTML(f.Enjin.TranslateShortcodes(text, ctx))
	}
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
	r = lang.SetTag(r, pg.LanguageTag())
	for _, ptp := range feature.FilterTyped[feature.PageTypeProcessor](f.Enjin.Features().List()) {
		if v, _, processed, e := ptp.ProcessRequestPageType(r, pg); e != nil {
			log.ErrorF("error processing page type for org format indexing: %v - %v", pg.Url(), e)
		} else if processed {
			pg = v
		}
	}

	var rendered string
	if strings.HasSuffix(pg.Format(), ".tmpl") {
		renderer := f.Enjin.GetThemeRenderer(pg.Context())
		if rendered, err = renderer.RenderTextTemplateContent(f.Enjin.MustGetTheme(), pg.Context(), pg.Content()); err != nil {
			err = fmt.Errorf("error rendering .org.tmpl content: %v", err)
			return
		}
	} else {
		rendered = pg.Content()
	}

	if rendered, err = f.RenderOrgMode(rendered); err != nil {
		return
	}

	doc := NewOrgModeDocument(pg.Language(), pg.Url(), pg.Title())
	var parsed *html.Node
	if parsed, err = html.Parse(strings.NewReader(rendered)); err != nil {
		return
	}

	skipNext := false
	addLinkNext := false
	addHeadingNext := false

	contents := ""

	var slurp func(node *html.Node) (slurped string)
	slurp = func(node *html.Node) (slurped string) {
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.TextNode {
				slurped = beStrings.AppendWithSpace(slurped, c.Data)
			} else {
				slurped = beStrings.AppendWithSpace(slurped, slurp(c))
			}
		}
		return
	}

	var walk func(node *html.Node)
	walk = func(node *html.Node) {

		if node.Type == html.ElementNode {
			switch node.Data {
			case "a":
				addLinkNext = true
				for _, attr := range node.Attr {
					if attr.Key == "href" {
						if strings.HasPrefix(attr.Val, "#footnote-reference") {
							addLinkNext = false
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
						case "footnote-body":
							footnote := slurp(node)
							if !beStrings.Empty(footnote) {
								// log.DebugF("adding org-mode footnote: %v", footnote)
								doc.AddFootnote(footnote)
							}
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
						// log.DebugF("adding org-mode link: %v", data)
						doc.AddLink(data)
						contents = beStrings.AppendWithSpace(contents, data)
					} else if addHeadingNext {
						addHeadingNext = false
						// log.DebugF("adding org-mode heading: %v", data)
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
		// log.DebugF("adding org-mode contents:\n%v", contents)
	}

	out = doc
	err = nil
	return
}