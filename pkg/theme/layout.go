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
	"html/template"
	"os"
	"strings"
	"sync"

	beFs "github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/log"
	bePath "github.com/go-enjin/be/pkg/path"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

type Layout struct {
	Path string
	Name string
	Keys []string
	Tmpl *template.Template

	efs beFs.FileSystem
	fm  template.FuncMap
	// s map[string]string

	sync.RWMutex
}

func NewLayout(path string, efs beFs.FileSystem, fm template.FuncMap) (layout *Layout, err error) {
	layout = new(Layout)
	layout.Path = path
	layout.Name = bePath.Base(path)
	layout.efs = efs
	layout.fm = fm
	err = layout.Reload()
	return
}

func (l *Layout) Reload() (err error) {
	l.Lock()
	defer l.Unlock()

	var tmpl *template.Template
	if tmpl, err = template.New("empty").Parse(`{{/* empty */}}`); err != nil {
		return
	}
	l.Keys = make([]string, 0)
	l.Tmpl = tmpl

	var walker func(p string) (e error)
	walker = func(p string) (e error) {
		var entries []string
		if entries, e = l.efs.ListAllFiles(p); e != nil {
			return
		}
		for _, entry := range entries {
			var ee error

			entryBase := strings.TrimPrefix(entry, p+"/")
			entryName := l.Name + string(os.PathSeparator) + entryBase
			entryPath := entry

			// var shasum string
			// if shasum, ee = l.efs.Shasum(entryPath); ee != nil {
			// 	log.ErrorF("error efs.Shasum: %v", ee)
			// 	continue
			// }
			// if v, ok := l.s[entryName]; ok {
			// 	if v == shasum {
			// 		log.TraceF("validated known entry: %v", entryName)
			// 		continue
			// 	} else {
			// 		log.TraceF("overwriting known entry: %v", entryName)
			// 	}
			// }
			// l.s[entryName] = shasum

			var data []byte
			if data, ee = l.efs.ReadFile(entryPath); ee != nil {
				log.ErrorF("error efs.ReadFile: %v", ee)
				continue
			}

			if _, ee = l.Tmpl.New(entryName).Funcs(l.fm).Parse(string(data)); ee != nil {
				log.ErrorF("template initial parse error: %v", ee)
				continue
			}

			if !beStrings.StringInStrings(entryName, l.Keys...) {
				l.Keys = append(l.Keys, entryName)
			}
			log.TraceF("layout %v entry set: %v", entryName, entryPath)
		}
		return
	}

	err = walker(l.Path)
	return
}

func (l *Layout) HasKey(key string) bool {
	for _, k := range l.Keys {
		if k == key {
			return true
		}
	}
	return false
}

func (l *Layout) Lookup(names ...string) (tmpl *template.Template) {
	// log.DebugF("checking %v layout%v", l.Name, l.Tmpl.DefinedTemplates())
	for _, name := range names {
		if tmpl = l.Tmpl.Lookup(name); tmpl != nil {
			log.DebugF("using %v layout template: %v", l.Name, name)
			return
			// } else {
			// 	log.DebugF("%v layout template not found: %v", l.Name, name)
		}
	}
	return
}