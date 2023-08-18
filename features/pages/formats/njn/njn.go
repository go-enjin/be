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

package njn

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/blevesearch/bleve/v2/mapping"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/features/pages/formats/njn/blocks/card"
	"github.com/go-enjin/be/features/pages/formats/njn/blocks/carousel"
	"github.com/go-enjin/be/features/pages/formats/njn/blocks/content"
	"github.com/go-enjin/be/features/pages/formats/njn/blocks/header"
	"github.com/go-enjin/be/features/pages/formats/njn/blocks/icon"
	"github.com/go-enjin/be/features/pages/formats/njn/blocks/image"
	"github.com/go-enjin/be/features/pages/formats/njn/blocks/index"
	"github.com/go-enjin/be/features/pages/formats/njn/blocks/linkList"
	"github.com/go-enjin/be/features/pages/formats/njn/blocks/notice"
	"github.com/go-enjin/be/features/pages/formats/njn/blocks/pair"
	"github.com/go-enjin/be/features/pages/formats/njn/blocks/sidebar"
	"github.com/go-enjin/be/features/pages/formats/njn/blocks/toc"
	"github.com/go-enjin/be/features/pages/formats/njn/fields/anchor"
	"github.com/go-enjin/be/features/pages/formats/njn/fields/code"
	"github.com/go-enjin/be/features/pages/formats/njn/fields/container"
	"github.com/go-enjin/be/features/pages/formats/njn/fields/details"
	"github.com/go-enjin/be/features/pages/formats/njn/fields/fa"
	"github.com/go-enjin/be/features/pages/formats/njn/fields/fieldset"
	"github.com/go-enjin/be/features/pages/formats/njn/fields/figure"
	"github.com/go-enjin/be/features/pages/formats/njn/fields/footnote"
	"github.com/go-enjin/be/features/pages/formats/njn/fields/img"
	"github.com/go-enjin/be/features/pages/formats/njn/fields/inline"
	"github.com/go-enjin/be/features/pages/formats/njn/fields/input"
	"github.com/go-enjin/be/features/pages/formats/njn/fields/list"
	"github.com/go-enjin/be/features/pages/formats/njn/fields/literal"
	"github.com/go-enjin/be/features/pages/formats/njn/fields/optgroup"
	"github.com/go-enjin/be/features/pages/formats/njn/fields/option"
	"github.com/go-enjin/be/features/pages/formats/njn/fields/p"
	"github.com/go-enjin/be/features/pages/formats/njn/fields/picture"
	"github.com/go-enjin/be/features/pages/formats/njn/fields/pre"
	_select "github.com/go-enjin/be/features/pages/formats/njn/fields/select"
	"github.com/go-enjin/be/features/pages/formats/njn/fields/table"
	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/feature"
	beForms "github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/slices"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

var (
	DefaultStringTags = []string{
		"b", "del", "em", "i", "ins", "kbd", "mark",
		"q", "s", "small", "strong", "sub", "sup", "u",
		"var", "code",
	}
)

const Tag feature.Tag = "pages-formats-njn"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type Feature interface {
	feature.Feature
	feature.PageFormat
	feature.EnjinSystem
}

type MakeFeature interface {
	AddField(field feature.EnjinField) MakeFeature
	AddBlock(block feature.EnjinBlock) MakeFeature
	AddStringTags(names ...string) MakeFeature

	Defaults() MakeFeature

	Make() Feature
}

type CFeature struct {
	feature.CFeature

	fields map[feature.NjnClass]map[string]feature.EnjinField
	blocks map[feature.NjnClass]map[string]feature.EnjinBlock

	stringtags []string
}

func New() MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.FeatureTag = Tag
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)

	f.fields = make(map[feature.NjnClass]map[string]feature.EnjinField)
	f.blocks = make(map[feature.NjnClass]map[string]feature.EnjinBlock)
	f.stringtags = make([]string, 0)
}

func (f *CFeature) Setup(enjin feature.Internals) {
	f.CFeature.Setup(enjin)
	for _, blocks := range f.blocks {
		for _, block := range blocks {
			block.Setup(enjin)
		}
	}
}

func (f *CFeature) AddField(field feature.EnjinField) MakeFeature {

	add := func(tagClass feature.NjnClass) {
		if _, ok := f.fields[tagClass]; !ok {
			f.fields[tagClass] = make(map[string]feature.EnjinField)
		}
		for _, name := range field.NjnFieldNames() {
			f.fields[tagClass][name] = field
			log.TraceF("added %v field: %v", tagClass.String(), name)
		}
	}

	switch field.NjnClass() {
	case feature.AnyNjnClass:
		add(feature.InlineNjnClass)
		add(feature.ContainerNjnClass)
	case feature.ContainerNjnClass:
		add(feature.ContainerNjnClass)
	case feature.InlineNjnClass:
		add(feature.InlineNjnClass)
	default:
		log.FatalDF(1, "unsupported feature.NjnClass: %v", field.NjnClass())
	}

	return f
}

func (f *CFeature) AddBlock(block feature.EnjinBlock) MakeFeature {

	add := func(tagClass feature.NjnClass) {
		if _, ok := f.blocks[tagClass]; !ok {
			f.blocks[tagClass] = make(map[string]feature.EnjinBlock)
		}
		name := block.NjnBlockType()
		f.blocks[tagClass][name] = block
		log.TraceF("added %v block: %v", tagClass.String(), name)
	}

	switch block.NjnClass() {
	case feature.AnyNjnClass:
		add(feature.InlineNjnClass)
		add(feature.ContainerNjnClass)
	case feature.ContainerNjnClass:
		add(feature.ContainerNjnClass)
	case feature.InlineNjnClass:
		add(feature.InlineNjnClass)
	default:
		log.FatalDF(1, "unsupported feature.NjnClass: %v", block.NjnClass())
	}

	return f
}

func (f *CFeature) AddStringTags(names ...string) MakeFeature {
	for _, name := range names {
		if !slices.Present(name, f.stringtags...) {
			f.stringtags = append(f.stringtags, name)
			log.TraceF("added %v shortcode", name)
		}
	}
	return f
}

func (f *CFeature) Defaults() MakeFeature {
	// all inline fields
	f.AddField(anchor.New().Make())
	f.AddField(fa.New().Make())
	f.AddField(figure.New().Make())
	f.AddField(img.New().Make())
	f.AddField(inline.New().Defaults().Make())
	f.AddField(input.New().Make())
	f.AddField(literal.New().Defaults().Make())
	f.AddField(optgroup.New().Make())
	f.AddField(option.New().Make())
	f.AddField(picture.New().Make())
	f.AddField(_select.New().Make())
	f.AddField(footnote.New().Make())
	// all container fields
	f.AddField(details.New().Make())
	f.AddField(p.New().Make())
	f.AddField(table.New().Make())
	f.AddField(pre.New().Make())
	f.AddField(literal.New().SetTagClass(feature.ContainerNjnClass).AddTag("hr").Make())
	f.AddField(code.New().Make())
	f.AddField(container.New().Defaults().Make())
	f.AddField(list.New().Defaults().Make())
	f.AddField(fieldset.New().Make())
	// all inline blocks
	f.AddBlock(header.New().Make())
	f.AddBlock(notice.New().Make())
	f.AddBlock(linkList.New().Make())
	f.AddBlock(toc.New().Make())
	f.AddBlock(image.New().Make())
	f.AddBlock(icon.New().Make())
	f.AddBlock(card.New().Make())
	f.AddBlock(content.New().Make())
	f.AddBlock(index.New().Make())
	// all container blocks
	f.AddBlock(carousel.New().Make())
	f.AddBlock(pair.New().Make())
	f.AddBlock(sidebar.New().Make())
	// stringtags (text-level tags such as `<u>` and `<i>`)
	f.AddStringTags(DefaultStringTags...)
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Name() (name string) {
	name = "njn"
	return
}

func (f *CFeature) Extensions() (extensions []string) {
	extensions = append(extensions, "njn", "njn.tmpl")
	return
}

func (f *CFeature) Label() (label string) {
	label = "Enjin"
	return
}

func (f *CFeature) InlineFields() (fields map[string]feature.EnjinField) {
	fields, _ = f.fields[feature.InlineNjnClass]
	return
}

func (f *CFeature) ContainerFields() (fields map[string]feature.EnjinField) {
	fields, _ = f.fields[feature.ContainerNjnClass]
	return
}

func (f *CFeature) InlineBlocks() (blocks map[string]feature.EnjinBlock) {
	blocks, _ = f.blocks[feature.InlineNjnClass]
	return
}

func (f *CFeature) ContainerBlocks() (blocks map[string]feature.EnjinBlock) {
	blocks, _ = f.blocks[feature.ContainerNjnClass]
	return
}

func (f *CFeature) StringTags() (names []string) {
	return f.stringtags
}

func (f *CFeature) FindField(tagClass feature.NjnClass, fieldType string) (field feature.EnjinField, ok bool) {
	switch tagClass {
	case feature.AnyNjnClass:
		if field, ok = f.fields[feature.ContainerNjnClass][fieldType]; !ok {
			field, ok = f.fields[feature.InlineNjnClass][fieldType]
		}
	case feature.InlineNjnClass, feature.ContainerNjnClass:
		field, ok = f.fields[tagClass][fieldType]
	}
	return
}

func (f *CFeature) FindBlock(tagClass feature.NjnClass, blockType string) (block feature.EnjinBlock, ok bool) {
	switch tagClass {
	case feature.AnyNjnClass:
		if block, ok = f.blocks[feature.ContainerNjnClass][blockType]; !ok {
			block, ok = f.blocks[feature.InlineNjnClass][blockType]
		}
	case feature.InlineNjnClass, feature.ContainerNjnClass:
		block, ok = f.blocks[tagClass][blockType]
	}
	return
}

func (f *CFeature) Prepare(ctx context.Context, content string) (out context.Context, err error) {
	return
}

func (f *CFeature) Process(ctx context.Context, content string) (html template.HTML, redirect string, err error) {
	var data interface{}
	if e := json.Unmarshal([]byte(content), &data); e != nil {
		switch errType := e.(type) {
		case *json.SyntaxError:
			output := template.HTMLEscapeString(content[:errType.Offset])
			output += fmt.Sprintf(`<span style="color:red;weight:bold;" id="json-error">&lt;-- %v</span>`, errType.Error())
			output += template.HTMLEscapeString(content[errType.Offset:])
			err = errors.NewEnjinError(
				"json syntax error",
				fmt.Sprintf(`<a style="color:red;" href="#json-error">[%d] %v</a>`, errType.Offset, errType.Error()),
				output,
			)
		case *json.UnmarshalTypeError:
			output := template.HTMLEscapeString(content[:errType.Offset])
			output += fmt.Sprintf(`<span style="color:red;weight:bold;" id="json-error">&lt;-- %v</span>`, errType.Error())
			output += template.HTMLEscapeString(content[errType.Offset:])
			err = errors.NewEnjinError(
				"json unmarshal error",
				fmt.Sprintf(`<a style="color:red;" href="#json-error">[%d] %v</a>`, errType.Offset, errType.Error()),
				output,
			)
		default:
			err = errors.NewEnjinError(
				"json decoding error",
				errType.Error(),
				content,
			)
		}
		return
	}
	html, redirect, err = renderNjnData(f, ctx, data)
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

	doc := NewEnjinDocument(pg.Language(), pg.Url(), pg.Title())

	r, _ := http.NewRequest("GET", pg.Url(), nil)
	r = lang.SetTag(r, pg.LanguageTag())
	for _, ptp := range feature.FilterTyped[feature.PageTypeProcessor](f.Enjin.Features().List()) {
		if v, _, processed, e := ptp.ProcessRequestPageType(r, pg); e != nil {
			log.ErrorF("error processing page type for njn format indexing: %v - %v", pg.Url(), e)
		} else if processed {
			pg = v
		}
	}

	var rendered string
	if strings.HasSuffix(pg.Format(), ".tmpl") {
		renderer := f.Enjin.GetThemeRenderer(pg.Context())
		if rendered, err = renderer.RenderTextTemplateContent(pg.Context(), pg.Content()); err != nil {
			err = fmt.Errorf("error rendering .njn.tmpl content: %v", err)
			return
		}
	} else {
		rendered = pg.Content()
	}

	var data []interface{}
	if err = json.Unmarshal([]byte(rendered), &data); err != nil {
		err = fmt.Errorf("error parsing content: %v", err)
		log.ErrorF("error parsing content (data):\n%v", rendered)
		return
	}

	var walker func(data []interface{}) (contents string, err error)
	var parser func(data map[string]interface{}) (contents string, err error)
	switcher := func(value interface{}) (contents string, err error) {
		if values, ok := value.([]interface{}); ok {
			contents, err = walker(values)
		} else if text, ok := value.(string); ok {
			contents = text
		} else if thing, ok := value.(map[string]interface{}); ok {
			contents, err = parser(thing)
		} else {
			err = fmt.Errorf("unsupported structure: %T %+v", value, value)
		}
		return
	}

	parser = func(data map[string]interface{}) (contents string, err error) {
		if dType, ok := data["type"]; ok {
			switch dType {

			case "footnote":
				// looking for text and note
				if textValue, ok := data["text"]; ok {
					var text, note string
					if text, err = switcher(textValue); err != nil {
						err = fmt.Errorf("error parsing footnote text: %v", err)
						return
					}
					if noteValue, ok := data["note"]; ok {
						if note, err = switcher(noteValue); err != nil {
							err = fmt.Errorf("error parsing footnote note: %v", err)
							return
						}
						doc.AddFootnote(text + ": " + note)
						contents = beStrings.AppendWithSpace(contents, text)
					}
				} else {
					err = fmt.Errorf("footnote text not found: %+v", data)
				}

			case "a":
				// looking for text
				if textValue, ok := data["text"]; ok {
					var text string
					if text, err = switcher(textValue); err != nil {
						err = fmt.Errorf("error parsing anchor text: %v", err)
						return
					}
					doc.AddLink(text)
					contents = beStrings.AppendWithSpace(contents, text)
				} else {
					err = fmt.Errorf("anchor text not found: %+v", data)
				}

			default:

				if linkText, ok := data["link-text"].(string); ok {
					doc.AddLink(linkText)
				}

				if dataContent, ok := data["content"].(map[string]interface{}); ok {

					if headerData, ok := dataContent["header"]; ok {
						if headerList, ok := headerData.([]interface{}); ok {
							var text string
							if text, err = walker(headerList); err != nil {
								err = fmt.Errorf("error walking header list: %v", err)
							} else if text != "" {
								doc.AddHeading(text)
							}
						} else {
							err = fmt.Errorf("invalid header structure: %T %+v", headerData, headerData)
							return
						}
					}

					if sectionData, ok := dataContent["section"]; ok {
						if sectionList, ok := sectionData.([]interface{}); ok {
							var text string
							if text, err = walker(sectionList); err != nil {
								err = fmt.Errorf("error walking section list: %v", err)
							} else if text != "" {
								contents = beStrings.AppendWithSpace(contents, text)
							}
						} else {
							err = fmt.Errorf("invalid section structure: %T %+v", sectionData, sectionData)
							return
						}
					}

					if footerData, ok := dataContent["footer"]; ok {
						if footerList, ok := footerData.([]interface{}); ok {
							var text string
							if text, err = walker(footerList); err != nil {
								err = fmt.Errorf("error walking footer list: %v", err)
							} else if text != "" {
								contents = beStrings.AppendWithSpace(contents, text)
							}
						} else {
							err = fmt.Errorf("invalid footer structure: %T %+v", footerData, footerData)
							return
						}
					}

					if blocksData, ok := dataContent["blocks"]; ok {
						if blocksList, ok := blocksData.([]interface{}); ok {
							var text string
							if text, err = walker(blocksList); err != nil {
								err = fmt.Errorf("error walking blocks list: %v", err)
							} else if text != "" {
								contents = beStrings.AppendWithSpace(contents, text)
							}
						} else {
							err = fmt.Errorf("invalid blocks structure: %T %+v", blocksData, blocksData)
							return
						}
					}

				} else if dataContentValue, ok := data["content"]; ok {

					err = fmt.Errorf("invalid content structure: %T %+v", dataContentValue, dataContentValue)
					return

				} else if textValue, ok := data["text"]; ok {

					var text string
					if text, err = switcher(textValue); err != nil {
						err = fmt.Errorf("error parsing text: %v", err)
						return
					}

					contents = beStrings.AppendWithSpace(contents, text)
				}
			}
		}
		return
	}

	walker = func(data []interface{}) (contents string, err error) {
		for _, datum := range data {
			switch dt := datum.(type) {
			case string:
				contents = beStrings.AppendWithSpace(contents, dt)
			case map[string]interface{}:
				var text string
				if text, err = parser(dt); err != nil {
					return
				} else if text != "" {
					contents = beStrings.AppendWithSpace(contents, text)
				}
			default:
				err = fmt.Errorf("invalid data structure: %T %+v", dt, dt)
			}
		}
		return
	}

	var contents string
	if contents, err = walker(data); err != nil {
		return
	}
	doc.AddContent(beForms.StrictPolicy(beStrings.StripTmplTags(contents)))

	out = doc
	return
}