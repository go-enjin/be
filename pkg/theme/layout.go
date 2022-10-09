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

	fileSystem beFs.FileSystem
	funcMap    template.FuncMap
	lastMods   map[string]int64
	cache      map[string]*template.Template

	sync.RWMutex
}

func NewLayout(path string, efs beFs.FileSystem, fm template.FuncMap) (layout *Layout, err error) {
	layout = new(Layout)
	layout.Path = path
	layout.Name = bePath.Base(path)
	layout.fileSystem = efs
	layout.funcMap = fm
	layout.lastMods = make(map[string]int64)
	layout.cache = make(map[string]*template.Template)
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
	l.Tmpl = tmpl.Funcs(l.funcMap)

	var walker func(p string) (e error)
	walker = func(p string) (e error) {
		var entries []string
		if entries, e = l.fileSystem.ListAllFiles(p); e != nil {
			return
		}
		for _, entry := range entries {
			var ee error

			entryBase := strings.TrimPrefix(entry, p+"/")
			entryName := l.Name + string(os.PathSeparator) + entryBase
			entryPath := entry

			var lastMod int64
			if lastMod, ee = l.fileSystem.LastModified(entryPath); ee != nil {
				log.ErrorF("error fileSystem.LastModified: %V", ee)
				continue
			} else if v, ok := l.lastMods[entryName]; ok {
				if v == lastMod {
					// log.TraceF("validated known entry: %v (%v == %v)", entryName, v, lastMod)
					if _, eee := l.Tmpl.AddParseTree(entryName, l.cache[entryName].Tree); eee != nil {
						log.ErrorF("error adding %v parse tree: %v", entryName, eee)
						delete(l.lastMods, entryName)
						delete(l.cache, entryName)
						continue
					}
					if !beStrings.StringInStrings(entryName, l.Keys...) {
						l.Keys = append(l.Keys, entryName)
					}
					continue
				}
				log.TraceF("updating known entry: %v", entryName)
				delete(l.lastMods, entryName)
				delete(l.cache, entryName)
			} else {
				// log.TraceF("caching new entry: %v (%v)", entryName, lastMod)
			}

			var data []byte
			if data, ee = l.fileSystem.ReadFile(entryPath); ee != nil {
				e = fmt.Errorf("error fileSystem.ReadFile: %v", ee)
				return
			}

			l.lastMods[entryName] = lastMod
			if l.cache[entryName], ee = l.Tmpl.New(entryName).Parse(string(data)); ee != nil {
				e = ee
				delete(l.lastMods, entryName)
				delete(l.cache, entryName)
				return
			}

			if !beStrings.StringInStrings(entryName, l.Keys...) {
				l.Keys = append(l.Keys, entryName)
			}

			// log.TraceF("layout %v entry set: %v", entryName, entryPath)
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
	for _, name := range names {
		if tmpl = l.Tmpl.Lookup(name); tmpl != nil {
			log.DebugF("using %v layout template: %v", l.Name, name)
			return
		}
	}
	return
}