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

package icon

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/maps"
)

const (
	Tag feature.Tag = "NjnIconBlock"
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
	name = "icon"
	return
}

func (f *CBlock) ProcessBlock(re feature.EnjinRenderer, blockType string, block map[string]interface{}) (html template.HTML, err error) {
	if blockType != "icon" {
		err = fmt.Errorf("%v does not implement %v block type", f.Tag(), blockType)
		return
	}

	var blockDataContent map[string]interface{}
	if blockDataContent, err = re.PrepareGenericBlockData(block["content"]); err != nil {
		return
	}
	preparedData := re.PrepareGenericBlock("icon", block)

	if heading, ok := re.ParseBlockHeader(blockDataContent); ok {
		preparedData["Heading"] = heading
	}

	if iconMap, ok := blockDataContent["icon"].(map[string]interface{}); ok {
		icon := make(map[string]interface{})

		if v, ok := iconMap["align"].(string); ok {
			v = strings.ToLower(v)
			switch v {
			case "left", "center", "right":
				icon["Align"] = v
			default:
				err = fmt.Errorf("invalid icon block icon alignment: %v", v)
				return
			}
		} else {
			icon["Align"] = "center"
		}

		if v, ok := iconMap["href"].(string); ok {
			icon["Href"] = v
		}

		if v, ok := iconMap["target"].(string); ok {
			icon["Target"] = v
		}

		if v, ok := iconMap["caption"].(string); ok {
			icon["Caption"] = v
		}

		if attrs, _, _, e := maps.ParseNjnFieldAttributes(iconMap); e == nil {
			if icon["Attributes"], e = maps.FinalizeNjnFieldAttributes(attrs); e != nil {
				err = fmt.Errorf("error finalizing field attributes: %v", e)
				return
			}
		} else {
			err = fmt.Errorf("error parsing field attributes: %v", e)
			return
		}

		if v, ok := iconMap["class"]; ok {
			switch t := v.(type) {
			case string:
				icon["Class"] = t
			case []interface{}:
				var classes []string
				for _, i := range t {
					if vi, ok := i.(string); ok {
						classes = append(classes, vi)
					} else {
						err = fmt.Errorf("invalid icon block icon class type: %T %+v", i, i)
						return
					}
				}
				icon["class"] = strings.Join(classes, " ")
			}
		} else {
			err = fmt.Errorf("icon block missing icon class: %+v", iconMap)
			return
		}

		if v, ok := iconMap["name"].(string); ok {
			icon["Name"] = v
		} else {
			err = fmt.Errorf("icon block missing icon name: %+v", iconMap)
			return
		}

		preparedData["Icon"] = icon
	}

	if footer, ok := re.ParseBlockFooter(blockDataContent); ok {
		preparedData["Footer"] = footer
	}

	html, err = re.RenderNjnTemplate("block/icon", preparedData)

	return
}