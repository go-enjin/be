// Copyright (c) 2024  The Go-Enjin Authors
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

//go:build !exclude_pages_formats && !exclude_pages_format_njn

package njn_tagcloud

import (
	"fmt"
	"html/template"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/features/pages/tagcloud"
	"github.com/go-enjin/be/pkg/feature"
)

const (
	Tag feature.Tag = "tagcloud-njn-block"
)

var (
	_ Block     = (*CBlock)(nil)
	_ MakeBlock = (*CBlock)(nil)
)

type Block interface {
	feature.EnjinBlock
}

type MakeBlock interface {
	SetTagCloud(tag feature.Tag) MakeBlock

	Make() Block
}

type CBlock struct {
	feature.CEnjinBlock

	providerTag feature.Tag
	provider    feature.TagCloudProvider
}

func New() (field MakeBlock) {
	f := new(CBlock)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = Tag
	f.providerTag = tagcloud.Tag
	return f
}

func (f *CBlock) Init(this interface{}) {
	f.CEnjinBlock.Init(this)
}

func (f *CBlock) SetTagCloud(tag feature.Tag) MakeBlock {
	f.providerTag = tag
	return f
}

func (f *CBlock) Make() Block {
	return f
}

func (f *CBlock) PostStartup(ctx *cli.Context) (err error) {
	if f.providerTag.IsNil() {
		err = fmt.Errorf("%v feature requires .SetTagCloud", f.Tag())
		return
	}
	if f.provider, err = feature.GetTyped[feature.TagCloudProvider](f.providerTag, f.Enjin.Features().List()); f.provider == feature.TagCloudProvider(nil) {
		err = fmt.Errorf("%v feature did not find the %q feature.TagCloudProvider", f.Tag(), f.providerTag)
	}
	return
}

func (f *CBlock) NjnClass() (tagClass feature.NjnClass) {
	tagClass = feature.InlineNjnClass
	return
}

func (f *CBlock) NjnBlockType() (name string) {
	name = "tagcloud"
	return
}

func (f *CBlock) PrepareBlock(re feature.EnjinRenderer, blockType string, data map[string]interface{}) (block map[string]interface{}, redirect string, err error) {
	if blockType != "tagcloud" {
		err = fmt.Errorf("%v does not implement %v block type", f.Tag(), blockType)
		return
	}

	var blockDataContent map[string]interface{}
	if blockDataContent, err = re.PrepareGenericBlockData(data["content"]); err != nil {
		return
	}

	block = re.PrepareGenericBlock("tagcloud", data)

	if heading, ok := re.PrepareBlockHeader(blockDataContent); ok {
		block["Heading"] = heading
	}

	tc := f.provider.GetTagCloud()
	tc.Sort()
	block["TagCloud"] = tc

	if footer, ok := re.PrepareBlockFooter(blockDataContent); ok {
		block["Footer"] = footer
	}

	block["SiteContext"] = re.RequestContext()
	return
}

func (f *CBlock) RenderPreparedBlock(re feature.EnjinRenderer, block map[string]interface{}) (html template.HTML, err error) {
	html, err = re.RenderNjnTemplate("block/tagcloud", block)
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
