// Copyright (c) 2023  The Go-Enjin Authors
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

package menu

import (
	"encoding/json"

	"github.com/go-enjin/be/pkg/context"
)

type EditItem struct {
	Text   string `json:"text"`
	Href   string `json:"href,omitempty"`
	Lang   string `json:"lang,omitempty"`
	Target string `json:"target,omitempty"`

	Icon   string `json:"icon,omitempty"`
	Image  string `json:"image,omitempty"`
	ImgAlt string `json:"img-alt,omitempty"`

	Active bool `json:"active,omitempty"`

	SubMenu EditMenu `json:"sub-menu,omitempty"`

	Context context.Context `json:"context,omitempty"`

	Expand       string `json:"expand,omitempty"`
	ExpandExtras string `json:"expand-extras,omitempty"`
	Delete       bool   `json:"delete,omitempty"`
	MoveUp       bool   `json:"move-up,omitempty"`
	MoveDown     bool   `json:"move-down,omitempty"`
	Append       bool   `json:"append,omitempty"`
}

func (i EditItem) String() (value string) {
	if data, err := json.MarshalIndent(i, "", "\t"); err == nil {
		value = string(data)
	}
	return
}

func (i EditItem) AsItem() (clone *Item) {
	var subMenu Menu
	if len(i.SubMenu) > 0 {
		subMenu = i.SubMenu.AsMenu()
	}
	clone = &Item{
		Text:    i.Text,
		Href:    i.Href,
		Lang:    i.Lang,
		Target:  i.Target,
		Icon:    i.Icon,
		Image:   i.Image,
		ImgAlt:  i.ImgAlt,
		Active:  i.Active,
		SubMenu: subMenu,
		Context: i.Context.Copy(),
	}
	return
}
