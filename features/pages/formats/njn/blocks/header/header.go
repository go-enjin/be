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

package header

import (
	"fmt"
	"html/template"

	"github.com/go-enjin/be/features/pages/formats/njn/fields/anchor"
	"github.com/go-enjin/be/pkg/feature"
)

const (
	Tag feature.Tag = "NjnHeaderBlock"
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
	name = "header"
	return
}

func (f *CBlock) ProcessBlock(re feature.EnjinRenderer, blockType string, block map[string]interface{}) (html template.HTML, err error) {
	if blockType != "header" {
		err = fmt.Errorf("%v does not implement %v block type", f.Tag(), blockType)
		return
	}

	var blockDataContent map[string]interface{}
	if blockDataContent, err = re.PrepareGenericBlockData(block["content"]); err != nil {
		return
	}

	var hr, hl, hc int
	hc = re.GetHeadingCount()
	hl = re.GetHeadingLevel()
	hl, hr, _ /*hl*/ = re.ParseBlockHeadingLevel(hc, hl, block)
	re.IncHeadingCount() // total number of header blocks on the page
	re.SetHeadingLevel(hl)

	preparedData := re.PrepareGenericBlock("header", block)

	// tag, _ := block["tag"]
	// log.DebugF("tag=%v, count=%v, level=%v, hr=%v, hl=%v", tag, re.headingCount, re.headingLevel, hr, hl)
	if hr == -255 /*&& hl == -255*/ {
		re.IncHeadingLevel() // header blocks cause further blocks to be level+1
	}

	if heading, ok := re.ParseBlockHeader(blockDataContent); ok {
		preparedData["Heading"] = heading
	}

	anchorField := anchor.New().Make()

	if list, ok := blockDataContent["nav"].([]interface{}); ok {
		var navItems []map[string]interface{}
		for _, item := range list {
			var navItem map[string]interface{}
			if v, ok := item.(map[string]interface{}); ok {
				if vType, ok := v["type"].(string); ok {

					switch vType {
					case "a":

						if navItem, err = anchorField.PrepareNjnData(re, "a", v); err != nil {
							return
						}

					default:
						err = fmt.Errorf("unsupported heading nav item type: %v", vType)
						return
					}

				} else {
					err = fmt.Errorf("heading nav item missing type: %+v", v)
					return
				}
			}
			navItems = append(navItems, navItem)
		}
		preparedData["Nav"] = navItems
	}

	if footer, ok := re.ParseBlockFooter(blockDataContent); ok {
		preparedData["Footer"] = footer
	}

	// log.DebugF("prepared header: %v", preparedData)
	html, err = re.RenderNjnTemplate("block/header", preparedData)

	return
}