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
	"github.com/go-enjin/be/pkg/maps"
	bePath "github.com/go-enjin/be/pkg/path"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

type Layout struct {
	Path string
	Name string
	Keys []string

	fileSystem beFs.FileSystem
	funcMap    template.FuncMap
	lastMods   map[string]int64
	cache      map[string]string

	sync.RWMutex
}

func NewLayout(path string, efs beFs.FileSystem, fm template.FuncMap) (layout *Layout, err error) {
	layout = new(Layout)
	layout.Path = path
	layout.Name = bePath.Base(path)
	layout.fileSystem = efs
	layout.funcMap = fm
	layout.lastMods = make(map[string]int64)
	layout.cache = make(map[string]string)
	err = layout.Reload()
	return
}

func (l *Layout) Reload() (err error) {
	l.Lock()
	defer l.Unlock()

	l.Keys = make([]string, 0)

	var walker func(p string) (e error)
	walker = func(p string) (e error) {
		var entries []string
		if entries, e = l.fileSystem.ListAllFiles(p); e != nil {
			if strings.Contains(e.Error(), "no such file or directory") {
				e = nil
			}
			return
		}

		for _, entry := range entries {
			entryBase := strings.TrimPrefix(entry, p+"/")
			entryName := l.Name + string(os.PathSeparator) + entryBase
			entryPath := entry

			var ee error
			var lastMod int64
			if lastMod, ee = l.fileSystem.LastModified(entryPath); ee != nil {
				log.ErrorF("error fileSystem.LastModified: %V", ee)
				continue
			} else if v, ok := l.lastMods[entryName]; ok {
				if v == lastMod {
					if !beStrings.StringInStrings(entryName, l.Keys...) {
						l.Keys = append(l.Keys, entryName)
					}
					continue
				}
				log.TraceF("updating known entry: %v", entryName)
				delete(l.lastMods, entryName)
				delete(l.cache, entryName)
			}

			var data []byte
			if data, ee = l.fileSystem.ReadFile(entryPath); ee != nil {
				e = fmt.Errorf("error fileSystem.ReadFile: %v", ee)
				return
			}

			l.lastMods[entryName] = lastMod
			l.cache[entryName] = string(data)

			if !beStrings.StringInStrings(entryName, l.Keys...) {
				l.Keys = append(l.Keys, entryName)
			}

			log.TraceF("cached %v layout %v data: %v", l.Name, entryName, entryPath)
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

func (l *Layout) NewTemplate() (tmpl *template.Template, err error) {
	if tmpl, err = template.New(l.Name).Parse(`{{/* empty */}}`); err == nil {
		err = l.Apply(tmpl)
	}
	return
}

func (l *Layout) Apply(tt *template.Template) (err error) {
	tt.Funcs(l.funcMap)
	for _, name := range maps.SortedKeys(l.cache) {
		if _, err = tt.New(name).Parse(l.cache[name]); err != nil {
			err = fmt.Errorf("error parsing cached template: %v - %v", name, err)
			return
		} else {
			// log.TraceF("parsed %v into %v", name, tt.Name())
		}
	}
	return
}