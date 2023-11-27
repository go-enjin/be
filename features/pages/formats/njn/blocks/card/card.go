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

package card

import (
	"fmt"
	"html/template"

	"github.com/go-enjin/be/pkg/feature"
)

const (
	Tag feature.Tag = "njn-blocks-card"
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
	name = "card"
	return
}

func (f *CBlock) PrepareBlock(re feature.EnjinRenderer, blockType string, data map[string]interface{}) (block map[string]interface{}, redirect string, err error) {
	if blockType != "card" {
		err = fmt.Errorf("%v does not implement %v block type", f.Tag(), blockType)
		return
	}

	var blockDataContent map[string]interface{}
	if blockDataContent, err = re.PrepareGenericBlockData(data["content"]); err != nil {
		return
	}

	block = re.PrepareGenericBlock("card", data)

	block["LinkHref"], _ = data["link-href"]
	block["LinkTarget"], _ = data["link-target"]

	block["Image"], _ = data["image"]
	block["NoImage"], _ = data["no-image"]
	block["Layout"], _ = data["layout"]
	block["ProfileImgSrc"], _ = data["profile-img-src"]

	if background, ok := blockDataContent["background"].(map[string]interface{}); ok {
		if bgType, ok := background["type"].(string); ok {
			if bgType != "picture" {
				err = fmt.Errorf("card block background is not a picture field: %v", bgType)
				return
			}
			if renderBg, e := re.PrepareContainerField(background); e != nil {
				err = e
				return
			} else {
				block["Background"] = renderBg
			}
		} else {
			err = fmt.Errorf("card block missing background type")
			return
		}
	} else {
		err = fmt.Errorf("background type: %T", blockDataContent["background"])
		return
	}

	if heading, ok := re.PrepareBlockHeader(blockDataContent); ok {
		block["Heading"] = heading
	}

	if sections, ok := blockDataContent["section"].([]interface{}); ok {
		// TODO: prevent anchor tags within njn card block content when link-href is set, browsers mangle the DOM in order to make sense of links-within-links
		if block["Section"], err = re.PrepareContainerFields(sections); err != nil {
			return
		}
	}

	// TODO: decide what to do with njn card block footers
	if footer, ok := re.PrepareBlockFooter(blockDataContent); ok {
		block["Footer"] = footer
	}

	block["SiteContext"] = re.RequestContext()
	return
}

func (f *CBlock) RenderPreparedBlock(re feature.EnjinRenderer, block map[string]interface{}) (html template.HTML, err error) {
	html, err = re.RenderNjnTemplate("block/card", block)
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
