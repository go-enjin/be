//go:build driver_fs_zip || drivers_fs || drivers || zips || all

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

package zip

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spkg/zipfs"

	"github.com/go-enjin/github-com-djherbis-times"

	beFs "github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/gob"
	bePath "github.com/go-enjin/be/pkg/path"
	bePathZip "github.com/go-enjin/be/pkg/path/zip"
	beStrings "github.com/go-enjin/be/pkg/strings"
	"github.com/go-enjin/be/types/page/matter"
)

func init() {
	gob.Register(FileSystem{})
}

type FileSystem struct {
	origin string
	path   string
	wrap   string
	zip    *zipfs.FileSystem
	id     string
}

func New(origin string, path string, zfs *zipfs.FileSystem) (out FileSystem, err error) {
	out = FileSystem{
		origin: origin,
		path:   path,
		wrap:   "",
		zip:    zfs,
		id:     fmt.Sprintf("%v://%v", origin, path),
	}
	return
}

func (f FileSystem) ID() (id string) {
	return f.id
}

func (f FileSystem) Name() (name string) {
	name = f.path
	return
}

func (f FileSystem) CloneROFS() (cloned beFs.FileSystem) {
	cloned = FileSystem{
		origin: f.origin,
		path:   f.path,
		wrap:   f.wrap,
		zip:    f.zip,
		id:     f.id,
	}
	return
}

func (f FileSystem) Exists(path string) (exists bool) {
	if fh, err := f.zip.Open(f.realpath(path)); err == nil {
		_ = fh.Close()
		exists = true
	}
	return
}

func (f FileSystem) Open(path string) (file fs.File, err error) {
	file, err = f.zip.Open(f.realpath(path))
	return
}

func (f FileSystem) ListDirs(path string) (paths []string, err error) {
	if paths, err = bePathZip.ListDirs(f.realpath(path), f.zip); err == nil {
		paths = beFs.PruneRootFrom(f.path, paths)
	}
	return
}

func (f FileSystem) ListFiles(path string) (paths []string, err error) {
	if paths, err = bePathZip.ListFiles(f.realpath(path), f.zip); err == nil {
		paths = beFs.PruneRootFrom(f.path, paths)
	}
	return
}

func (f FileSystem) ListAllDirs(path string) (paths []string, err error) {
	if paths, err = bePathZip.ListAllDirs(f.realpath(path), f.zip); err == nil {
		paths = beFs.PruneRootFrom(f.path, paths)
	}
	return
}

func (f FileSystem) ListAllFiles(path string) (paths []string, err error) {
	if paths, err = bePathZip.ListAllFiles(f.realpath(path), f.zip); err == nil {
		paths = beFs.PruneRootFrom(f.path, paths)
	}
	return
}

func (f FileSystem) ReadDir(path string) (paths []fs.DirEntry, err error) {
	paths, err = bePathZip.ReadDir(f.realpath(path), f.zip)
	return
}

func (f FileSystem) ReadFile(path string) (content []byte, err error) {
	content, err = bePathZip.ReadFile(f.realpath(path), f.zip)
	return
}

func (f FileSystem) MimeType(path string) (mime string, err error) {
	if mime, err = bePathZip.Mime(f.realpath(path), f.zip); err != nil {
		mime = "application/octet-stream"
	}
	return
}

func (f FileSystem) Shasum(path string) (shasum string, err error) {
	shasum, err = bePathZip.Shasum(f.realpath(path), f.zip)
	return
}

func (f FileSystem) FileCreated(_ string) (created int64, err error) {
	var info times.Timespec
	if info, err = globals.BuildFileInfo(); err == nil && info.HasBirthTime() {
		created = info.BirthTime().Unix()
	}
	return
}

func (f FileSystem) LastModified(_ string) (modTime int64, err error) {
	var info times.Timespec
	if info, err = globals.BuildFileInfo(); err == nil {
		modTime = info.ModTime().Unix()
	}
	return
}

func (f FileSystem) FileStats(path string) (mime, shasum string, created, updated time.Time, err error) {
	if mime, err = f.MimeType(path); err != nil {
		return
	}
	if shasum, err = f.Shasum(path); err != nil {
		return
	}
	var ts times.Timespec
	if ts, err = globals.BuildFileInfo(); err != nil {
		return
	}
	updated = ts.ModTime()
	if ts.HasBirthTime() {
		created = ts.BirthTime()
	}
	return
}

func (f FileSystem) FindFilePath(prefix string, extensions ...string) (path string, err error) {

	realpath := f.realpath(prefix)
	if filepath.Ext(realpath) != "" {
		if bePath.IsFile(realpath) {
			path = beFs.PruneRootFrom(f.path, realpath)
			return
		}
	}

	sort.Sort(beStrings.SortByLengthDesc(extensions))

	realpath = strings.TrimSuffix(realpath, "/")
	var paths []string
	for _, extension := range extensions {
		paths = append(paths, realpath+"."+extension)
	}

	for _, p := range paths {
		if bePath.IsFile(p) {
			path = beFs.PruneRootFrom(f.path, p)
			return
		}
	}

	err = os.ErrNotExist
	return
}

func (f FileSystem) ReadPageMatter(path string) (pm *matter.PageMatter, err error) {

	if f.Exists(path) {
		var data []byte
		if data, err = f.ReadFile(path); err != nil {
			return
		}
		_, _, created, updated, _ := f.FileStats(path)
		pm, err = matter.ParsePageMatter(f.origin, path, created, updated, data)
		return
	}

	err = os.ErrNotExist
	return
}