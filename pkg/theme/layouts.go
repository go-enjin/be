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
	"html/template"
	"strings"
	"sync"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	bePath "github.com/go-enjin/be/pkg/path"
)

type Layouts struct {
	t *Theme
	m map[string]*Layout

	sync.RWMutex
}

func NewLayouts(t *Theme) (layouts *Layouts, err error) {
	layouts = &Layouts{
		t: t,
		m: make(map[string]*Layout),
	}
	err = layouts.Reload()
	return
}

func (l *Layouts) Reload() (err error) {
	l.Lock()
	defer l.Unlock()

	var paths []string
	if paths, err = l.t.FileSystem.ListDirs("layouts"); err != nil {
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
		if exists := l.GetLayout(name); exists != nil {
			if err = exists.Reload(); err != nil {
				err = fmt.Errorf("%v theme: error reloading %v layout: %v", l.t.Config.Name, name, err)
				return
			}
			log.DebugF("%v theme: reloaded %v layout", l.t.Config.Name, exists.Name)
			continue
		}
		if layout, e := NewLayout(path, l.t.FileSystem, l.t); e != nil {
			err = fmt.Errorf("error creating new layout: %v - %v", path, e)
			return
		} else {
			l.SetLayout(name, layout)
			log.TraceF("%v theme: loaded %v layout", l.t.Config.Name, layout.Name)
		}
	}

	return
}

func (l *Layouts) GetLayout(name string) (layout *Layout) {
	l.RLock()
	defer l.RUnlock()
	if v, ok := l.m[name]; ok {
		layout = v
	}
	return
}

func (l *Layouts) SetLayout(name string, layout *Layout) {
	l.Lock()
	defer l.Unlock()
	l.m[name] = layout
	return
}

func (l *Layouts) NewTemplate(name string, ctx context.Context) (tmpl *template.Template, err error) {
	l.RLock()
	defer l.RUnlock()

	if tmpl, err = template.New(name).Parse(`{{/* empty */}}`); err == nil {

		if partials := l.GetLayout("partials"); partials != nil {
			if err = partials.Apply(tmpl, ctx); err != nil {
				return
			}
		}

		if _default := l.GetLayout("_default"); _default != nil {
			if err = _default.Apply(tmpl, ctx); err != nil {
				return
			}
		}

		for _, layoutName := range maps.SortedKeys(l.m) {
			switch layoutName {
			case "partials", "_default":
				continue
			}

			if layout, ok := l.m[layoutName]; ok {
				if err = layout.Apply(tmpl, ctx); err != nil {
					return
				}
			} else {
				log.ErrorF("inconsistent cache key: %v", layoutName)
			}
		}

	}
	return
}