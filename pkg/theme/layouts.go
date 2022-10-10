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
		if strings.Contains(err.Error(), "no such file or directory") {
			err = nil
			return
		} else {
			err = fmt.Errorf("error listing layouts: %v", err)
		}
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
			log.TraceF("%v theme: loaded %v layout", l.t.Config.Name, layout.Name)
		}
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
	}
}

func (l *Layouts) AddLayoutsToTemplate(tt *template.Template) {
	l.RLock()
	defer l.RUnlock()

	partials := l.getLayout("partials")
	defaultLayout := l.getLayout("_default")

	if defaultLayout != nil {
		if partials != nil {
			for _, tmpl := range partials.Tmpl.Templates() {
				if _, err := defaultLayout.Tmpl.AddParseTree(tmpl.Name(), tmpl.Tree); err != nil {
					log.ErrorF("error adding %v to template: %v", tmpl.Name(), err)
				} else {
					// log.DebugF("adding %v template to _default layout", tmpl.Name())
				}
			}
		}
	}

	for name, layout := range l.m {
		switch name {
		case "partials", "_default":
			continue
		}
		if partials != nil {
			for _, tmpl := range partials.Tmpl.Templates() {
				if _, err := tt.AddParseTree(tmpl.Name(), tmpl.Tree); err != nil {
					log.ErrorF("error adding %v to template: %v", tmpl.Name(), err)
				} else {
					// log.DebugF("adding %v template to %v layout", tmpl.Name(), name)
				}
			}
		}
		if defaultLayout != nil {
			for _, tmpl := range defaultLayout.Tmpl.Templates() {
				if _, err := tt.AddParseTree(tmpl.Name(), tmpl.Tree); err != nil {
					log.ErrorF("error adding %v to template: %v", tmpl.Name(), err)
				} else {
					// log.DebugF("adding %v template to %v layout", tmpl.Name(), name)
				}
			}
		}
		for _, tmpl := range layout.Tmpl.Templates() {
			if _, err := tt.AddParseTree(tmpl.Name(), tmpl.Tree); err != nil {
				log.ErrorF("error adding %v template to %v layout: %v", tmpl.Name(), name, err)
			} else {
				// log.DebugF("adding %v template to %v layout", tmpl.Name(), name)
			}
		}
	}
}