//go:build embed_public || embeds || all

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

package public

import (
	"embed"
	"fmt"
	"net/http"
	"sort"

	"github.com/fvbommel/sortorder"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	beFs "github.com/go-enjin/be/pkg/fs"
	beFsEmbed "github.com/go-enjin/be/pkg/fs/embed"
	"github.com/go-enjin/be/pkg/log"
	bePath "github.com/go-enjin/be/pkg/path"
)

var (
	_ Feature = (*CFeature)(nil)
)

var (
	DefaultCacheControl = "max-age=604800, must-revalidate"
)

const Tag feature.Tag = "EmbedPublic"

type Feature interface {
	feature.Middleware
}

type CFeature struct {
	feature.CMiddleware

	paths   map[string]string
	setup   map[string]embed.FS
	mounted map[string]beFs.FileSystem

	cacheControl string
}

type MakeFeature interface {
	MountPathFs(mount, path string, efs embed.FS) MakeFeature
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
	f.mounted = make(map[string]beFs.FileSystem)
}

func (f *CFeature) Tag() (tag feature.Tag) {
	tag = Tag
	return
}

func (f *CFeature) listMountPoints() (mounts []string) {
	for mount, _ := range f.setup {
		mounts = append(mounts, mount)
	}
	sort.Sort(sortorder.Natural(mounts))
	return
}

func (f *CFeature) Build(_ feature.Buildable) (err error) {
	for _, mount := range f.listMountPoints() {
		if _, ok := f.mounted[mount]; ok {
			err = fmt.Errorf(`"%v" already mounted`, mount)
			return
		}
		if f.mounted[mount], err = beFsEmbed.New(f.paths[mount], f.setup[mount]); err != nil {
			log.FatalF(`error mounting filesystem: %v`, err)
			return nil
		}
		beFs.RegisterFileSystem(mount, f.mounted[mount])
		log.DebugF("mounted embed public filesystem on %v to %v", mount, f.paths[mount])
	}
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	return
}

func (f *CFeature) Use(s feature.System) feature.MiddlewareFn {
	mounts := f.listMountPoints()
	log.DebugF("including embed public middleware: %v", f.setup)
	return func(next http.Handler) (this http.Handler) {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := bePath.TrimSlash(r.URL.Path)
			path = forms.TrimQueryParams(path)
			if len(path) > 1 {
				for _, m := range mounts {
					if data, mime, filePath, ok := beFs.CheckForFileData(f.mounted[m], path, m); ok {
						if f.cacheControl == "" && DefaultCacheControl != "" {
							w.Header().Set("Cache-Control", DefaultCacheControl)
						} else if f.cacheControl != "" {
							w.Header().Set("Cache-Control", f.cacheControl)
						}
						s.ServeData(data, mime, w, r)
						log.DebugF("served embed %v public: %v (%v)", m, filePath, mime)
						return
					}
				}
			}
			// log.DebugF("not embed public: %v", path)
			next.ServeHTTP(w, r)
		})
	}
}
