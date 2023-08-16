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

package theme

import (
	"sync"

	"github.com/go-enjin/be/pkg/feature"
)

var (
	gThemeRegistry = struct {
		known map[string]*Theme
		sync.RWMutex
	}{
		known: make(map[string]*Theme),
	}
)

func (t *Theme) GetParent() (parent feature.Theme) {
	if v := t.getParentTheme(); v != nil {
		parent = v
	}
	return
}

func (t *Theme) registerTheme() {
	gThemeRegistry.Lock()
	defer gThemeRegistry.Unlock()
	gThemeRegistry.known[t.Name()] = t
}

func (t *Theme) getParentTheme() (parent *Theme) {
	if t.config.Parent != "" {
		gThemeRegistry.RLock()
		defer gThemeRegistry.RUnlock()
		if v, ok := gThemeRegistry.known[t.config.Parent]; ok {
			parent = v
		}
	}
	return
}