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

	"github.com/go-enjin/be/features/pages/formats/njn/blocks/content"
	"github.com/go-enjin/be/features/pages/formats/njn/blocks/header"
	"github.com/go-enjin/be/features/pages/formats/njn/blocks/icon"
	"github.com/go-enjin/be/features/pages/formats/njn/blocks/image"
	"github.com/go-enjin/be/features/pages/formats/njn/blocks/linkList"
	"github.com/go-enjin/be/features/pages/formats/njn/blocks/notice"
	"github.com/go-enjin/be/features/pages/formats/njn/blocks/toc"
	"github.com/go-enjin/be/features/pages/formats/njn/fields/anchor"
	"github.com/go-enjin/be/features/pages/formats/njn/fields/code"
	"github.com/go-enjin/be/features/pages/formats/njn/fields/container"
	"github.com/go-enjin/be/features/pages/formats/njn/fields/details"
	"github.com/go-enjin/be/features/pages/formats/njn/fields/fa"
	"github.com/go-enjin/be/features/pages/formats/njn/fields/fieldset"
	"github.com/go-enjin/be/features/pages/formats/njn/fields/figure"
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
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/theme/types"
)

var (
	_ Feature      = (*CFeature)(nil)
	_ MakeFeature  = (*CFeature)(nil)
	_ types.Format = (*CFeature)(nil)
)

var _instance *CFeature

type Feature interface {
	feature.Feature
	types.Format
}

type MakeFeature interface {
	AddInlineField(field feature.EnjinField) MakeFeature
	AddContainerField(field feature.EnjinField) MakeFeature
	AddInlineBlock(field feature.EnjinBlock) MakeFeature
	AddContainerBlock(field feature.EnjinBlock) MakeFeature

	Defaults() MakeFeature

	Make() Feature
}

type CFeature struct {
	feature.CFeature

	inlineFields    map[string]feature.EnjinField
	containerFields map[string]feature.EnjinField
	inlineBlocks    map[string]feature.EnjinBlock
	containerBlocks map[string]feature.EnjinBlock
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

	f.inlineFields = make(map[string]feature.EnjinField)
	f.containerFields = make(map[string]feature.EnjinField)
	f.inlineBlocks = make(map[string]feature.EnjinBlock)
	f.containerBlocks = make(map[string]feature.EnjinBlock)
}

func (f *CFeature) Defaults() MakeFeature {
	// all inline fields
	f.AddInlineField(anchor.New().Make())
	f.AddInlineField(fa.New().Make())
	f.AddInlineField(figure.New().Make())
	f.AddInlineField(img.New().Make())
	f.AddInlineField(inline.New().Defaults().Make())
	f.AddInlineField(input.New().Make())
	f.AddInlineField(literal.New().Defaults().Make())
	f.AddInlineField(optgroup.New().Make())
	f.AddInlineField(option.New().Make())
	f.AddInlineField(picture.New().Make())
	f.AddInlineField(_select.New().Make())
	// all container fields
	f.AddContainerField(details.New().Make())
	f.AddContainerField(p.New().Make())
	f.AddContainerField(table.New().Make())
	f.AddContainerField(pre.New().Make())
	f.AddContainerField(literal.New().AddTag("hr").Make())
	f.AddContainerField(code.New().Make())
	f.AddContainerField(container.New().Defaults().Make())
	f.AddContainerField(list.New().Defaults().Make())
	f.AddContainerField(fieldset.New().Make())
	// all inline blocks
	f.AddInlineBlock(header.New().Make())
	f.AddInlineBlock(notice.New().Make())
	f.AddInlineBlock(linkList.New().Make())
	f.AddInlineBlock(toc.New().Make())
	f.AddInlineBlock(image.New().Make())
	f.AddInlineBlock(icon.New().Make())
	f.AddInlineBlock(content.New().Make())
	return f
}

func (f *CFeature) AddInlineField(field feature.EnjinField) MakeFeature {
	for _, name := range field.NjnFieldNames() {
		f.inlineFields[name] = field
		log.DebugF("added inline field: %v", name)
	}
	return f
}

func (f *CFeature) AddContainerField(field feature.EnjinField) MakeFeature {
	for _, name := range field.NjnFieldNames() {
		f.containerFields[name] = field
		log.DebugF("added container field: %v", name)
	}
	return f
}

func (f *CFeature) AddInlineBlock(block feature.EnjinBlock) MakeFeature {
	name := block.NjnBlockType()
	f.inlineBlocks[name] = block
	log.DebugF("added inline block: %v", name)
	return f
}

func (f *CFeature) AddContainerBlock(block feature.EnjinBlock) MakeFeature {
	name := block.NjnBlockType()
	f.containerBlocks[name] = block
	log.DebugF("added container block: %v", name)
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

func (f *CFeature) InlineFields() (field map[string]feature.EnjinField) {
	return f.inlineFields
}

func (f *CFeature) ContainerFields() (field map[string]feature.EnjinField) {
	return f.containerFields
}

func (f *CFeature) InlineBlocks() (field map[string]feature.EnjinBlock) {
	return f.inlineBlocks
}

func (f *CFeature) ContainerBlocks() (field map[string]feature.EnjinBlock) {
	return f.containerBlocks
}

func (f *CFeature) Process(ctx context.Context, t types.Theme, content string) (html template.HTML, err error) {
	var data interface{}
	if e := json.Unmarshal([]byte(content), &data); e != nil {
		switch t := e.(type) {
		case *json.SyntaxError:
			err = fmt.Errorf("json syntax error")
			html = template.HTML(fmt.Sprintf(`<p>json syntax error: <a style="color:red;" href="#json-error">[%d] %v</a></p>`, t.Offset, t.Error()))
			html += "\n<pre>\n"
			html += template.HTML(template.HTMLEscapeString(content[:t.Offset]))
			html += template.HTML(fmt.Sprintf(`<span style="color:red;weight:bold;" id="json-error">&lt;-- %v</span>`, t.Error()))
			html += template.HTML(template.HTMLEscapeString(content[t.Offset:]))
			html += "\n</pre>\n"
		case *json.UnmarshalTypeError:
			err = fmt.Errorf("json unmarshal error")
			html = template.HTML(fmt.Sprintf(`<p>json unmarshal error: <a style="color:red;" href="#json-error">[%d] %v</a></p>`, t.Offset, t.Error()))
			html += "\n<pre>\n"
			html += template.HTML(template.HTMLEscapeString(content[:t.Offset]))
			html += template.HTML(fmt.Sprintf(`<span style="color:red;weight:bold;" id="json-error">&lt;-- %v</span>`, t.Error()))
			html += template.HTML(template.HTMLEscapeString(content[t.Offset:]))
			html += "\n</pre>\n"
		default:
			err = fmt.Errorf("json decoding error")
			html = template.HTML(fmt.Sprintf(`<p>json decoding error: %v</p>`, t.Error()))
			html += "\n<pre>\n"
			html += template.HTML(template.HTMLEscapeString(content))
			html += "\n</pre>\n"
		}
		return
	}
	html, err = renderNjnData(f, ctx, t, data)
	return
}