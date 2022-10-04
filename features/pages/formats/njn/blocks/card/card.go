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

package card

import (
	"fmt"
	"html/template"

	"github.com/go-enjin/be/pkg/feature"
)

const (
	Tag feature.Tag = "NjnCardBlock"
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

func (f *CBlock) NjnBlockType() (name string) {
	name = "card"
	return
}

func (f *CBlock) ProcessBlock(re feature.EnjinRenderer, blockType string, block map[string]interface{}) (html template.HTML, err error) {
	if blockType != "card" {
		err = fmt.Errorf("%v does not implement %v block type", f.Tag(), blockType)
		return
	}

	var blockDataContent map[string]interface{}
	if blockDataContent, err = re.PrepareGenericBlockData(block["content"]); err != nil {
		return
	}
	preparedData := re.PrepareGenericBlock("card", block)

	preparedData["LinkHref"], _ = block["link-href"]
	preparedData["LinkTarget"], _ = block["link-target"]

	preparedData["Image"], _ = block["image"]
	preparedData["NoImage"], _ = block["no-image"]
	preparedData["Layout"], _ = block["layout"]
	preparedData["ProfileImgSrc"], _ = block["profile-img-src"]

	if background, ok := blockDataContent["background"].(map[string]interface{}); ok {
		if bgType, ok := background["type"].(string); ok {
			if bgType != "picture" {
				err = fmt.Errorf("card block background is not a picture field: %v", bgType)
				return
			}
			if renderBg, e := re.RenderContainerField(background); e != nil {
				err = e
				return
			} else {
				preparedData["Background"] = renderBg[0]
			}
		} else {
			err = fmt.Errorf("card block missing background type")
			return
		}
	} else {
		err = fmt.Errorf("background type: %T", blockDataContent["background"])
		return
	}

	if heading, ok := re.ParseBlockHeader(blockDataContent); ok {
		preparedData["Heading"] = heading
	}

	if sections, ok := blockDataContent["section"].([]interface{}); ok {
		if preparedData["Section"], err = re.RenderContainerFields(sections); err != nil {
			return
		}
	}

	if footer, ok := re.ParseBlockFooter(blockDataContent); ok {
		preparedData["Footer"] = footer
	}

	html, err = re.RenderNjnTemplate("block/card", preparedData)

	return
}