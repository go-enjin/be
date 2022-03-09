//go:build database_content || databases || all

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

	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/features/database"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net"
	"github.com/go-enjin/be/pkg/page"
)

var _databaseContent *Feature

var _ feature.Feature = (*Feature)(nil)

var _ feature.Middleware = (*Feature)(nil)

const Tag feature.Tag = "DatabaseContent"

type Feature struct {
	feature.CMiddleware

	tables  map[string]string
	mounted map[string]*page.Table
}

type MakeFeature interface {
	feature.MakeFeature

	MountTable(mount, name string) MakeFeature
}

func New() MakeFeature {
	if _databaseContent == nil {
		_databaseContent = new(Feature)
		_databaseContent.Init(_databaseContent)
	}
	return _databaseContent
}

func (f *Feature) MountTable(mount, name string) MakeFeature {
	if _, ok := f.tables[mount]; ok {
		log.FatalF("%v table mount already exists", mount)
		return nil
	}
	log.DebugF("requesting table mount: %v -> %v", name, mount)
	f.tables[mount] = strcase.ToSnake(name)
	return f
}

func (f *Feature) Init(this interface{}) {
	f.CMiddleware.Init(this)
	f.tables = make(map[string]string)
	f.mounted = make(map[string]*page.Table)
}

func (f *Feature) Depends() (tags feature.Tags) {
	tags = []feature.Tag{
		database.Tag,
	}
	return
}

func (f *Feature) Tag() (tag feature.Tag) {
	tag = Tag
	return
}

func (f *Feature) Build(_ feature.Buildable) (err error) {
	return
}

func (f *Feature) Startup(ctx *cli.Context) (err error) {
	if len(f.tables) == 0 {
		err = fmt.Errorf("database content included without any tables mounted")
		return
	}
	log.DebugF("database content startup")
	for mount, name := range f.tables {
		f.mounted[mount] = page.NewTable(mount, name)
		if err = f.mounted[mount].Migrate(); err != nil {
			return
		}
		log.DebugF("including database content table: %v, mounted to: %v", name, mount)
	}
	return
}

func (f *Feature) Use(s feature.System) feature.MiddlewareFn {
	log.DebugF("including database content middleware")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := net.TrimQueryParams(r.URL.Path)
			for mount, table := range f.mounted {
				if p, err := table.Get(path); err == nil {
					if err := s.ServePage(p, w, r); err != nil {
						log.ErrorF("error serving database content: %v", err)
					} else {
						log.DebugF("%v database content served: %v", mount, path)
					}
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}