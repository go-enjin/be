//go:build stock_pgc || pages || all

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

package pgc

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-enjin/golang-org-x-text/language"
	"github.com/gofrs/uuid"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/page"
	"github.com/go-enjin/be/pkg/pagecache"
	strings2 "github.com/go-enjin/be/pkg/strings"
	"github.com/go-enjin/be/pkg/types/theme-types"
)

// TODO: Add, Remove, Update within existing cache
// TODO: stub a map of all redirections and translations, optimize feature-fs-content.Find{Redirection,Translations}
// TODO: figure out a better way of preparing sitemaps
// TODO: replace ListAll with pre-built indexing
// TODO: make cache process query expression instead of per-page

type Mount struct {
	Point string
	Path  string
	FS    fs.FileSystem

	TotalCached uint64
}

type Cache struct {
	mount map[string]*Mount

	All          []*pagecache.Stub
	Stubs        map[string]map[language.Tag]map[string]*pagecache.Stub
	Translations map[string][]*pagecache.Stub
	Redirections map[string]*pagecache.Stub

	Formats  types.FormatProvider
	LangMode lang.Mode
	Fallback language.Tag

	TotalCached uint64

	search pagecache.SearchEnjinFeature

	enjin feature.Internals

	sync.RWMutex
}

func newCache(enjin feature.Internals, formats types.FormatProvider, langMode lang.Mode, fallback language.Tag, search pagecache.SearchEnjinFeature) (c *Cache) {
	c = new(Cache)
	c.enjin = enjin
	c.mount = make(map[string]*Mount)
	c.Stubs = make(map[string]map[language.Tag]map[string]*pagecache.Stub)
	c.Translations = make(map[string][]*pagecache.Stub)
	c.Redirections = make(map[string]*pagecache.Stub)
	c.Formats = formats
	c.LangMode = langMode
	c.Fallback = fallback
	c.search = search
	return
}

func (c *Cache) Mounted(path string) (ok bool) {
	c.RLock()
	defer c.RUnlock()
	for _, mount := range c.mount {
		if ok = mount.Path == path; ok {
			return
		}
	}
	return
}

func (c *Cache) Mount(mount, path string, mfs fs.FileSystem) {
	c.Lock()
	defer c.Unlock()
	c.mount[path] = &Mount{
		Point: mount,
		Path:  path,
		FS:    mfs,
	}
}

func (c *Cache) Rebuild() (ok bool, errs []error) {
	c.Lock()
	defer c.Unlock()
	if c.Formats == nil {
		return
	}

	var totalCached, mountCached uint64
	var batchStart time.Time
	rebuildStart := time.Now()

	updateCacheFile := func(point, mount, file, path, shasum string, tag language.Tag, bfs fs.FileSystem) {
		fileStart := time.Now()

		var err error
		var stub *pagecache.Stub
		var p *page.Page
		if stub, p, err = pagecache.NewStub(c.enjin.Context(), bfs, point, file, shasum, tag); err != nil {
			errs = append(errs, err)
			return
		}
		if p, err = stub.Make(c.Formats); err != nil {
			errs = append(errs, err)
			return
		}
		stub.Language = p.LanguageTag

		if _, ok := c.Stubs[mount][p.LanguageTag]; !ok {
			c.Stubs[mount][p.LanguageTag] = make(map[string]*pagecache.Stub)
		}
		c.All = append(c.All, stub)
		c.Stubs[mount][p.LanguageTag][file] = stub
		c.Stubs[mount][p.LanguageTag][p.Url] = stub
		if p.Permalink != uuid.Nil {
			c.Stubs[mount][p.LanguageTag]["/"+p.PermalinkSha] = stub
			c.Stubs[mount][p.LanguageTag]["/"+p.Permalink.String()] = stub
		}
		mountCached += 1

		for _, redirect := range p.Redirections() {
			c.Redirections[redirect] = stub
		}
		if p.Translates != "" {
			c.Translations[p.Translates] = append(c.Translations[p.Translates], stub)
		}

		if c.search != nil {
			// log.WarnF("adding to search index: %v", stub.Source)
			if ee := c.search.AddToSearchIndex(stub, p); ee != nil {
				ee = fmt.Errorf("error adding page to search index: %v - %v", p.Url, ee)
				log.ErrorF("%v", ee)
				errs = append(errs, ee)
				return
			}
		}

		for _, feat := range c.enjin.Features() {
			if indexer, ok := feat.(pagecache.PageIndexFeature); ok {
				// log.WarnF("adding to page index: %v - %v", feat.Tag(), stub.Source)
				if err = indexer.AddToIndex(stub, p); err != nil {
					err = fmt.Errorf("error adding to search index feature: %v", err)
					return
				}
			}
		}

		log.TraceF("cached [%v/%v] %v mount: %v (%v)", tag, p.Language, mount, path, p.Url)
		delta := time.Now().Sub(fileStart)
		var reportThreshold uint64
		if delta.Milliseconds() >= 3 {
			reportThreshold = 250
		} else if delta.Milliseconds() >= 2 {
			reportThreshold = 2500
		} else {
			reportThreshold = 25000
		}
		if mountCached > 0 && mountCached%reportThreshold == 0 {
			log.DebugF("cache %v progress %d pages (%v) (last: %v)", bfs.Name(), mountCached, time.Now().Sub(batchStart), delta)
			batchStart = time.Now()
		}
		return
	}

	updateCacheDir := func(point, mount string, tag language.Tag, bfs fs.FileSystem, ignore []string) {
		batchStart = time.Now()
		if paths, e := bfs.ListAllFiles("."); e == nil {
			// log.WarnF("hit: %v", len(paths))
			for _, file := range paths {
				if checkIgnored(file, ignore) {
					continue
				}
				if shasum, ee := bfs.Shasum(file); ee == nil {
					pgFile := strings2.TrimPrefixes(file, tag.String())
					updateCacheFile(point, mount, file, pgFile, shasum, tag, bfs)
				} else {
					errs = append(errs, fmt.Errorf("error: shasum %v mount %v - %v", mount, file, ee))
				}
			}
		} else {
			errs = append(errs, fmt.Errorf("error: list all files %v mount - %v", mount, e))
		}
	}

	// add new pages to cache
	for _, mount := range maps.SortedKeys(c.mount) {
		mfs := c.mount[mount]
		// log.WarnF("processing mount: %v - %v", mount, mfs.FS.Name())
		mountCached = 0
		mountStart := time.Now()

		if v, ok := c.Stubs[mount]; !ok || v == nil {
			c.Stubs[mount] = make(map[language.Tag]map[string]*pagecache.Stub)
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

		batchStart = time.Now()
		updateCacheDir(mfs.Point, mount, language.Und, mfs.FS, ignore)
		for _, tag := range lang.SortedLanguageTagKeys(updates) {
			bfs := updates[tag]
			// log.WarnF("updating cache directory: [%v] %v", tag.String(), bfs.Name())
			updateCacheDir(mfs.Point, mount, tag, bfs, nil)
		}

		totalCached += mountCached
		mfs.TotalCached += mountCached
		log.DebugF("cache %v updated %d pages (%v)", mfs.FS.Name(), mountCached, time.Now().Sub(mountStart))
	}

	numErrs := len(errs)
	if ok = numErrs == 0; !ok {
		if numErrs > 10 {
			log.ErrorF("errors (%d) during cache rebuilding: (too many to output)\n%v", len(errs), errs[0])
		} else {
			log.ErrorF("errors (%d) during cache rebuilding:\n%v", len(errs), errs)
		}
	}

	c.TotalCached = totalCached
	log.DebugF("cache updated %d total pages (%v)", totalCached, time.Now().Sub(rebuildStart))
	return
}

func (c *Cache) Lookup(tag language.Tag, url string) (mount, path string, p *page.Page, err error) {
	c.RLock()
	defer c.RUnlock()

	process := func(langTag language.Tag) (mount, path string, p *page.Page, ok bool) {
		for m, locales := range c.Stubs {
			var stubs map[string]*pagecache.Stub
			if stubs, ok = locales[langTag]; ok {
				var stub *pagecache.Stub
				if stub, ok = stubs[url]; ok {
					if p, err = stub.Make(c.Formats); err != nil {
						log.ErrorF("error making page from stub: %v", err)
						continue
					} else if path, ok = p.Match(url); !ok {
						continue
					}
					mount = m
					return
				}
			}
		}
		return
	}

	var ok bool
	if mount, path, p, ok = process(tag); ok {
		return
	} else if mount, path, p, ok = process(language.Und); ok {
		return
	}

	err = fmt.Errorf("page not found")
	return
}

func (c *Cache) LookupTranslations(url string) (pgs []*page.Page) {
	c.RLock()
	defer c.RUnlock()

	if found, ok := c.Translations[url]; ok {
		for _, stub := range found {
			if p, e := stub.Make(c.Formats); e == nil {
				pgs = append(pgs, p)
			} else {
				log.ErrorF("error making page from stub: %v", e)
			}
		}
	}

	return
}

func (c *Cache) LookupRedirect(url string) (p *page.Page, ok bool) {
	c.RLock()
	defer c.RUnlock()
	var e error
	var stub *pagecache.Stub
	if stub, ok = c.Redirections[url]; ok {
		p, e = stub.Make(c.Formats)
		ok = e == nil
	}
	return
}

func (c *Cache) LookupPrefix(prefix string) (found []*page.Page) {
	c.RLock()
	defer c.RUnlock()
	for _, locales := range c.Stubs {
		for _, cache := range locales {
			for url, stub := range cache {
				if strings.HasPrefix(url, prefix) {
					if pg, err := stub.Make(c.Formats); err == nil {
						found = append(found, pg)
					}
				}
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