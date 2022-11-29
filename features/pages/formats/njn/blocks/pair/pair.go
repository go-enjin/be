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

package pair

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/go-enjin/be/pkg/feature"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

const (
	Tag feature.Tag = "NjnPairBlock"
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
	name = "pair"
	return
}

func (f *CBlock) PrepareBlock(re feature.EnjinRenderer, blockType string, data map[string]interface{}) (block map[string]interface{}, redirect string, err error) {
	if blockType != "pair" {
		err = fmt.Errorf("%v does not implement %v block type", f.Tag(), blockType)
		return
	}

	var blockDataContent map[string]interface{}
	if blockDataContent, err = re.PrepareGenericBlockData(data["content"]); err != nil {
		return
	}

	block = re.PrepareGenericBlock("pair", data)

	pairOrderReversed := false
	if v, ok := data["pair-order"].(string); ok {
		v = strings.ToLower(v)
		switch v {
		case "normal":
		case "reverse":
			pairOrderReversed = true
		default:
			err = fmt.Errorf("unknown pair block order requested: %v", v)
			return
		}
		block["PairOrder"] = v
	}
	if heading, ok := re.PrepareBlockHeader(blockDataContent); ok {
		block["Heading"] = heading
	}

	if sections, ok := blockDataContent["section"].([]interface{}); ok {
		if len(sections) != 2 {
			err = fmt.Errorf("pair block requires two items, %d present", len(sections))
			return
		}
		var combined []map[string]interface{}
		re.IncCurrentDepth()
		for idx, section := range sections {
			if sectionBlock, ok := section.(map[string]interface{}); ok {
				sectionBlockType, _ := re.ParseTypeName(sectionBlock)
				if idx == 0 {
					sectionBlock = beStrings.AddClassNamesToNjnBlock(sectionBlock, "first", sectionBlockType)
				} else {
					sectionBlock = beStrings.AddClassNamesToNjnBlock(sectionBlock, "second", sectionBlockType)
				}
				// TODO: restrict pair block contents to only inline blocks
				if prepared, redir, e := re.PrepareBlock(sectionBlock); e != nil {
					err = e
					return
				} else if redir != "" {
					redirect = redir
					return
				} else {
					combined = append(combined, prepared)
				}
			}
		}
		re.DecCurrentDepth()

		if pairOrderReversed {
			combined = append(
				[]map[string]interface{}{
					combined[1],
				},
				combined[0],
			)
		}

		block["Section"] = combined
	}

	if block["Footnotes"], err = re.PrepareFootnotes(re.GetBlockIndex()); err != nil {
		return
	}

	if footer, ok := re.PrepareBlockFooter(blockDataContent); ok {
		block["Footer"] = footer
	}

	block["SiteContext"] = re.RequestContext()
	return
}

func (f *CBlock) RenderPreparedBlock(re feature.EnjinRenderer, block map[string]interface{}) (html template.HTML, err error) {
	html, err = re.RenderNjnTemplate("block/pair", block)
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