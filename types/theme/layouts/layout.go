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
	"github.com/go-enjin/be/pkg/maps"
	bePath "github.com/go-enjin/be/pkg/path"
)

var (
	_ feature.ThemeLayout = (*Layout)(nil)
)

type Layout struct {
	path string
	name string

	fileSystem beFs.FileSystem
	cache      map[string]string

	sync.RWMutex
}

func NewLayout(path string, efs beFs.FileSystem) (layout feature.ThemeLayout, err error) {
	l := new(Layout)
	l.path = path
	l.name = bePath.Base(path)
	l.fileSystem = efs
	l.cache = make(map[string]string)
	if err = l.load(); err == nil {
		layout = l
	}
	return
}

func (l *Layout) load() (err error) {
	l.Lock()
	defer l.Unlock()

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
			var data []byte
			if data, ee = l.fileSystem.ReadFile(entryPath); ee != nil {
				e = fmt.Errorf("error fileSystem.ReadFile: %v", ee)
				return
			}

			//l.lastMods[entryName] = lastMod
			l.cache[entryName] = lang.PruneTranslatorComments(string(data))

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

func (l *Layout) Apply(other feature.ThemeLayout) {
	for _, name := range other.CacheKeys() {
		l.cache[name] = other.CacheValue(name)
	}
}

func (l *Layout) CacheKeys() (keys []string) {
	keys = maps.SortedKeys(l.cache)
	return
}

func (l *Layout) CacheValue(key string) (value string) {
	value, _ = l.cache[key]
	return
}