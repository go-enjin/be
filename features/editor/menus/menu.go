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

package menus

import (
	"encoding/json"

	"github.com/mrz1836/go-sanitize"

	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/menu"
	"github.com/go-enjin/be/pkg/slices"
)

type Item struct {
	Text string `json:"text"`
	Href string `json:"href,omitempty"`
	Lang string `json:"lang,omitempty"`

	Active bool `json:"active,omitempty"`

	Attributes map[string]string `json:"attributes,omitempty"`

	SubMenu Menu `json:"sub-menu,omitempty"`

	Expand   string `json:"expand,omitempty"`
	Delete   bool   `json:"delete,omitempty"`
	MoveUp   bool   `json:"move-up,omitempty"`
	MoveDown bool   `json:"move-down,omitempty"`
	Append   bool   `json:"append,omitempty"`
}

func (i Item) String() (value string) {
	if data, err := json.MarshalIndent(i, "", "\t"); err == nil {
		value = string(data)
	}
	return
}

func (i Item) AsItem() (clone *menu.Item) {
	var subMenu menu.Menu
	if len(i.SubMenu) > 0 {
		subMenu = i.SubMenu.AsMenu()
	}
	clone = &menu.Item{
		Text:       i.Text,
		Href:       i.Href,
		Lang:       i.Lang,
		Active:     i.Active,
		Attributes: i.Attributes,
		SubMenu:    subMenu,
	}
	return
}

type Menu []*Item

func NewMenuFromJson(data []byte) (menu Menu, err error) {
	if err = json.Unmarshal(data, &menu); err != nil {
		return
	}
	return
}

func (m Menu) Len() (size int) {
	size = len(m)
	return
}

func (m Menu) String() (value string) {
	if data, err := json.MarshalIndent(m, "", "\t"); err == nil {
		value = string(data)
	}
	return
}

func (m Menu) AsMenu() (clone menu.Menu) {
	for _, item := range m {
		clone = append(clone, item.AsItem())
	}
	return
}

func (m Menu) ExpandAll() {
	for _, item := range m {
		item.Expand = "true"
		item.SubMenu.ExpandAll()
	}
}

func (m Menu) CollapseAll() {
	for _, item := range m {
		item.Expand = ""
		item.SubMenu.CollapseAll()
	}
}

func (m Menu) SanitizeAll() {
	for _, item := range m {
		item.Text = forms.StrictSanitize(item.Text)
		item.Href = sanitize.URL(item.Href)
	}
}

func (m Menu) ProcessAllChanges() (modified Menu) {
	var changed bool
	for modified, changed = m.ProcessChanges(); changed; modified, changed = modified.ProcessChanges() {
		// nop
	}
	return
}

func (m Menu) ProcessChanges() (modified Menu, changed bool) {
	var process func(list Menu, top bool) (spill *Item, forward bool, modified Menu, changed bool)
	process = func(list Menu, top bool) (spill *Item, forward bool, modified Menu, changed bool) {
		if last := len(list) - 1; last > -1 {
			// depth first
			for idx, item := range list {
				var s *Item
				var fwd bool
				if s, fwd, item.SubMenu, changed = process(item.SubMenu, false); s != nil {
					if fwd {
						if idx < last {
							// push after this
							modified = append(modified, item, s)
						}
					} else {
						if idx > 0 {
							// push before this
							modified = append(modified, s, item)
						}
					}
				} else {
					modified = append(modified, item)
				}
			}
			// then this
			last = len(modified) - 1 // revised last index
			for idx, item := range append(Menu{}, modified...) {
				if changed = item.MoveUp; changed {
					item.MoveUp = false
					modified = slices.Remove(modified, idx)
					if idx == 0 {
						if !top {
							spill = item
							forward = false
						}
					} else {
						modified = slices.Insert(modified, idx-1, item)
					}
					return
				} else if changed = item.MoveDown; changed {
					item.MoveDown = false
					modified = slices.Remove(modified, idx)
					if idx == last {
						if !top {
							spill = item
							forward = true
						}
					} else {
						modified = slices.Insert(modified, idx+1, item)
					}
					return
				} else if changed = item.Append; changed {
					item.Append = false
					item.SubMenu = append(item.SubMenu, &Item{})
					return
				} else if changed = item.Delete; changed {
					item.Delete = false
					modified = slices.Remove(modified, idx)
					return
				}
			}
		}
		return
	}
	_, _, modified, changed = process(m, true)
	return
}