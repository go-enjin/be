//go:build embed_content || embeds || all

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
	"embed"
	"fmt"
	"net/http"

	"github.com/blevesearch/bleve/v2"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/fs"
	beFsEmbed "github.com/go-enjin/be/pkg/fs/embed"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/page"
)

var (
	_ Feature = (*CFeature)(nil)
)

var (
	DefaultCacheControl = "max-age=604800, must-revalidate"
)

const Tag feature.Tag = "EmbedContent"

type Feature interface {
	feature.Middleware
	feature.PageProvider
}

type CFeature struct {
	feature.CMiddleware

	enjin feature.Internals

	paths map[string]string
	setup map[string]embed.FS
	cache *page.Cache

	cacheControl string
}

type MakeFeature interface {
	MountPathFs(mount, path string, fs embed.FS) MakeFeature
	SetCacheControl(values string) MakeFeature

	Make() Feature
}

func New() MakeFeature {
	f := new(CFeature)
	f.Init(f)
	return f
}

func (f *CFeature) MountPathFs(mount, path string, efs embed.FS) MakeFeature {
	f.paths[mount] = path
	f.setup[mount] = efs
	return f
}

func (f *CFeature) SetCacheControl(values string) MakeFeature {
	f.cacheControl = values
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CMiddleware.Init(this)
	f.paths = make(map[string]string)
	f.setup = make(map[string]embed.FS)
}

func (f *CFeature) Tag() (tag feature.Tag) {
	tag = Tag
	return
}

func (f *CFeature) Build(_ feature.Buildable) (err error) {
	return
}

func (f *CFeature) Setup(enjin feature.Internals) {
	f.enjin = enjin
	t, _ := f.enjin.GetTheme()
	f.cache = page.NewCache(t)

	var err error
	for _, mount := range maps.SortedKeys(f.setup) {
		if f.cache.Mounted(mount) {
			log.FatalF(`"%v" already mounted`, mount)
			return
		}
		var lfs fs.FileSystem
		if lfs, err = beFsEmbed.New(f.paths[mount], f.setup[mount]); err != nil {
			log.FatalF(`error mounting filesystem: %v`, err)
			return
		}
		f.cache.Mount(mount, f.paths[mount], lfs)
		log.DebugF("mounted embed content filesystem on %v to %v", mount, f.paths[mount])
	}
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	f.cache.Rebuild()
	return
}

func (f *CFeature) Use(s feature.System) feature.MiddlewareFn {
	log.DebugF("including embed content middleware: %v", f.setup)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := forms.SanitizeRequestPath(r.URL.Path)
			if err := f.ServePath(path, s, w, r); err == nil {
				return
			} else if err.Error() != "path not found" {
				log.ErrorF("embed content error: %v", err)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (f *CFeature) ServePath(path string, s feature.System, w http.ResponseWriter, r *http.Request) (err error) {
	reqLangTag := lang.GetTag(r)
	path = forms.SanitizeRequestPath(path)
	if mount, mpath, pg, e := f.cache.Lookup(reqLangTag, path); e == nil {
		var cacheControl string
		if f.cacheControl == "" {
			cacheControl = DefaultCacheControl
		} else {
			cacheControl = f.cacheControl
		}
		cacheControl = pg.Context.String("CacheControl", cacheControl)
		pg.Context.SetSpecific("CacheControl", cacheControl)
		if err = s.ServePage(pg, w, r); err == nil {
			log.DebugF("served embed %v content: [%v] %v", mount, pg.Language, mpath)
			return
		}
		err = fmt.Errorf("serve embed %v content: %v - error: %v", mount, mpath, err)
		return
	}
	err = fmt.Errorf("path not found")
	return
}

func (f *CFeature) UpdateSearch(tag language.Tag, index bleve.Index) (err error) {
	f.cache.Rebuild()
	allPages := f.cache.ListAll()
	log.DebugF("embeds content search updating %d documents", len(allPages))
	for _, pg := range allPages {
		if language.Compare(pg.LanguageTag, tag) {
			if doc, e := pg.SearchDocument(); e != nil {
				err = fmt.Errorf("error preparing embeds search document: %v", e)
			} else if doc != nil {
				pgUrl := pg.Url
				if !language.Compare(pg.LanguageTag, f.enjin.SiteDefaultLanguage(), language.Und) {
					langMode := f.enjin.SiteLanguageMode()
					pgUrl = langMode.ToUrl(f.enjin.SiteDefaultLanguage(), pg.LanguageTag, pg.Url)
				}
				if ee := index.Index(pgUrl, doc.Self()); ee != nil {
					err = fmt.Errorf("error indexing embeds search document: %v", ee)
				} else {
					log.TraceF("updated embeds search index with document: %v", doc.GetUrl())
				}
			} else {
				log.TraceF("skipped embeds search index with document: %v", pg.Url)
			}
		}
	}
	return
}

func (f *CFeature) FindRedirection(path string) (p *page.Page) {
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

func (f *CFeature) FindTranslations(path string) (found []*page.Page) {
	path = forms.SanitizeRequestPath(path)
	found = f.cache.FindAll(path)
	return
}

func (f *CFeature) FindPage(tag language.Tag, path string) (p *page.Page) {
	path = forms.SanitizeRequestPath(path)
	if _, _, pg, e := f.cache.Lookup(tag, path); e == nil {
		p = pg
	}
	return
}

func (f *CFeature) FindPages(path string) (pages []*page.Page) {
	path = forms.SanitizeRequestPath(path)
	pages = f.cache.FindAllPrefix(path)
	return
}

func (f *CFeature) MatchQL(query string) (pages []*page.Page) {
	pages = f.cache.MatchQL(query)
	return
}
