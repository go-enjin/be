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

package layouts

import (
	"fmt"
	"strings"
	"sync"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	bePath "github.com/go-enjin/be/pkg/path"
)

var (
	_ feature.ThemeLayouts = (*Layouts)(nil)
)

type Layouts struct {
	theme feature.Theme
	cache map[string]feature.ThemeLayout

	sync.RWMutex
}

func NewLayouts(t feature.Theme) (layouts *Layouts, err error) {
	layouts = &Layouts{
		theme: t,
		cache: make(map[string]feature.ThemeLayout),
	}
	err = layouts.load()
	return
}

func (l *Layouts) load() (err error) {
	l.Lock()
	defer l.Unlock()
	tree := []feature.Theme{
		l.theme,
	}
	for t := l.theme.GetParent(); t != nil; t = t.GetParent() {
		tree = append(tree, t)
	}
	for i := len(tree) - 1; i >= 0; i-- {
		if err = l.loadTheme(tree[i].ThemeFS()); err != nil {
			return
		}
	}
	return
}

func (l *Layouts) loadTheme(tfs fs.FileSystem) (err error) {

	var paths []string
	if paths, err = tfs.ListDirs("layouts"); err != nil {
		if strings.Contains(err.Error(), "no such file or directory") ||
			strings.Contains(err.Error(), "file does not exist") {
			err = nil
			return
		} else {
			err = fmt.Errorf("error listing layouts: %v", err)
		}
		return
	}

	for _, path := range paths {
		name := bePath.Base(path)
		if layout, e := NewLayout(path, tfs); e != nil {
			err = fmt.Errorf("error creating new layout: %v - %v", path, e)
			return
		} else if _, present := l.cache[name]; present {
			l.cache[name].Apply(layout)
		} else {
			l.cache[name] = layout
			log.TraceF("%v theme: loaded %v layout", l.theme.Name(), layout.Name())
		}
	}

	return
}

func (l *Layouts) ListLayouts() (names []string) {
	l.RLock()
	defer l.RUnlock()
	unique := make(map[string]struct{})
	if parent := l.theme.GetParent(); parent != nil {
		if parentLayouts := parent.Layouts(); parentLayouts != nil {
			for _, name := range parentLayouts.ListLayouts() {
				unique[name] = struct{}{}
			}
		}
	}
	for name := range l.cache {
		unique[name] = struct{}{}
	}
	delete(unique, globals.DefaultThemeLayoutName)
	delete(unique, globals.PartialThemeLayoutName)
	names = append([]string{globals.DefaultThemeLayoutName}, maps.SortedKeys(unique)...)
	return
}

func (l *Layouts) GetLayout(name string) (layout feature.ThemeLayout) {
	l.RLock()
	defer l.RUnlock()
	if v, ok := l.cache[name]; ok {
		layout = v
	} else if pt := l.theme.GetParent(); pt != nil {
		layout = pt.Layouts().GetLayout(name)
	}
	return
}

func (l *Layouts) SetLayout(name string, layout feature.ThemeLayout) {
	l.Lock()
	defer l.Unlock()
	l.cache[name] = layout
	return
}
