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

func (f *CBlock) NjnClass() (tagClass feature.NjnClass) {
	tagClass = feature.InlineNjnClass
	return
}

func (f *CBlock) NjnBlockType() (name string) {
	name = "image"
	return
}

func (f *CBlock) PrepareBlock(re feature.EnjinRenderer, blockType string, data map[string]interface{}) (block map[string]interface{}, err error) {
	if blockType != "image" {
		err = fmt.Errorf("%v does not implement %v block type", f.Tag(), blockType)
		return
	}

	var blockDataContent map[string]interface{}
	if blockDataContent, err = re.PrepareGenericBlockData(data["content"]); err != nil {
		return
	}

	block = re.PrepareGenericBlock("image", data)

	if v, ok := data["constraint"].(string); ok {
		v = strings.ToLower(v)
		switch v {
		case "width", "height":
			block["Constraint"] = v
		default:
			err = fmt.Errorf("invalid image block constraint: %v", v)
			return
		}
	}

	if v, ok := data["fitting"].(string); ok {
		v = strings.ToLower(v)
		switch v {
		case "cover", "fill", "contain", "none", "scale-down":
			block["Fitting"] = v
		default:
			err = fmt.Errorf("invalid image block fitting: %v", v)
			return
		}
	} else {
		block["Fitting"] = "cover"
	}

	if v, ok := data["position"].(string); ok {
		v = strings.ToLower(v)
		switch v {
		case "center", "top", "top-left", "left", "bottom-left", "bottom", "bottom-right", "right", "top-right":
			block["Position"] = v
		default:
			err = fmt.Errorf("invalid image block position: %v", v)
			return
		}
	} else {
		block["Position"] = "center"
	}

	if v, ok := data["size"].(string); ok {
		v = strings.ToLower(v)
		switch v {
		case "sliver", "thin", "banner", "normal", "tall", "huge", "actual":
			block["Size"] = v
		default:
			err = fmt.Errorf("invalid image block size: %v", v)
			return
		}
	} else {
		block["Size"] = "normal"
	}

	if heading, ok := re.PrepareBlockHeader(blockDataContent); ok {
		block["Heading"] = heading
	}

	if picture, ok := blockDataContent["picture"].(map[string]interface{}); ok {
		if combined, e := re.PrepareContainerFields([]interface{}{picture}); e != nil {
			err = e
			return
		} else {
			block["Picture"] = combined
		}
	} else {
		err = fmt.Errorf("image block missing images: %+v", data)
		return
	}

	if footer, ok := re.PrepareBlockFooter(blockDataContent); ok {
		block["Footer"] = footer
	}

	return
}

func (f *CBlock) RenderPreparedBlock(re feature.EnjinRenderer, block map[string]interface{}) (html template.HTML, err error) {
	html, err = re.RenderNjnTemplate("block/image", block)
	return
}

func (f *CBlock) ProcessBlock(re feature.EnjinRenderer, blockType string, data map[string]interface{}) (html template.HTML, err error) {
	if block, e := f.PrepareBlock(re, blockType, data); e != nil {
		err = e
		return
	} else {
		html, err = f.RenderPreparedBlock(re, block)
	}
	return
}