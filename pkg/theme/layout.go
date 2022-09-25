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

	beFs "github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/log"
	bePath "github.com/go-enjin/be/pkg/path"
)

type Layout struct {
	Path string
	Name string
	Keys []string
	Tmpl *template.Template
}

func NewLayout(path string, efs beFs.FileSystem, fm template.FuncMap) (layout *Layout, err error) {
	var tmpl *template.Template
	if tmpl, err = template.New("empty").Parse(`{{/* empty */}}`); err != nil {
		return
	}

	layout = new(Layout)
	layout.Path = path
	layout.Name = bePath.Base(path)
	layout.Keys = make([]string, 0)
	layout.Tmpl = tmpl

	var walker func(p string) (e error)
	walker = func(p string) (e error) {
		var entries []string
		if entries, e = efs.ListAllFiles(p); e != nil {
			return
		}
		for _, entry := range entries {
			entryBase := strings.TrimPrefix(entry, p+"/")
			entryName := layout.Name + string(os.PathSeparator) + entryBase
			entryPath := entry
			var data []byte
			if data, e = efs.ReadFile(entryPath); e != nil {
				log.ErrorF("error efs.ReadFile: %v", e)
				return
			}
			if _, e = layout.Tmpl.New(entryName).Funcs(fm).Parse(string(data)); e != nil {
				log.ErrorF("template initial parse error: %v", e)
				return
			}
			layout.Keys = append(layout.Keys, entryName)
			log.DebugF("including layout %v, from file: %v", entryName, entryPath)
		}
		return
	}

	if err = walker(path); err != nil {
		layout = nil
	}
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