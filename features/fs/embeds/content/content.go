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
	"sort"

	"github.com/fvbommel/sortorder"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/fs"
	beFsEmbed "github.com/go-enjin/be/pkg/fs/embed"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net"
	"github.com/go-enjin/be/pkg/page"
)

var _embedContent *Feature

var _ feature.Feature = (*Feature)(nil)

var _ feature.Middleware = (*Feature)(nil)

const Tag feature.Tag = "EmbedContent"

type Feature struct {
	feature.CMiddleware

	paths   map[string]string
	setup   map[string]embed.FS
	mounted map[string]fs.FileSystem
}

type MakeFeature interface {
	feature.MakeFeature

	MountPathFs(mount, path string, fs embed.FS) MakeFeature
}

func New() MakeFeature {
	if _embedContent == nil {
		_embedContent = new(Feature)
		_embedContent.Init(_embedContent)
	}
	return _embedContent
}

func (f *Feature) MountPathFs(mount, path string, fs embed.FS) MakeFeature {
	f.paths[mount] = path
	f.setup[mount] = fs
	return f
}

func (f *Feature) Init(this interface{}) {
	f.CMiddleware.Init(this)
	f.paths = make(map[string]string)
	f.setup = make(map[string]embed.FS)
	f.mounted = make(map[string]fs.FileSystem)
}

func (f *Feature) Tag() (tag feature.Tag) {
	tag = Tag
	return
}

func (f *Feature) listMountPoints() (mounts []string) {
	for mount, _ := range f.setup {
		mounts = append(mounts, mount)
	}
	sort.Sort(sortorder.Natural(mounts))
	return
}

func (f *Feature) Build(_ feature.Buildable) (err error) {
	for _, mount := range f.listMountPoints() {
		if _, ok := f.mounted[mount]; ok {
			err = fmt.Errorf(`"%v" already mounted`, mount)
			return
		}
		path := f.paths[mount]
		var lfs fs.FileSystem
		if lfs, err = beFsEmbed.New(path, f.setup[mount]); err != nil {
			log.FatalF(`error mounting filesystem: %v`, err)
			return nil
		}
		f.mounted[mount] = lfs
		log.DebugF("mounted embed content filesystem on %v to %v", mount, path)
	}
	return
}

func (f *Feature) Startup(ctx *cli.Context) (err error) {
	return
}

func (f *Feature) Use(s feature.System) feature.MiddlewareFn {
	mounts := f.listMountPoints()
	log.DebugF("including embed content %v middleware: %v", page.Extensions, f.setup)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := net.TrimQueryParams(r.URL.Path)
			for _, m := range mounts {
				if pages, err := fs.FindAllFilePages(f.mounted[m], m, ""); err == nil {
					for _, p := range pages {
						if p.Url == path {
							if err = s.ServePage(p, w, r); err == nil {
								log.DebugF("embed content served: %v", path)
								return
							} else {
								log.ErrorF("serve embed %v content: %v - error: %v", m, path, err)
							}
						}
					}
				}
			}
			// log.DebugF("not embed content: %v", path)
			next.ServeHTTP(w, r)
		})
	}
}