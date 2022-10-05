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
	Tag feature.Tag = "NjnLinkListBlock"
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
	name = "link-list"
	return
}

func (f *CBlock) ProcessBlock(re feature.EnjinRenderer, blockType string, block map[string]interface{}) (html template.HTML, err error) {
	if blockType != "link-list" {
		err = fmt.Errorf("%v does not implement %v block type", f.Tag(), blockType)
		return
	}

	var blockDataContent map[string]interface{}
	if blockDataContent, err = re.PrepareGenericBlockData(block["content"]); err != nil {
		return
	}
	preparedData := re.PrepareGenericBlock("link-list", block)

	if heading, ok := re.ParseBlockHeader(blockDataContent); ok {
		preparedData["Heading"] = heading
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

		if preparedData["Section"], err = re.RenderContainerFields([]interface{}{
			map[string]interface{}{
				"type": "ul",
				"list": sectionFields,
			},
		}); err != nil {
			return
		}
	}

	if footer, ok := re.ParseBlockFooter(blockDataContent); ok {
		preparedData["Footer"] = footer
	}

	html, err = re.RenderNjnTemplate("block/link-list", preparedData)

	return
}