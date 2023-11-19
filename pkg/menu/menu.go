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

package menu

import (
	"encoding/json"
	"strconv"
)

type Menu []*Item

func NewMenuFromJson(data []byte) (menu Menu, err error) {
	if err = json.Unmarshal(data, &menu); err != nil {
		return
	}
	return
}

func (m Menu) String() (value string) {
	if data, err := json.MarshalIndent(m, "", "\t"); err == nil {
		value = string(data)
	}
	return
}

func (m Menu) DeepActive() (index string) {
	for idx, item := range m {
		if item.Active {
			index = strconv.Itoa(idx + 1)
			if deep := item.SubMenu.DeepActive(); deep != "" {
				index += "-" + deep
			}
			return
		}
	}
	return
}

func (m Menu) SanitizeAll() {
	SanitizeMenu(m)
	return
}