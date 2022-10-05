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

func (f *CBlock) NjnTagClass() (tagClass feature.NjnTagClass) {
	tagClass = feature.InlineNjnTag
	return
}

func (f *CBlock) NjnBlockType() (name string) {
	name = "toc"
	return
}

func (f *CBlock) ProcessBlock(re feature.EnjinRenderer, blockType string, block map[string]interface{}) (html template.HTML, err error) {
	if blockType != "toc" {
		err = fmt.Errorf("%v does not implement %v block type", f.Tag(), blockType)
		return
	}

	var blockDataContent map[string]interface{}
	if blockDataContent, err = re.PrepareGenericBlockData(block["content"]); err != nil {
		if err.Error() != "content not found" {
			return
		}
		err = nil
		blockDataContent = make(map[string]interface{})
	}

	preparedData := re.PrepareGenericBlock("toc", block)

	var blockTag string
	if v, ok := preparedData["Tag"]; ok {
		blockTag, _ = v.(string)
	}

	pageTitle := false
	if v, ok := block["page-title"]; ok {
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
			preparedData["PageTitle"] = "true"
		} else {
			preparedData["PageTitle"] = "false"
		}
	} else {
		preparedData["PageTitle"] = "false"
	}

	var withSelf bool
	if v, ok := block["with-self"]; ok {
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
			preparedData["WithSelf"] = "true"
		} else {
			preparedData["WithSelf"] = "false"
		}
	} else {
		preparedData["WithSelf"] = "false"
	}

	_, _, toc := walkTableOfContents(re, 0, 0, re.GetData())
	items := sortTableOfContents(toc)

	if heading, ok := re.ParseBlockHeader(blockDataContent); ok {
		preparedData["TocHeading"] = heading
		if withSelf {
			items = append([]*tocItem{
				{
					Tag:   blockTag,
					Title: heading,
				},
			}, items...)
		}
	}

	if footer, ok := re.ParseBlockFooter(blockDataContent); ok {
		preparedData["Footer"] = footer
	}

	preparedData["Items"] = items

	if v, ok := block["counter"].(string); ok {
		v = strings.ToLower(v)
		switch v {
		case "nested", "single":
			preparedData["Counter"] = v
		default:
			err = fmt.Errorf("invalid toc counter value: %v", v)
			return
		}
	} else {
		preparedData["Counter"] = "single"
	}

	// log.DebugF("prepared content: %v", preparedData)
	html, err = re.RenderNjnTemplate("block/toc", preparedData)

	return
}