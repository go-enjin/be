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

package image

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/go-enjin/be/pkg/feature"
)

const (
	Tag feature.Tag = "NjnImageBlock"
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
	name = "image"
	return
}

func (f *CBlock) ProcessBlock(re feature.EnjinRenderer, blockType string, block map[string]interface{}) (html template.HTML, err error) {
	if blockType != "image" {
		err = fmt.Errorf("%v does not implement %v block type", f.Tag(), blockType)
		return
	}

	var blockDataContent map[string]interface{}
	if blockDataContent, err = re.PrepareGenericBlockData(block["content"]); err != nil {
		return
	}
	preparedData := re.PrepareGenericBlock("image", block)

	if v, ok := block["constraint"].(string); ok {
		v = strings.ToLower(v)
		switch v {
		case "width", "height":
			preparedData["Constraint"] = v
		default:
			err = fmt.Errorf("invalid image block constraint: %v", v)
			return
		}
	} else {
		preparedData["Constraint"] = "width"
	}

	if v, ok := block["fitting"].(string); ok {
		v = strings.ToLower(v)
		switch v {
		case "cover", "fill", "contain", "none", "scale-down":
			preparedData["Fitting"] = v
		default:
			err = fmt.Errorf("invalid image block fitting: %v", v)
			return
		}
	} else {
		preparedData["Fitting"] = "cover"
	}

	if v, ok := block["position"].(string); ok {
		v = strings.ToLower(v)
		switch v {
		case "center", "top", "top-left", "left", "bottom-left", "bottom", "bottom-right", "right", "top-right":
			preparedData["Position"] = v
		default:
			err = fmt.Errorf("invalid image block position: %v", v)
			return
		}
	} else {
		preparedData["Position"] = "center"
	}

	if v, ok := block["size"].(string); ok {
		v = strings.ToLower(v)
		switch v {
		case "sliver", "thin", "banner", "normal", "tall", "huge", "actual":
			preparedData["Size"] = v
		default:
			err = fmt.Errorf("invalid image block size: %v", v)
			return
		}
	} else {
		preparedData["Size"] = "normal"
	}

	if heading, ok := re.ParseBlockHeader(blockDataContent); ok {
		preparedData["Heading"] = heading
	}

	if picture, ok := blockDataContent["picture"].(map[string]interface{}); ok {
		var combine []template.HTML
		if combine, err = re.RenderContainerFields([]interface{}{picture}); err != nil {
			return
		} else {
			var combined template.HTML
			for _, comb := range combine {
				combined += comb
			}
			preparedData["Picture"] = combined
		}
	} else {
		err = fmt.Errorf("image block missing images: %+v", block)
		return
	}

	if footer, ok := re.ParseBlockFooter(blockDataContent); ok {
		preparedData["Footer"] = footer
	}

	html, err = re.RenderNjnTemplate("block/image", preparedData)
	return
}