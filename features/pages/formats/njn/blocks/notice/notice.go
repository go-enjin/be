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
	f := new(CBlock)
	f.Init(f)
	return f
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

func (f *CBlock) PrepareBlock(re feature.EnjinRenderer, blockType string, data map[string]interface{}) (block map[string]interface{}, redirect string, err error) {
	if blockType != "notice" {
		err = fmt.Errorf("%v does not implement %v block type", f.Tag(), blockType)
		return
	}

	var blockContent map[string]interface{}
	if blockContent, err = re.PrepareGenericBlockData(data["content"]); err != nil {
		return
	}

	block = re.PrepareGenericBlock("notice", data)

	if v, ok := data["dismiss"].(string); ok && beStrings.IsTrue(v) {
		block["Dismiss"] = "true"
	} else {
		block["Dismiss"] = "false"
	}

	if v, ok := data["open"].(string); ok && beStrings.IsTrue(v) {
		block["Open"] = "true"
	} else {
		block["Open"] = "false"
	}

	if v, ok := data["notice-type"].(string); ok {
		v = strings.ToLower(v)
		switch v {
		case "info", "warn", "error":
			block["NoticeType"] = v
		default:
			err = fmt.Errorf("invalid notice type: %v", v)
			return
		}
	} else {
		block["NoticeType"] = "info"
	}

	var sectionFields []interface{}
	if sections, ok := blockContent["section"].([]interface{}); ok {
		if prepared, e := re.PrepareContainerFieldList(sections); e != nil {
			err = fmt.Errorf("error preparing notice content section: %v", e)
			return
		} else {
			block["Section"] = prepared
			sectionFields = prepared
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
		block["Summary"] = summary
	} else if len(sectionFields) > 0 {
		block["Summary"] = sectionFields
		delete(block, "Section")
	} else {
		err = fmt.Errorf("notice block missing summary and section: %v - %v", data, sectionFields)
		return
	}

	if footer, ok := re.PrepareBlockFooter(blockContent); ok {
		block["Footer"] = footer
	}

	block["SiteContext"] = re.RequestContext()
	return
}

func (f *CBlock) RenderPreparedBlock(re feature.EnjinRenderer, block map[string]interface{}) (html template.HTML, err error) {
	html, err = re.RenderNjnTemplate("block/notice", block)
	return
}

func (f *CBlock) ProcessBlock(re feature.EnjinRenderer, blockType string, data map[string]interface{}) (html template.HTML, redirect string, err error) {
	if block, redir, e := f.PrepareBlock(re, blockType, data); e != nil {
		err = e
		return
	} else if redir != "" {
		redirect = redir
		return
	} else {
		html, err = f.RenderPreparedBlock(re, block)
	}
	return
}