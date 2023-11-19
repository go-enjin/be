//go:build driver_fs_db_gorm || drivers_fs_db || drivers_fs || dbs || all

// Copyright (c) 2023  The Go-Enjin Authors
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

package filesystem

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"gorm.io/gorm"

	beFsGormDB "github.com/go-enjin/be/drivers/fs/db/gorm"
	"github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/log"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

type GormDBPathSupport[MakeTypedFeature interface{}] interface {
	// MountGormDBPath maps the database filesystem `path` to the enjin URL
	// `point`
	//
	// The `point` is pruned from the URL during an HTTP request and the `path`
	// prefixes the file's real path within the database. For example, it's
	// common for the following pattern:
	//
	//   f.MountGormDBPath("/", "public", "default")
	//
	// This configuration means to provide everything within the database path
	// of `./public/*` (recursively) at the root point of the URL, so for
	// example the URL `/favicon.ico` would translate to the local filesystem
	// path of `./public/favicon.ico`
	//
	// On enjin Startup the mounted gorm DB settings will create a table (if one
	// does not exist already) named: `<tag>_<point>`. For example, if the
	// feature is tagged with "fs-content" the mount point "/" results in
	// `fs_content`, whereas having a mount point of "/stuff" results in a table
	// name of  `fs_content_stuff`
	//
	// The given `path` prefixes all filenames within the database, which is
	// useful when using a single filesystem table for all content
	MountGormDBPath(point, path string, connection string) MakeTypedFeature
}

func (f *CFeature[MakeTypedFeature]) MountGormDBPath(mount, path string, connection string) (mtf MakeTypedFeature) {
	f._gormSupportBuild = append(f._gormSupportBuild, &cMountGormDB{
		mount:      mount,
		path:       path,
		connection: connection,
	})
	mtf, _ = f.This().(MakeTypedFeature)
	return
}

type cMountGormDB struct {
	mount      string
	path       string
	connection string
}

type CGormDBPathSupport[MakeTypedFeature interface{}] struct {
	_gormSupportBuild []*cMountGormDB
}

func (s CGormDBPathSupport[MakeTypedFeature]) initGormDBPathSupport(f *CFeature[MakeTypedFeature]) {
	s._gormSupportBuild = make([]*cMountGormDB, 0)
	return
}

func (s CGormDBPathSupport[MakeTypedFeature]) startupGormDBPathSupport(f *CFeature[MakeTypedFeature], ctx *cli.Context) (err error) {
	for _, mgdb := range s._gormSupportBuild {
		var ok bool
		var db *gorm.DB
		if v := f.Enjin.MustDB(mgdb.connection); v != nil {
			if db, ok = v.(*gorm.DB); !ok {
				err = fmt.Errorf("connection error: %v; expected *gorm.DB, found %T", mgdb.connection)
				return
			}
		} else {
			err = fmt.Errorf("database connection not found: %v", mgdb.connection)
			return
		}
		table := f.Tag().Snake()
		if mgdb.mount != "/" {
			table += "_" + beStrings.PathToSnake(mgdb.mount)
		}
		var gfs *beFsGormDB.DBFileSystem
		log.DebugF("mounting gorm db: mount=%v, path=%v, table=%v - %v", mgdb.mount, mgdb.path, table, mgdb.connection)
		if gfs, err = beFsGormDB.New(f.Tag().String(), mgdb.path, table, mgdb.connection, db); err != nil {
			log.FatalF("error mounting gorm db: %v", err)
		} else {
			f.MountPathRWFS(mgdb.path, mgdb.mount, gfs)
		}
	}
	return
}

func (f *CFeature[MakeTypedFeature]) GormTx(path string) (tx *gorm.DB) {
	for _, mp := range f.FindPathMountPoint(path) {
		if gfs, ok := mp.RWFS.(fs.GormFileSystem); ok {
			// first gorm filesystem wins
			return gfs.GormTx()
		}
	}
	return
}