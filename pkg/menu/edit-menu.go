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

	"github.com/go-corelibs/slices"
)

type EditMenu []*EditItem

func NewEditMenuFromJson(data []byte) (menu EditMenu, err error) {
	if err = json.Unmarshal(data, &menu); err != nil {
		return
	}
	return
}

func (m EditMenu) Len() (size int) {
	size = len(m)
	return
}

func (m EditMenu) String() (value string) {
	if data, err := json.MarshalIndent(m, "", "\t"); err == nil {
		value = string(data)
	}
	return
}

func (m EditMenu) AsMenu() (clone Menu) {
	for _, item := range m {
		clone = append(clone, item.AsItem())
	}
	return
}

func (m EditMenu) ExpandAll() {
	for _, item := range m {
		item.Expand = "true"
		item.SubMenu.ExpandAll()
	}
}

func (m EditMenu) CollapseAll() {
	for _, item := range m {
		item.Expand = ""
		item.SubMenu.CollapseAll()
	}
}

func (m EditMenu) SanitizeAll() {
	SanitizeMenu(m)
}

func (m EditMenu) ProcessAllChanges() (modified EditMenu) {
	var changed bool
	for modified, changed = m.ProcessChanges(); changed; modified, changed = modified.ProcessChanges() {
		// nop
	}
	return
}

func (m EditMenu) ProcessChanges() (modified EditMenu, changed bool) {
	var process func(list EditMenu, top bool) (spill *EditItem, forward bool, modified EditMenu, changed bool)
	process = func(list EditMenu, top bool) (spill *EditItem, forward bool, modified EditMenu, changed bool) {
		if last := len(list) - 1; last > -1 {
			// depth first
			for idx, item := range list {
				var s *EditItem
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
			for idx, item := range append(EditMenu{}, modified...) {
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
					item.SubMenu = append(item.SubMenu, &EditItem{})
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
