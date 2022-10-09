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
	"sync"

	"github.com/go-enjin/be/pkg/log"
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
	var paths []string
	if paths, err = l.t.FileSystem.ListDirs("layouts"); err != nil {
		err = fmt.Errorf("error listing layouts: %v", err)
		return
	}

	for _, path := range paths {
		name := bePath.Base(path)
		// TODO: figure out better means of hot-reloading, for now, new layouts each time
		// if exists := l.getLayout(name); exists != nil {
		// 	if err = exists.Reload(); err != nil {
		// 		err = fmt.Errorf("%v theme: error reloading %v layout: %v", l.t.Config.Name, name, err)
		// 		return
		// 	}
		// 	log.DebugF("%v theme: reloaded %v layout", l.t.Config.Name, exists.Name)
		// 	continue
		// }
		if layout, e := NewLayout(path, l.t.FileSystem, l.t.FuncMap); e != nil {
			err = fmt.Errorf("error creating new layout: %v - %v", path, e)
			return
		} else {
			l.setLayout(name, layout)
			log.DebugF("%v theme: loaded %v layout", l.t.Config.Name, layout.Name)
		}
	}

	if partials := l.getLayout("partials"); partials != nil {
		l.RLock()
		for k, layout := range l.m {
			if k == "partials" {
				continue
			}
			// add partials to other layouts
			for _, tmpl := range partials.Tmpl.Templates() {
				if _, err = layout.Tmpl.AddParseTree(tmpl.Name(), tmpl.Tree); err != nil {
					return
				}
			}
		}
		l.RUnlock()
	} else {
		log.WarnF("partials layout not found")
	}

	if defaultLayout := l.getLayout("_default"); defaultLayout != nil {
		l.RLock()
		for k, layout := range l.m {
			switch k {
			case "partials", "_default":
				continue
			}
			// add defaults to other layouts
			for _, tmpl := range defaultLayout.Tmpl.Templates() {
				if _, err = layout.Tmpl.AddParseTree(tmpl.Name(), tmpl.Tree); err != nil {
					return
				}
			}
		}
		l.RUnlock()
	}
	return
}

func (l *Layouts) getLayout(name string) (layout *Layout) {
	l.RLock()
	defer l.RUnlock()
	if v, ok := l.m[name]; ok {
		layout = v
	}
	return
}

func (l *Layouts) setLayout(name string, layout *Layout) {
	l.Lock()
	defer l.Unlock()
	l.m[name] = layout
	return
}

func (l *Layouts) AddPartialsToTemplate(tt *template.Template) {
	l.RLock()
	defer l.RUnlock()
	if partials := l.getLayout("partials"); partials != nil {
		for _, tmpl := range partials.Tmpl.Templates() {
			if _, err := tt.AddParseTree(tmpl.Name(), tmpl.Tree); err != nil {
				log.ErrorF("error adding partials to template: %v", err)
			}
		}
	} else {
		log.ErrorF("partials layout not found")
	}
}