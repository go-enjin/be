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
	"os"
	"strings"
	"sync"

	"github.com/go-enjin/be/pkg/feature"
	beFs "github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	bePath "github.com/go-enjin/be/pkg/path"
	"github.com/go-enjin/be/pkg/slices"
)

var (
	_ feature.ThemeLayout = (*Layout)(nil)
)

type Layout struct {
	path string
	name string
	keys []string

	fileSystem beFs.FileSystem
	lastMods   map[string]int64
	cache      map[string]string

	sync.RWMutex
}

func NewLayout(path string, efs beFs.FileSystem) (layout feature.ThemeLayout, err error) {
	l := new(Layout)
	l.path = path
	l.name = bePath.Base(path)
	l.fileSystem = efs
	l.lastMods = make(map[string]int64)
	l.cache = make(map[string]string)
	if err = l.load(); err == nil {
		layout = l
	}
	return
}

func (l *Layout) load() (err error) {
	l.Lock()
	defer l.Unlock()

	l.keys = make([]string, 0)

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
			entryName := l.name + string(os.PathSeparator) + entryBase
			entryPath := entry

			var ee error
			var lastMod int64
			if lastMod, ee = l.fileSystem.LastModified(entryPath); ee != nil {
				log.ErrorF("error fileSystem.LastModified: %v", ee)
				continue
			} else if v, ok := l.lastMods[entryName]; ok {
				if v == lastMod {
					if !slices.Present(entryName, l.keys...) {
						l.keys = append(l.keys, entryName)
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
			l.cache[entryName] = lang.PruneTranslatorComments(string(data))

			if !slices.Present(entryName, l.keys...) {
				l.keys = append(l.keys, entryName)
			}

			log.TraceF("cached %v layout %v data: %v", l.name, entryName, entryPath)
		}
		return
	}

	err = walker(l.path)
	return
}

func (l *Layout) Name() (name string) {
	name = l.name
	return
}