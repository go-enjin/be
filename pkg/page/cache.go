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

package page

import (
	"fmt"
	"strings"
	"sync"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/log"
)

type Mount struct {
	Point string
	Path  string
	FS    fs.FileSystem
}

type Cache struct {
	mount map[string]*Mount
	cache map[string]map[language.Tag]map[string]*Page

	sync.RWMutex
}

func NewCache() (c *Cache) {
	c = new(Cache)
	c.mount = make(map[string]*Mount)
	c.cache = make(map[string]map[language.Tag]map[string]*Page)
	return
}

func (c *Cache) Mounted(mount string) (ok bool) {
	c.RLock()
	defer c.RUnlock()
	_, ok = c.mount[mount]
	return
}

func (c *Cache) Mount(mount, path string, mfs fs.FileSystem) {
	c.Lock()
	defer c.Unlock()
	c.mount[mount] = &Mount{
		Point: mount,
		Path:  path,
		FS:    mfs,
	}
}

func (c *Cache) Lookup(tag language.Tag, url string) (mount, path string, p *Page, err error) {
	c.RLock()
	defer c.RUnlock()

	process := func(localeCache map[string]*Page) (file string, pg *Page) {
		for _, lp := range localeCache {
			if m, ok := lp.Match(url); ok {
				file = m
				pg = lp.Copy()
				return
			}
		}
		return
	}

	// check for tag first
	for m, cache := range c.cache {
		if lc, ok := cache[tag]; ok {
			if f, pg := process(lc); pg != nil {
				p = pg.Copy()
				path = f
				mount = m
				return
			}
		}
	}

	// fallback to Und
	for m, cache := range c.cache {
		if lc, ok := cache[language.Und]; ok {
			if f, pg := process(lc); pg != nil {
				p = pg.Copy()
				path = f
				mount = m
				return
			}
		}
	}

	err = fmt.Errorf("page not found")
	return
}

func (c *Cache) FindAll(path string) (found []*Page) {
	c.RLock()
	defer c.RUnlock()
	for _, cache := range c.cache {
		for _, localeCache := range cache {
			for _, pg := range localeCache {
				if _, ok := pg.Match(path); ok {
					found = append(found, pg.Copy())
				}
			}
		}
	}
	return
}

func (c *Cache) ListAll() (found []*Page) {
	c.RLock()
	defer c.RUnlock()
	for _, cache := range c.cache {
		for _, localeCache := range cache {
			for _, pg := range localeCache {
				found = append(found, pg.Copy())
			}
		}
	}
	return
}

func checkIgnored(file string, ignore []string) (ok bool) {
	for _, ignored := range ignore {
		if ok = strings.HasPrefix(file, ignored+"/"); ok {
			return
		}
	}
	return
}

func trimPrefixes(value string, prefixes ...string) (trimmed string) {
	trimmed = value
	for _, prefix := range prefixes {
		trimmed = strings.TrimPrefix(trimmed, prefix)
		if len(trimmed) > 0 && trimmed[0] == '/' {
			trimmed = trimmed[1:]
		}
		if trimmed != value {
			return
		}
	}
	return
}

func (c *Cache) Rebuild() (ok bool, errs []error) {
	c.Lock()
	defer c.Unlock()

	for mount, cache := range c.cache {
		for tag, pages := range cache {
			for file, pg := range pages {
				if shasum, err := c.mount[mount].FS.Shasum(file); err == nil {
					if pgShasum := pg.Context.String("Shasum", shasum); pgShasum != shasum {
						delete(c.cache[mount][tag], file)
						log.DebugF("cache reset: %v - %v != %v", pg.Url, pgShasum, shasum)
					} else {
						log.TraceF("cache valid: %v", pg.Url)
					}
				} else {
					delete(c.cache[mount][tag], file)
					log.DebugF("cache clear: %v (%v)", file, err)
				}
			}
		}
	}

	updateCacheFile := func(mount, file, path, shasum string, tag language.Tag, bfs fs.FileSystem) {
		if data, eee := bfs.ReadFile(file); eee == nil {
			path = trimPrefixes(path, tag.String())
			// log.DebugF("caching path: %v, %v, %v", file, path, tag.String())
			var created, updated int64

			if epoch, err := bfs.FileCreated(file); err == nil {
				created = epoch
				// log.ErrorF("setting created: %v", epoch)
			} else {
				log.ErrorF("error getting page created: %v", err)
			}

			if epoch, err := bfs.LastModified(file); err == nil {
				updated = epoch
			} else {
				log.ErrorF("error getting page last modified: %v", err)
			}

			if p, eeee := New(path, string(data), created, updated); eeee == nil {

				p.Context.Set("Shasum", shasum)
				if language.Compare(p.LanguageTag, language.Und) {
					p.SetLanguage(tag)
				}

				if _, ok := c.cache[mount][p.LanguageTag]; !ok {
					c.cache[mount][p.LanguageTag] = make(map[string]*Page)
				}
				if _, ok := c.cache[mount][p.LanguageTag][file]; ok {
					// log.DebugF("cache exists after pruning: %v", path)
					return
				}

				c.cache[mount][p.LanguageTag][file] = p
				log.DebugF("cached [%v/%v] %v mount: %v (%v)", tag, p.Language, mount, path, p.Url)
			} else {
				errs = append(errs, fmt.Errorf("error: new %v mount page %v - %v", mount, path, eeee))
			}
		} else {
			errs = append(errs, fmt.Errorf("error: read %v mount file - %v", mount, eee))
		}
	}

	updateCacheDir := func(mount string, tag language.Tag, bfs fs.FileSystem, ignore []string) {
		if paths, e := bfs.ListAllFiles("."); e == nil {
			for _, file := range paths {
				if checkIgnored(file, ignore) {
					continue
				}
				if shasum, ee := bfs.Shasum(file); ee == nil {
					pgFile := trimPrefixes(file, tag.String())
					updateCacheFile(mount, file, pgFile, shasum, tag, bfs)
				} else {
					errs = append(errs, fmt.Errorf("error: shasum %v mount %v - %v", mount, file, ee))
				}
			}
		} else {
			errs = append(errs, fmt.Errorf("error: list all files %v mount - %v", mount, e))
		}
	}

	// add new pages to cache
	for mount, mfs := range c.mount {
		// log.WarnF("processing mount: %v - %v", mount, mfs.FS.Name())

		if v, ok := c.cache[mount]; !ok || v == nil {
			c.cache[mount] = make(map[language.Tag]map[string]*Page)
		}

		var ignore []string
		updates := make(map[language.Tag]fs.FileSystem)

		if dirs, e := mfs.FS.ListDirs("."); e == nil {
			for _, dir := range dirs {
				if dt, ee := language.Parse(dir); ee == nil {
					ignore = append(ignore, dir)
					if bfs, eee := fs.Wrap(dir, mfs.FS); eee == nil {
						// log.DebugF("wrapped locale dir: %v - %v", dir, bfs.Name())
						updates[dt] = bfs
					} else {
						// 	log.ErrorF("error wrapping locale dir: %v - %v", dir, eee)
					}
				}
			}
		}

		updateCacheDir(mount, language.Und, mfs.FS, ignore)
		for tag, bfs := range updates {
			// log.WarnF("updating cache directory: [%v] %v", tag.String(), bfs.Name())
			updateCacheDir(mount, tag, bfs, nil)
		}
	}

	if ok = len(errs) == 0; !ok {
		log.ErrorF("errors (%d) during cache rebuilding: %v", len(errs), errs)
	}
	return
}