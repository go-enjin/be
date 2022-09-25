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
	cache map[string]map[string]*Page
}

func NewCache() (c *Cache) {
	c = new(Cache)
	c.mount = make(map[string]*Mount)
	c.cache = make(map[string]map[string]*Page)
	return
}

func (c *Cache) Mounted(mount string) (ok bool) {
	_, ok = c.mount[mount]
	return
}

func (c *Cache) Mount(mount, path string, mfs fs.FileSystem) {
	c.mount[mount] = &Mount{
		Point: mount,
		Path:  path,
		FS:    mfs,
	}
}

func (c *Cache) Lookup(url string) (mount, path string, p *Page, err error) {
	for m, cache := range c.cache {
		for mp, pg := range cache {
			if pg.Url == url {
				p = pg
				mount = m
				path = mp
				// log.DebugF("FOUND m=%v, mp=%v, pg=%v, url=%v", m, mp, pg.Url, url)
				return
			}
			// log.DebugF("SKIP m=%v, mp=%v, pg=%v, url=%v", m, mp, pg, url)
		}
	}
	err = fmt.Errorf("page not found")
	return
}

func (c *Cache) Rebuild() (ok bool, errs []error) {
	// prune existing cache
	for mount, mc := range c.cache {
		for path, _ := range mc {
			if _, e := c.mount[mount].FS.Shasum(path); e != nil {
				// remove page from cache
				delete(c.cache[mount], path)
				log.DebugF("removing page from %v cache: %v", mount, path)
			}
		}
	}

	updateCache := func(mount, path, shasum string) {
		if data, eee := c.mount[mount].FS.ReadFile(path); eee == nil {
			slug := strings.TrimPrefix(path, c.mount[mount].Path)
			if p, eeee := NewFromString(slug, string(data)); eeee == nil {
				p.Context.Set("Shasum", shasum)
				c.cache[mount][path] = p
				log.DebugF("updated %v mount path: %v (%v)", mount, path, p.Url)
			} else {
				errs = append(errs, fmt.Errorf("error: new %v mount page %v - %v", mount, path, eeee))
			}
		} else {
			errs = append(errs, fmt.Errorf("error: read %v mount file - %v", mount, eee))
		}
	}

	// add new pages to cache
	for mount, mfs := range c.mount {
		if v, ok := c.cache[mount]; !ok || v == nil {
			c.cache[mount] = make(map[string]*Page)
		}

		if paths, e := mfs.FS.ListAllFiles(""); e == nil {
			for _, path := range paths {

				if shasum, ee := mfs.FS.Shasum(path); ee == nil {

					if pg, ok := c.cache[mount][path]; ok {
						if pg.Context.String("Shasum", "") != shasum {
							// update
							updateCache(mount, path, shasum)
						} else {
							// valid cache
							log.TraceF("validated %v mount path: %v (%v)", mount, path, pg.Url)
						}
					} else {
						// new
						updateCache(mount, path, shasum)
					}

				} else {
					errs = append(errs, fmt.Errorf("error: shasum %v mount %v - %v", mount, path, ee))
				}
			}
		} else {
			errs = append(errs, fmt.Errorf("error: list all files %v mount - %v", mount, e))
		}
	}

	if ok = len(errs) == 0; !ok {
		log.DebugF("errors (%d) during cache rebuilding: %v", len(errs), errs)
	}
	return
}