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

package feature

import "html/template"

type EnjinProvider interface {
	InlineFields() (field map[string]EnjinField)
	ContainerFields() (field map[string]EnjinField)
	InlineBlocks() (field map[string]EnjinBlock)
	ContainerBlocks() (field map[string]EnjinBlock)
}

type EnjinRenderer interface {
	RenderNjnTemplate(tag string, data map[string]interface{}) (html template.HTML, err error)

	ProcessBlock(blockData map[string]interface{}) (html template.HTML, err error)
	PrepareGenericBlockData(contentData interface{}) (blockDataContent map[string]interface{}, err error)
	PrepareGenericBlock(typeName string, data map[string]interface{}) (preparedData map[string]interface{})

	GetData() (data interface{})

	GetHeadingCount() (count int)
	SetHeadingCount(count int)
	IncHeadingCount()

	GetHeadingLevel() (level int)
	SetHeadingLevel(level int)
	IncHeadingLevel()
	DecHeadingLevel()

	ParseBlockHeadingLevel(count, current int, blockData map[string]interface{}) (level, headingReset, headingLevel int)
	ParseBlockHeader(content map[string]interface{}) (html template.HTML, ok bool)
	ParseBlockFooter(content map[string]interface{}) (html template.HTML, ok bool)

	RenderInlineField(field map[string]interface{}) (combined []template.HTML, err error)
	RenderInlineFields(fields []interface{}) (combined []template.HTML, err error)
	RenderInlineFieldList(list []interface{}) (html template.HTML, err error)
	RenderInlineFieldText(field map[string]interface{}) (text template.HTML, err error)

	RenderContainerField(field map[string]interface{}) (combined []template.HTML, err error)
	RenderContainerFields(fields []interface{}) (combined []template.HTML, err error)
	RenderContainerFieldList(list []interface{}) (html template.HTML, err error)
	RenderContainerFieldText(field map[string]interface{}) (text template.HTML, err error)
}

type EnjinField interface {
	Feature

	NjnFieldNames() (tagNames []string)
	PrepareNjnData(re EnjinRenderer, tagName string, field map[string]interface{}) (data map[string]interface{}, err error)
}

type CEnjinField struct {
	CFeature
}

type EnjinBlock interface {
	Feature

	NjnBlockType() (name string)
	ProcessBlock(re EnjinRenderer, blockType string, block map[string]interface{}) (html template.HTML, err error)
}

type CEnjinBlock struct {
	CFeature
}