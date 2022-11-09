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

package carousel

import (
	"fmt"
	"html/template"
	"strconv"

	"github.com/go-enjin/be/pkg/feature"
)

const (
	Tag feature.Tag = "NjnCarouselBlock"
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
	tagClass = feature.ContainerNjnClass
	return
}

func (f *CBlock) NjnBlockType() (name string) {
	name = "carousel"
	return
}

func (f *CBlock) PrepareBlock(re feature.EnjinRenderer, blockType string, data map[string]interface{}) (block map[string]interface{}, err error) {
	if blockType != "carousel" {
		err = fmt.Errorf("%v does not implement %v block type", f.Tag(), blockType)
		return
	}

	var blockDataContent map[string]interface{}
	if blockDataContent, err = re.PrepareGenericBlockData(data["content"]); err != nil {
		return
	}

	block = re.PrepareGenericBlock("carousel", data)

	if cardGap, ok := data["card-gap"].(string); ok {
		switch cardGap {
		case "standard":
			block["CardGap"] = cardGap
		default:
			err = fmt.Errorf("unsupported card-gap: %v", cardGap)
			return
		}
	}

	if navCtrlType, ok := data["nav-ctrl-type"].(string); ok {
		switch navCtrlType {
		case "data", "icon":
			block["NavCtrlType"] = navCtrlType
		default:
			err = fmt.Errorf("unsupported nav-ctrl-type: %v", navCtrlType)
			return
		}
	}

	if navCtrlStyle, ok := data["nav-ctrl-style"].(string); ok {
		switch navCtrlStyle {
		case "chevron", "arrow", "caret", "chevron-circle", "arrow-circle":
			block["NavCtrlStyle"] = navCtrlStyle
		default:
			err = fmt.Errorf("unsupported nav-ctrl-style: %v", navCtrlStyle)
			return
		}
	}

	if heading, ok := re.PrepareBlockHeader(blockDataContent); ok {
		block["Heading"] = heading
	}

	var cards []map[string]interface{}
	if sections, ok := blockDataContent["section"].([]interface{}); ok {
		re.IncCurrentDepth()
		for _, section := range sections {
			if cardBlock, name, ok := re.ParseFieldAndTypeName(section); ok {
				if name == "card" {
					if card, e := re.PrepareBlock(cardBlock); e != nil {
						err = e
						return
					} else {
						card["CardIndex"] = len(cards)
						cards = append(cards, card)
					}
				} else {
					err = fmt.Errorf("carousel item is not a card: %v", name)
					return
				}
			}
		}
		re.DecCurrentDepth()
	}

	numCards := len(cards)
	if numCards < 1 {
		err = fmt.Errorf("at least one card is required")
		return
	}

	bookends := 0
	if bookendsValue, ok := data["bookends"]; ok {
		switch bookendsType := bookendsValue.(type) {
		case int:
			bookends = bookendsType
		case string:
			bookends, _ = strconv.Atoi(bookendsType)
		case float64:
			bookends = int(bookendsType)
		default:
			err = fmt.Errorf("unsupported bookends structure: %T", bookendsType)
			return
		}
	}

	if bookends > 2 {
		err = fmt.Errorf("too many bookends specified (%d), 0-2 allowed", bookends)
		return
	}

	block["Bookends"] = bookends

	for idx, card := range cards {
		switch idx {
		case 0:
			// first wraps to last
			card["PreviousCard"] = cards[numCards-1]
			card["PreviousCardIndex"] = numCards - 1
			card["NextCard"] = cards[idx+1]
			card["NextCardIndex"] = idx + 1
		case numCards - 1:
			// last wraps to first
			card["PreviousCard"] = cards[idx-1]
			card["PreviousCardIndex"] = idx - 1
			card["NextCard"] = cards[0]
			card["NextCardIndex"] = 0
		default:
			// middle cards
			card["PreviousCardIndex"] = idx - 1
			card["PreviousCard"] = cards[idx-1]
			card["NextCard"] = cards[idx+1]
			card["NextCardIndex"] = idx + 1
		}
	}

	block["Cards"] = cards
	block["LastCard"] = numCards - 1

	if block["Footnotes"], err = re.PrepareFootnotes(re.GetBlockIndex()); err != nil {
		return
	}

	if footer, ok := re.PrepareBlockFooter(blockDataContent); ok {
		block["Footer"] = footer
	}

	return
}

func (f *CBlock) RenderPreparedBlock(re feature.EnjinRenderer, block map[string]interface{}) (html template.HTML, err error) {
	html, err = re.RenderNjnTemplate("block/carousel", block)
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