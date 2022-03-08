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
	beFs "github.com/go-enjin/be/pkg/fs"
	beFsEmbed "github.com/go-enjin/be/pkg/fs/embed"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net"
	bePath "github.com/go-enjin/be/pkg/path"
)

var _embedPublic *Feature

var _ feature.Feature = (*Feature)(nil)

var _ feature.Middleware = (*Feature)(nil)

const Tag feature.Tag = "EmbedPublic"

type Feature struct {
	feature.CMiddleware

	paths   map[string]string
	setup   map[string]embed.FS
	mounted map[string]beFs.FileSystem
}

type MakeFeature interface {
	feature.MakeFeature

	MountPathFs(mount, path string, efs embed.FS) MakeFeature
}

func New() MakeFeature {
	if _embedPublic == nil {
		_embedPublic = new(Feature)
		_embedPublic.Init(_embedPublic)
	}
	return _embedPublic
}

func (f *Feature) MountPathFs(mount, path string, efs embed.FS) MakeFeature {
	f.paths[mount] = path
	f.setup[mount] = efs
	return f
}

func (f *Feature) Init(this interface{}) {
	f.CMiddleware.Init(this)
	f.paths = make(map[string]string)
	f.setup = make(map[string]embed.FS)
	f.mounted = make(map[string]beFs.FileSystem)
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
		if f.mounted[mount], err = beFsEmbed.New(f.paths[mount], f.setup[mount]); err != nil {
			log.FatalF(`error mounting filesystem: %v`, err)
			return nil
		}
		beFs.RegisterFileSystem(mount, f.mounted[mount])
		log.DebugF("mounted embed public filesystem on %v to %v", mount, f.paths[mount])
	}
	return
}

func (f *Feature) Startup(ctx *cli.Context) (err error) {
	return
}

func (f *Feature) Use(s feature.System) feature.MiddlewareFn {
	mounts := f.listMountPoints()
	log.DebugF("including embed public middleware: %v", f.setup)
	return func(next http.Handler) (this http.Handler) {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := bePath.TrimSlash(r.URL.Path)
			path = net.TrimQueryParams(path)
			if len(path) > 1 {
				for _, m := range mounts {
					if data, path, mime, ok := beFs.CheckForFileData(f.mounted[m], path, m); ok {
						s.ServeData(data, mime, w, r)
						log.DebugF("served embed %v public: %v (%v)", m, path, mime)
						return
					}
				}
			}
			log.DebugF("not embed public: %v", path)
			next.ServeHTTP(w, r)
		})
	}
}