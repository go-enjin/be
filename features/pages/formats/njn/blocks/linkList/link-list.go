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

package linkList

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/maps"
)

const (
	Tag feature.Tag = "njn-blocks-link-list"
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
	name = "link-list"
	return
}

func (f *CBlock) PrepareBlock(re feature.EnjinRenderer, blockType string, data map[string]interface{}) (block map[string]interface{}, redirect string, err error) {
	if blockType != "link-list" {
		err = fmt.Errorf("%v does not implement %v block type", f.Tag(), blockType)
		return
	}

	var blockDataContent map[string]interface{}
	if blockDataContent, err = re.PrepareGenericBlockData(data["content"]); err != nil {
		return
	}

	block = re.PrepareGenericBlock("link-list", data)

	if heading, ok := re.PrepareBlockHeader(blockDataContent); ok {
		block["Heading"] = heading
	}

	if sections, ok := blockDataContent["section"].([]interface{}); ok {
		sectionFields := make([]interface{}, 0)
		for _, si := range sections {
			switch st := si.(type) {
			case map[string]interface{}:
				if name, ok := st["type"].(string); ok {
					if name == "a" {
						st["decorated"] = "true"
						if attrs, classes, _, e := maps.ParseNjnFieldAttributes(st); e == nil {
							classes = append(classes, "decorated")
							attrs["class"] = strings.Join(classes, " ")
							if st["attributes"], e = maps.FinalizeNjnFieldAttributes(attrs); e != nil {
								err = e
								return
							}
						} else {
							if st["attributes"], e = maps.FinalizeNjnFieldAttributes(map[string]interface{}{
								"class": "decorated",
							}); e != nil {
								err = e
								return
							}
						}
						sectionFields = append(sectionFields, st)
					} else {
						err = fmt.Errorf("link-list block has more than just anchor tags: %+v", st)
						return
					}
				}
			}
		}

		if block["Section"], err = re.PrepareContainerFieldList([]interface{}{
			map[string]interface{}{
				"type": "ul",
				"list": sectionFields,
			},
		}); err != nil {
			return
		}
	}

	if footer, ok := re.PrepareBlockFooter(blockDataContent); ok {
		block["Footer"] = footer
	}

	block["SiteContext"] = re.RequestContext()
	return
}

func (f *CBlock) RenderPreparedBlock(re feature.EnjinRenderer, block map[string]interface{}) (html template.HTML, err error) {
	html, err = re.RenderNjnTemplate("block/link-list", block)
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