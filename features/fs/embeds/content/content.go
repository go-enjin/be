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
	"runtime"
	"sort"

	"github.com/fvbommel/sortorder"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/features/defaults/pgc"
	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/fs"
	beFsEmbed "github.com/go-enjin/be/pkg/fs/embed"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/page"
	"github.com/go-enjin/be/pkg/pagecache"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

var (
	DefaultCacheControl = "max-age=604800, must-revalidate"
)

const (
	Tag    feature.Tag = "EmbedContent"
	Bucket string      = "embed-content"
)

type Feature interface {
	feature.Middleware
	feature.PageProvider
}

type CFeature struct {
	feature.CMiddleware

	enjin feature.Internals

	paths map[string]string
	setup map[string]embed.FS
	cache pagecache.CacheEnjinFeature

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
	f.paths[path] = mount
	f.setup[path] = efs
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
	for _, feat := range f.enjin.Features() {
		if cef, ok := feat.(pagecache.CacheEnjinFeature); ok {
			f.cache = cef
		}
	}
	if f.cache == nil {
		log.FatalF("enjin is missing a pagecache.CacheEnjinFeature")
	} else {
		if err := f.cache.NewCache(Bucket); err != nil {
			log.FatalF("error creating new cache bucket: %v - %v", Bucket, err)
		}
	}

	var err error
	for _, path := range maps.SortedKeys(f.paths) {
		if f.cache.Mounted(Bucket, path) {
			log.FatalF(`"%v" already mounted`, path)
			return
		}
		var lfs fs.FileSystem
		if lfs, err = beFsEmbed.New(path, f.setup[path]); err != nil {
			log.FatalF(`error mounting filesystem: %v`, err)
			return
		}
		mount := f.paths[path]
		f.cache.Mount(Bucket, mount, path, lfs)
		log.DebugF("mounted embed content filesystem %v to %v", path, mount)
	}
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	f.cache.Rebuild()
	runtime.GC()
	return
}

func (f *CFeature) FilterPageContext(themeCtx context.Context, pageCtx context.Context, r *http.Request) (out context.Context) {
	out = themeCtx
	totalCached := out.Uint64("SiteTotalPages", 0)
	totalCached += f.cache.TotalCached(Bucket)
	out.SetSpecific("SiteTotalPages", totalCached)
	return
}

func (f *CFeature) Use(s feature.System) feature.MiddlewareFn {
	mounts := f.listMountPaths()
	log.DebugF("including embed content middleware: %v", mounts)
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
	if mount, mpath, pg, e := f.cache.Lookup(Bucket, reqLangTag, path); e == nil {
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

func (f *CFeature) FindRedirection(path string) (p *page.Page) {
	p, _ = f.cache.LookupRedirect(Bucket, path)
	return
}

func (f *CFeature) FindTranslations(path string) (found []*page.Page) {
	found = f.cache.LookupTranslations(Bucket, path)
	return
}

func (f *CFeature) FindPage(tag language.Tag, path string) (p *page.Page) {
	if _, _, pg, e := f.cache.Lookup(Bucket, tag, path); e == nil {
		p = pg
	}
	return
}

func (f *CFeature) LookupPrefixed(path string) (pages []*page.Page) {
	pages = f.cache.LookupPrefix(Bucket, path)
	return
}

func (f *CFeature) listMountPaths() (paths []string) {
	for path, _ := range f.paths {
		paths = append(paths, path)
	}
	sort.Sort(sortorder.Natural(paths))
	return
}
