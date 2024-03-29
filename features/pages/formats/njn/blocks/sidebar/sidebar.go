//go:build !exclude_pages_formats && !exclude_pages_format_njn

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

package sidebar

import (
	"fmt"
	"html/template"

	"github.com/go-enjin/be/pkg/feature"
)

const (
	Tag feature.Tag = "njn-blocks-sidebar"
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
	f.PackageTag = Tag
	f.FeatureTag = Tag
	return f
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
	name = "sidebar"
	return
}

func (f *CBlock) PrepareBlock(re feature.EnjinRenderer, blockType string, data map[string]interface{}) (block map[string]interface{}, redirect string, err error) {
	if blockType != "sidebar" {
		err = fmt.Errorf("%v does not implement %v block type", f.Tag(), blockType)
		return
	}

	var blockDataContent map[string]interface{}
	if blockDataContent, err = re.PrepareGenericBlockData(data["content"]); err != nil {
		return
	}

	block = re.PrepareGenericBlock("sidebar", data)

	block["Side"], _ = data["side"]
	block["Stack"], _ = data["stack"]
	block["Sticky"], _ = data["sticky"]

	if heading, ok := re.PrepareBlockHeader(blockDataContent); ok {
		block["Heading"] = heading
	}

	if contentBlocks, ok := blockDataContent["blocks"].([]interface{}); ok {
		var combined []map[string]interface{}
		for _, contentBlock := range contentBlocks {
			if contentBlockData, ok := contentBlock.(map[string]interface{}); ok {
				if prepared, redir, e := re.PrepareBlock(contentBlockData); e != nil {
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
		block["Blocks"] = combined
	}

	if asides, ok := blockDataContent["aside"].([]interface{}); ok {
		var combined []map[string]interface{}
		re.IncCurrentDepth()
		re.SetWithinAside(true)
		for _, aside := range asides {
			if asideBlock, ok := aside.(map[string]interface{}); ok {
				if prepared, redir, e := re.PrepareBlock(asideBlock); e != nil {
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
		re.SetWithinAside(false)
		re.DecCurrentDepth()
		block["Aside"] = combined
	}

	// block["Footnotes"] = re.PrepareFootnotes(re.GetBlockIndex())

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
