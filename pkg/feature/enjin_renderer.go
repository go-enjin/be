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

import (
	"html/template"

	"golang.org/x/net/html"
)

type EnjinRenderer interface {
	RenderNjnTemplate(tag string, data map[string]interface{}) (html template.HTML, err error)

	ProcessBlock(data map[string]interface{}) (html template.HTML, err error)

	PrepareBlock(data map[string]interface{}) (block map[string]interface{}, err error)
	RenderPreparedBlock(block map[string]interface{}) (html template.HTML, err error)

	PrepareGenericBlockData(contentData interface{}) (blockDataContent map[string]interface{}, err error)
	PrepareGenericBlock(typeName string, data map[string]interface{}) (preparedData map[string]interface{})

	GetData() (data interface{})
	GetBlockIndex() (index int)

	GetWithinAside() (within bool)
	SetWithinAside(within bool)

	GetCurrentDepth() (depth int)
	IncCurrentDepth() (depth int)
	DecCurrentDepth() (depth int)

	GetHeadingCount() (count int)
	SetHeadingCount(count int)
	IncHeadingCount()

	GetHeadingLevel() (level int)
	SetHeadingLevel(level int)
	IncHeadingLevel()
	DecHeadingLevel()

	AddFootnote(blockIndex int, field map[string]interface{}) (index int)
	PrepareFootnotes(blockIndex int) (footnotes []map[string]interface{}, err error)

	ParseTypeName(data map[string]interface{}) (name string, ok bool)
	ParseFieldAndTypeName(data interface{}) (field map[string]interface{}, name string, ok bool)

	PrepareStringTags(text string) (data []interface{}, err error)
	WalkStringTags(doc *html.Node) (prepared []interface{})

	PrepareBlockHeader(content map[string]interface{}) (combined []interface{}, ok bool)
	PrepareBlockFooter(content map[string]interface{}) (combined []interface{}, ok bool)

	ParseBlockHeadingLevel(count, current int, blockData map[string]interface{}) (level, headingReset, headingLevel int)
	RenderBlockHeader(content map[string]interface{}) (html template.HTML, ok bool)
	RenderBlockFooter(content map[string]interface{}) (html template.HTML, ok bool)

	PrepareInlineFieldText(field map[string]interface{}) (combined []interface{}, err error)
	PrepareInlineFieldList(list []interface{}) (combined []interface{}, err error)
	PrepareInlineFields(fields []interface{}) (combined []interface{}, err error)
	PrepareInlineField(field map[string]interface{}) (prepared map[string]interface{}, err error)

	PrepareContainerFieldText(field map[string]interface{}) (fields []interface{}, err error)
	PrepareContainerFieldList(list []interface{}) (fields []interface{}, err error)
	PrepareContainerFields(fields []interface{}) (combined []map[string]interface{}, err error)
	PrepareContainerField(field map[string]interface{}) (prepared map[string]interface{}, err error)

	RenderInlineField(field map[string]interface{}) (combined []template.HTML, err error)
	RenderInlineFields(fields []interface{}) (combined []template.HTML, err error)
	RenderInlineFieldList(list []interface{}) (html template.HTML, err error)
	RenderInlineFieldText(field map[string]interface{}) (text template.HTML, err error)

	RenderContainerField(field map[string]interface{}) (combined []template.HTML, err error)
	RenderContainerFields(fields []interface{}) (combined []template.HTML, err error)
	RenderContainerFieldList(list []interface{}) (html template.HTML, err error)
	RenderContainerFieldText(field map[string]interface{}) (text template.HTML, err error)
}