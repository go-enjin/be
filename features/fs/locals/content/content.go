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
	"strings"

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

func (f *Feature) Startup(ctx *cli.Context) (err error) {
	f.cache.Rebuild()
	return
}

func (f *Feature) Use(s feature.System) feature.MiddlewareFn {
	log.DebugF("including local content %v middleware: %v", page.Extensions, f.setup)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := net.TrimQueryParams(r.URL.Path)
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
	path = net.TrimQueryParams(path)
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		path = "/"
	}
	if mount, mpath, pg, e := f.cache.Lookup(path); e == nil {
		if err = s.ServePage(pg, w, r); err == nil {
			log.DebugF("served local %v content: %v", mount, mpath)
			return
		}
		err = fmt.Errorf("serve local %v content: %v - error: %v", mount, mpath, err)
		return
	}
	err = fmt.Errorf("path not found")
	return
}