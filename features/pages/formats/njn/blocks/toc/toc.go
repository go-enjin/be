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

package toc

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/go-enjin/be/pkg/feature"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

const (
	Tag feature.Tag = "NjnTableOfContentsBlock"
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
	name = "toc"
	return
}

func (f *CBlock) PrepareBlock(re feature.EnjinRenderer, blockType string, data map[string]interface{}) (block map[string]interface{}, err error) {
	if blockType != "toc" {
		err = fmt.Errorf("%v does not implement %v block type", f.Tag(), blockType)
		return
	}

	var blockDataContent map[string]interface{}
	if blockDataContent, err = re.PrepareGenericBlockData(data["content"]); err != nil {
		if err.Error() != "content not found" {
			return
		}
		err = nil
		blockDataContent = make(map[string]interface{})
	}

	block = re.PrepareGenericBlock("toc", data)

	pageTitle := false
	if v, ok := data["page-title"]; ok {
		switch t := v.(type) {
		case string:
			pageTitle = beStrings.IsTrue(t)
		case bool:
			pageTitle = t
		case int:
			pageTitle = t > 0
		case float64:
			pageTitle = t > 0
		}
		if pageTitle {
			block["PageTitle"] = "true"
		} else {
			block["PageTitle"] = "false"
		}
	} else {
		block["PageTitle"] = "false"
	}

	var withSelf bool
	if v, ok := data["with-self"]; ok {
		switch t := v.(type) {
		case string:
			withSelf = beStrings.IsTrue(t)
		case bool:
			withSelf = t
		case int:
			withSelf = t > 0
		case float64:
			withSelf = t > 0
		}
		if withSelf {
			block["WithSelf"] = "true"
		} else {
			block["WithSelf"] = "false"
		}
	} else {
		block["WithSelf"] = "false"
	}

	_, _, toc := walkTableOfContents(re, 0, 0, re.GetData())
	items := sortTableOfContents(toc)

	blockTag, _ := block["Tag"].(string)
	var tmp []*tocItem
	for _, item := range items {
		if item.Tag != blockTag {
			tmp = append(tmp, item)
		}
	}
	items = tmp
	if heading, ok := re.PrepareBlockHeader(blockDataContent); ok {
		block["Heading"] = heading
		headingRendered, _ := re.RenderBlockHeader(blockDataContent)
		block["TocHeading"] = headingRendered
		if withSelf {
			items = append([]*tocItem{
				{
					Tag:   blockTag,
					Title: headingRendered,
				},
			}, items...)
		}
	}

	if footer, ok := re.PrepareBlockFooter(blockDataContent); ok {
		block["Footer"] = footer
	}

	block["Items"] = items

	if v, ok := data["counter"].(string); ok {
		v = strings.ToLower(v)
		switch v {
		case "nested", "single":
			block["Counter"] = v
		default:
			err = fmt.Errorf("invalid toc counter value: %v", v)
			return
		}
	} else {
		block["Counter"] = "single"
	}

	return
}

func (f *CBlock) RenderPreparedBlock(re feature.EnjinRenderer, block map[string]interface{}) (html template.HTML, err error) {
	html, err = re.RenderNjnTemplate("block/toc", block)
	return
}

func (f *CBlock) ProcessBlock(re feature.EnjinRenderer, blockType string, data map[string]interface{}) (html template.HTML, err error) {
	if block, e := f.PrepareBlock(re, blockType, data); e != nil {
		err = e
		return
	} else {
		html, err = f.RenderPreparedBlock(re, block)
	}
	return
}