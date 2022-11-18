//go:build local_public || locals || all

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
	"fmt"
	"net/http"
	"sort"

	"github.com/fvbommel/sortorder"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	beFs "github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/fs/local"
	"github.com/go-enjin/be/pkg/log"
	bePath "github.com/go-enjin/be/pkg/path"
)

var (
	_ Feature = (*CFeature)(nil)
)

var (
	DefaultCacheControl = "max-age=604800, must-revalidate"
)

const Tag feature.Tag = "LocalPublic"

type Feature interface {
	feature.Middleware
}

type CFeature struct {
	feature.CMiddleware

	setup   map[string]string
	mounted map[string]beFs.FileSystem

	cacheControl string
}

type MakeFeature interface {
	MountPath(mount, path string) MakeFeature
	SetCacheControl(values string) MakeFeature

	Make() Feature
}

func New() MakeFeature {
	f := new(CFeature)
	f.Init(f)
	return f
}

func (f *CFeature) MountPath(mount, path string) MakeFeature {
	f.setup[mount] = path
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
	f.setup = make(map[string]string)
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
		path := f.setup[mount]
		var lfs beFs.FileSystem
		if lfs, err = local.New(path); err != nil {
			log.FatalF(`error mounting filesystem: %v`, err)
			return nil
		}
		f.mounted[mount] = lfs
		beFs.RegisterFileSystem(mount, f.mounted[mount])
		log.DebugF("mounted local public filesystem on %v to %v", mount, path)
	}
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	return
}

func (f *CFeature) Use(s feature.System) feature.MiddlewareFn {
	mounts := f.listMountPoints()
	log.DebugF("including local public middleware: %v", f.setup)
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
						log.DebugF("served local %v public: %v (%v)", m, filePath, mime)
						return
					}
				}
			}
			// log.DebugF("not local public: %v", path)
			next.ServeHTTP(w, r)
		})
	}
}
