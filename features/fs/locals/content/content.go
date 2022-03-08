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

	"github.com/fvbommel/sortorder"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/fs/local"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net"
	"github.com/go-enjin/be/pkg/page"
)

var _localContent *Feature

var _ feature.Feature = (*Feature)(nil)

var _ feature.Middleware = (*Feature)(nil)

const Tag feature.Tag = "LocalContent"

type Feature struct {
	feature.CMiddleware

	setup   map[string]string
	mounted map[string]fs.FileSystem
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
		path := f.setup[mount]
		var lfs fs.FileSystem
		if lfs, err = local.New(path); err != nil {
			log.FatalF(`error mounting filesystem: %v`, err)
			return nil
		}
		f.mounted[mount] = lfs
		log.DebugF("mounted local content filesystem on %v to %v", mount, path)
	}
	return
}

func (f *Feature) Startup(ctx *cli.Context) (err error) {
	return
}

func (f *Feature) Use(s feature.System) feature.MiddlewareFn {
	mounts := f.listMountPoints()
	log.DebugF("including local content %v middleware: %v", page.Extensions, f.setup)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := net.TrimQueryParams(r.URL.Path)
			for _, m := range mounts {
				if pages, err := fs.FindAllFilePages(f.mounted[m], m, ""); err == nil {
					for _, p := range pages {
						log.DebugF("checking page: %v (%v)", p.Url, path)
						if p.Url == path {
							if err = s.ServePage(p, w, r); err == nil {
								log.DebugF("local content served: %v", path)
								return
							} else {
								log.ErrorF("serve local %v content: %v - error: %v", m, path, err)
							}
						}
					}
				}
			}
			log.DebugF("not local content: %v", path)
			next.ServeHTTP(w, r)
		})
	}
}