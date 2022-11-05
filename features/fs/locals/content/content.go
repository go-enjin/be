//go:build local_content || locals || all

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

package content

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/blevesearch/bleve/v2"
	"github.com/fvbommel/sortorder"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/fs/local"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/page"
)

var _localContent *Feature

var (
	_ feature.Feature      = (*Feature)(nil)
	_ feature.Middleware   = (*Feature)(nil)
	_ feature.PageProvider = (*Feature)(nil)
)

const Tag feature.Tag = "LocalContent"

type Feature struct {
	feature.CMiddleware

	enjin feature.Internals

	setup map[string]string
	cache *page.Cache
}

type MakeFeature interface {
	feature.MakeFeature

	MountPath(mount, path string) MakeFeature
}

func New() MakeFeature {
	if _localContent == nil {
		_localContent = new(Feature)
		_localContent.Init(_localContent)
	}
	return _localContent
}

func (f *Feature) MountPath(mount, path string) MakeFeature {
	f.setup[mount] = path
	return f
}

func (f *Feature) Init(this interface{}) {
	f.CMiddleware.Init(this)
	f.setup = make(map[string]string)
	f.cache = page.NewCache()
}

func (f *Feature) Tag() (tag feature.Tag) {
	tag = Tag
	return
}

func (f *Feature) Build(_ feature.Buildable) (err error) {
	var mounts []string
	for mount, _ := range f.setup {
		mounts = append(mounts, mount)
	}
	sort.Sort(sortorder.Natural(mounts))
	for _, mount := range mounts {
		if f.cache.Mounted(mount) {
			err = fmt.Errorf(`"%v" already mounted`, mount)
			return
		}
		path := f.setup[mount]
		var lfs fs.FileSystem
		if lfs, err = local.New(path); err != nil {
			log.FatalF(`error mounting filesystem: %v`, err)
			return nil
		}
		f.cache.Mount(mount, f.setup[mount], lfs)
		log.DebugF("mounted local content filesystem on %v to %v", mount, path)
	}
	return
}

func (f *Feature) Setup(enjin feature.Internals) {
	f.enjin = enjin
}

func (f *Feature) Startup(ctx *cli.Context) (err error) {
	f.cache.Rebuild()
	return
}

func (f *Feature) Use(s feature.System) feature.MiddlewareFn {
	log.DebugF("including local content %v middleware: %v", page.Extensions, f.setup)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := forms.SanitizeRequestPath(r.URL.Path)
			if err := f.ServePath(path, s, w, r); err == nil {
				return
			} else if err.Error() != "path not found" {
				log.ErrorF("local content error: %v", err)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (f *Feature) ServePath(path string, s feature.System, w http.ResponseWriter, r *http.Request) (err error) {
	f.cache.Rebuild()
	reqLangTag := lang.GetTag(r)
	path = forms.SanitizeRequestPath(path)
	if mount, mpath, pg, e := f.cache.Lookup(reqLangTag, path); e == nil && pg.Context.String("type", "page") == "page" {
		if err = s.ServePage(pg, w, r); err == nil {
			log.DebugF("served local %v content: [%v] %v", mount, pg.Language, mpath)
			return
		}
		err = fmt.Errorf("serve local %v content: %v - error: %v", mount, mpath, err)
		return
	}
	err = fmt.Errorf("path not found")
	return
}

func (f *Feature) UpdateSearch(tag language.Tag, index bleve.Index) (err error) {
	f.cache.Rebuild()
	allPages := f.cache.ListAll()
	log.DebugF("locals content search updating %d documents", len(allPages))
	for _, pg := range allPages {
		if language.Compare(pg.LanguageTag, tag) {
			if doc, e := pg.SearchDocument(); e != nil {
				err = fmt.Errorf("error preparing locals search document: %v", e)
			} else if doc != nil {
				pgUrl := pg.Url
				if !language.Compare(pg.LanguageTag, f.enjin.SiteDefaultLanguage(), language.Und) {
					langMode := f.enjin.SiteLanguageMode()
					pgUrl = langMode.ToUrl(f.enjin.SiteDefaultLanguage(), pg.LanguageTag, pg.Url)
				}
				if ee := index.Index(pgUrl, doc.Self()); ee != nil {
					err = fmt.Errorf("error indexing locals search document: %v", ee)
				} else {
					log.TraceF("updated locals search index with document: %v", doc.GetUrl())
				}
			} else {
				log.TraceF("skipped locals search index with document: %v", pg.Url)
			}
		}
	}
	return
}

func (f *Feature) FindRedirection(path string) (p *page.Page) {
	f.cache.Rebuild()
	path = forms.SanitizeRequestPath(path)
	if found := f.cache.FindAll(path); len(found) > 0 {
		for _, pg := range found {
			if pg.IsRedirection(path) {
				p = pg
				return
			}
		}
	}
	return
}

func (f *Feature) FindTranslations(path string) (found []*page.Page) {
	f.cache.Rebuild()
	path = forms.SanitizeRequestPath(path)
	found = f.cache.FindAll(path)
	return
}

func (f *Feature) FindPage(tag language.Tag, path string) (p *page.Page) {
	f.cache.Rebuild()
	path = forms.SanitizeRequestPath(path)
	if _, _, pg, e := f.cache.Lookup(tag, path); e == nil {
		p = pg.Copy()
	}
	return
}

func (f *Feature) FindPages(path string) (pages []*page.Page) {
	f.cache.Rebuild()
	path = forms.SanitizeRequestPath(path)
	pages = f.cache.FindAllPrefix(path)
	return
}