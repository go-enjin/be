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

package theme

import (
	"fmt"

	"github.com/go-enjin/be/pkg/theme/types"
)

var (
	knownThemes = map[string]*Theme{}
)

func addThemeInstance(t *Theme) (err error) {
	var ok bool
	if _, ok = knownThemes[t.Name]; ok {
		err = fmt.Errorf(`duplicate theme instance: "%v"`, t.Name)
	} else {
		knownThemes[t.Name] = t
	}
	return
}

func getParentTheme(parent string) (t *Theme) {
	if parent != "" {
		if v, ok := knownThemes[parent]; ok {
			t = v
		}
	}
	return
}

func (t *Theme) GetParent() (parent *Theme) {
	parent = getParentTheme(t.Config.Parent)
	return
}

func (t *Theme) GetParentTheme() (parent types.Theme) {
	parent = t.GetParent()
	return
}