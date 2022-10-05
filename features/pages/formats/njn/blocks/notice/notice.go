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

package notice

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/go-enjin/be/pkg/feature"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

const (
	Tag feature.Tag = "NjnNoticeBlock"
)

var (
	_ Block     = (*CBlock)(nil)
	_ MakeBlock = (*CBlock)(nil)
)

var _instance *CBlock

type Block interface {
	feature.EnjinBlock
}

type MakeBlock interface {
	Make() Block
}

type CBlock struct {
	feature.CEnjinBlock
}

func New() (field MakeBlock) {
	if _instance == nil {
		_instance = new(CBlock)
		_instance.Init(_instance)
	}
	field = _instance
	return
}

func (f *CBlock) Tag() feature.Tag {
	return Tag
}

func (f *CBlock) Init(this interface{}) {
	f.CEnjinBlock.Init(this)
}

func (f *CBlock) Make() Block {
	return f
}

func (f *CBlock) NjnClass() (tagClass feature.NjnClass) {
	tagClass = feature.InlineNjnClass
	return
}

func (f *CBlock) NjnBlockType() (name string) {
	name = "notice"
	return
}

func (f *CBlock) ProcessBlock(re feature.EnjinRenderer, blockType string, block map[string]interface{}) (html template.HTML, err error) {
	if blockType != "notice" {
		err = fmt.Errorf("%v does not implement %v block type", f.Tag(), blockType)
		return
	}

	var blockContent map[string]interface{}
	if blockContent, err = re.PrepareGenericBlockData(block["content"]); err != nil {
		return
	}
	preparedData := re.PrepareGenericBlock("notice", block)

	if v, ok := block["dismiss"].(string); ok && beStrings.IsTrue(v) {
		preparedData["Dismiss"] = "true"
	} else {
		preparedData["Dismiss"] = "false"
	}

	if v, ok := block["open"].(string); ok && beStrings.IsTrue(v) {
		preparedData["Open"] = "true"
	} else {
		preparedData["Open"] = "false"
	}

	if v, ok := block["notice-type"].(string); ok {
		v = strings.ToLower(v)
		switch v {
		case "info", "warn", "error":
			preparedData["NoticeType"] = v
		default:
			err = fmt.Errorf("invalid notice type: %v", v)
			return
		}
	} else {
		preparedData["NoticeType"] = "info"
	}

	var sectionHtml template.HTML
	if sections, ok := blockContent["section"].([]interface{}); ok {
		var renderedSectionFields []template.HTML
		if renderedSectionFields, err = re.RenderContainerFields(sections); err != nil {
			err = fmt.Errorf("error rendering notice section field: %v", err)
			return
		} else {
			preparedData["Section"] = renderedSectionFields
			for _, s := range renderedSectionFields {
				sectionHtml += s
			}
		}
	}

	if v, ok := blockContent["summary"].([]interface{}); ok {
		var summary string
		for idx, vv := range v {
			if vs, ok := vv.(string); ok {
				if idx > 0 {
					summary += " "
				}
				summary += vs
			}
		}
		preparedData["Summary"] = summary
	} else if sectionHtml != "" {
		preparedData["Summary"] = sectionHtml
		delete(preparedData, "Section")
	} else {
		err = fmt.Errorf("notice block missing summary: %v - %v", block, sectionHtml)
		return
	}

	if footer, ok := re.ParseBlockFooter(blockContent); ok {
		preparedData["Footer"] = footer
	}

	html, err = re.RenderNjnTemplate("block/notice", preparedData)
	return
}