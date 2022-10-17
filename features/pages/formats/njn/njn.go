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

	"github.com/blevesearch/bleve/v2/mapping"

	"github.com/go-enjin/be/features/pages/formats/njn/blocks/card"
	"github.com/go-enjin/be/features/pages/formats/njn/blocks/carousel"
	"github.com/go-enjin/be/features/pages/formats/njn/blocks/content"
	"github.com/go-enjin/be/features/pages/formats/njn/blocks/header"
	"github.com/go-enjin/be/features/pages/formats/njn/blocks/icon"
	"github.com/go-enjin/be/features/pages/formats/njn/blocks/image"
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
	"github.com/go-enjin/be/pkg/feature"
	beForms "github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/page"
	"github.com/go-enjin/be/pkg/search"
	beStrings "github.com/go-enjin/be/pkg/strings"
	"github.com/go-enjin/be/pkg/theme/types"
)

var (
	DefaultStringTags = []string{
		"b", "del", "em", "i", "ins", "kbd", "mark",
		"q", "s", "small", "strong", "sub", "sup", "u",
		"var", "code",
	}
)

var (
	_ Feature             = (*CFeature)(nil)
	_ MakeFeature         = (*CFeature)(nil)
	_ types.Format        = (*CFeature)(nil)
	_ feature.EnjinSystem = (*CFeature)(nil)
)

var _instance *CFeature

type Feature interface {
	feature.Feature
	types.Format
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
	if _instance == nil {
		_instance = new(CFeature)
		_instance.Init(_instance)
	}
	return _instance
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)

	f.fields = make(map[feature.NjnClass]map[string]feature.EnjinField)
	f.blocks = make(map[feature.NjnClass]map[string]feature.EnjinBlock)
	f.stringtags = make([]string, 0)
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
		if !beStrings.StringInStrings(name, f.stringtags...) {
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

func (f *CFeature) Tag() (tag feature.Tag) {
	tag = "PageFormatEnjin"
	return
}

func (f *CFeature) Name() (name string) {
	name = "njn"
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

func (f *CFeature) Process(ctx context.Context, t types.Theme, content string) (html template.HTML, err *types.EnjinError) {
	var data interface{}
	if e := json.Unmarshal([]byte(content), &data); e != nil {
		switch t := e.(type) {
		case *json.SyntaxError:
			output := template.HTMLEscapeString(content[:t.Offset])
			output += fmt.Sprintf(`<span style="color:red;weight:bold;" id="json-error">&lt;-- %v</span>`, t.Error())
			output += template.HTMLEscapeString(content[t.Offset:])
			err = types.NewEnjinError(
				"json syntax error",
				fmt.Sprintf(`<a style="color:red;" href="#json-error">[%d] %v</a>`, t.Offset, t.Error()),
				output,
			)
		case *json.UnmarshalTypeError:
			output := template.HTMLEscapeString(content[:t.Offset])
			output += fmt.Sprintf(`<span style="color:red;weight:bold;" id="json-error">&lt;-- %v</span>`, t.Error())
			output += template.HTMLEscapeString(content[t.Offset:])
			err = types.NewEnjinError(
				"json unmarshal error",
				fmt.Sprintf(`<a style="color:red;" href="#json-error">[%d] %v</a>`, t.Offset, t.Error()),
				output,
			)
		default:
			err = types.NewEnjinError(
				"json decoding error",
				t.Error(),
				content,
			)
		}
		return
	}
	html, err = renderNjnData(f, ctx, t, data)
	return
}

func (f *CFeature) AddSearchDocumentMapping(indexMapping *mapping.IndexMappingImpl) {
	indexMapping.AddDocumentMapping("njn", NewEnjinDocumentMapping())
}

func (f *CFeature) IndexDocument(ctx context.Context, content string) (doc search.Document, err error) {
	var url, title string
	if url = ctx.String("Url", ""); url == "" {
		err = fmt.Errorf("index document missing Url")
		return
	}
	if title = ctx.String("Title", ""); url == "" {
		err = fmt.Errorf("index document missing Title")
		return
	}

	d := NewEnjinDocument(url, title)
	// d.AddContent(content)

	var pg *page.Page
	if pg, err = page.NewFromString(url, content); err != nil {
		return
	}

	var data []interface{}
	if err = json.Unmarshal([]byte(pg.Content), &data); err != nil {
		return
	}

	var walker func(data []interface{}) (contents string)
	var parser func(data map[string]interface{}) (contents string)

	parser = func(data map[string]interface{}) (contents string) {
		if dType, ok := data["type"]; ok {
			switch dType {

			case "footnote":
				// looking for text and note
				if textValue, ok := data["text"]; ok {
					var text, note string
					if textList, ok := textValue.([]interface{}); ok {
						text = walker(textList)
					} else if textString, ok := textValue.(string); ok {
						text = textString
					} else if textThing, ok := textValue.(map[string]interface{}); ok {
						text = parser(textThing)
					} else {
						log.ErrorF("error parsing footnote text structure: %T %+v", textValue, textValue)
						return
					}
					if noteValue, ok := data["note"]; ok {
						if noteList, ok := noteValue.([]interface{}); ok {
							note = walker(noteList)
						} else if noteString, ok := noteValue.(string); ok {
							note = noteString
						} else if noteThing, ok := noteValue.(map[string]interface{}); ok {
							note = parser(noteThing)
						} else {
							log.ErrorF("error parsing footnote note structure: %T %+v", noteValue, noteValue)
							return
						}
						d.AddFootnote(text + ": " + note)
						contents = beStrings.AppendWithSpace(contents, text)
					}
				} else {
					log.ErrorF("error parsing footnote, missing text: %+v", data)
				}

			case "a":
				// looking for text
				if textValue, ok := data["text"]; ok {
					var text string
					if textList, ok := textValue.([]interface{}); ok {
						text = walker(textList)
					} else if textString, ok := textValue.(string); ok {
						text = textString
					} else if textThing, ok := textValue.(map[string]interface{}); ok {
						text = parser(textThing)
					} else {
						log.ErrorF("error parsing anchor text structure: %T %+v", textValue, textValue)
						return
					}
					d.AddLink(text)
					contents = beStrings.AppendWithSpace(contents, text)
				} else {
					log.ErrorF("error parsing anchor, missing text: %+v", data)
					return
				}

			default:

				if dataContent, ok := data["content"].(map[string]interface{}); ok {

					if dataHeader, ok := dataContent["header"]; ok {
						if headerList, ok := dataHeader.([]interface{}); ok {
							text := walker(headerList)
							if text != "" {
								d.AddHeading(text)
							}
						} else {
							log.ErrorF("error parsing header structure: %T %+v", dataHeader, dataHeader)
							return
						}
					}

					if dataSection, ok := dataContent["section"]; ok {
						if sectionList, ok := dataSection.([]interface{}); ok {
							section := walker(sectionList)
							if section != "" {
								contents = beStrings.AppendWithSpace(contents, section)
							}
						} else {
							log.ErrorF("error parsing section structure: %T %+v", dataSection, dataSection)
							return
						}
					}

					if dataFooter, ok := dataContent["footer"]; ok {
						if footerList, ok := dataFooter.([]interface{}); ok {
							footer := walker(footerList)
							if footer != "" {
								contents = beStrings.AppendWithSpace(contents, footer)
							}
						} else {
							log.ErrorF("error parsing footer structure: %T %+v", dataFooter, dataFooter)
							return
						}
					}

				} else if dataContentValue, ok := data["content"]; ok {
					log.ErrorF("error parsing data content structure: %T %+v", dataContentValue, dataContentValue)
					return

				} else if textValue, ok := data["text"]; ok {

					var text string
					if textList, ok := textValue.([]interface{}); ok {
						text = walker(textList)
					} else if textString, ok := textValue.(string); ok {
						text = textString
					} else if textThing, ok := textValue.(map[string]interface{}); ok {
						text = parser(textThing)
					} else {
						log.ErrorF("error parsing text structure: %T %+v", textValue, textValue)
						return
					}

					contents = beStrings.AppendWithSpace(contents, text)

				}

			}
		}
		return
	}

	walker = func(data []interface{}) (contents string) {
		for _, datum := range data {
			switch dt := datum.(type) {
			case string:
				contents += dt
			case map[string]interface{}:
				contents += parser(dt)
			default:
				log.ErrorF("unexpected data structure: %T %+v", dt, dt)
			}
		}
		return
	}

	contents := walker(data)
	d.AddContent(beForms.StripTags(beStrings.StripTmplTags(contents)))

	doc = d
	err = nil
	return
}