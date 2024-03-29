//go:build driver_fs_local || drivers_fs || drivers || locals || all

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

package local

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	berrs "github.com/go-enjin/be/pkg/errors"
	times "github.com/go-enjin/github-com-djherbis-times"

	clMime "github.com/go-corelibs/mime"
	clPath "github.com/go-corelibs/path"
	sha "github.com/go-corelibs/shasum"
	clStrings "github.com/go-corelibs/strings"
	beFs "github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/gob"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/types/page/matter"
)

var (
	DefaultDirMode  fs.FileMode = 0770
	DefaultFileMode fs.FileMode = 0660
)

func init() {
	gob.Register(FileSystem{})
}

type FileSystem struct {
	origin string
	root   string
	id     string

	sync.RWMutex
}

func New(origin string, path string) (out *FileSystem, err error) {
	if clPath.IsDir(path) {
		if filepath.IsAbs(path) {
			var relPath string
			if relPath, err = filepath.Rel(clPath.Pwd(), path); err != nil {
				err = fmt.Errorf("unable to find relative path: %v - %v", path, err)
				return
			} else {
				path = relPath
			}
		}
		out = &FileSystem{
			origin: origin,
			root:   path,
			id:     fmt.Sprintf("%v://%v", origin, path),
		}
		return
	}
	err = fmt.Errorf("error constructing FileSystem: %v - %v", berrs.ErrDirNotFound, path)
	return
}

func (f *FileSystem) ID() (id string) {
	return f.id
}

func (f *FileSystem) CloneROFS() (cloned beFs.FileSystem) {
	cloned = f.CloneRWFS()
	return
}

func (f *FileSystem) Name() (name string) {
	f.RLock()
	defer f.RUnlock()

	name = f.root
	return
}

func (f *FileSystem) Exists(path string) (exists bool) {
	f.RLock()
	defer f.RUnlock()

	realpath := f.realpath(path)
	_, err := os.Stat(realpath)
	exists = err == nil
	return
}

func (f *FileSystem) Open(path string) (file fs.File, err error) {
	f.RLock()
	defer f.RUnlock()

	file, err = os.Open(f.realpath(path))
	return
}

func (f *FileSystem) ListDirs(path string) (paths []string, err error) {
	f.RLock()
	defer f.RUnlock()

	if paths, err = clPath.ListDirs(f.realpath(path), true); err == nil {
		paths = beFs.PruneRootFrom(f.root, paths)
	}
	return
}

func (f *FileSystem) ListFiles(path string) (paths []string, err error) {
	f.RLock()
	defer f.RUnlock()

	if paths, err = clPath.ListFiles(f.realpath(path), true); err == nil {
		paths = beFs.PruneRootFrom(f.root, paths)
	}
	return
}

func (f *FileSystem) ListAllDirs(path string) (paths []string, err error) {
	f.RLock()
	defer f.RUnlock()

	if paths, err = clPath.ListAllDirs(f.realpath(path), true); err == nil {
		paths = beFs.PruneRootFrom(f.root, paths)
	}
	return
}

func (f *FileSystem) ListAllFiles(path string) (paths []string, err error) {
	f.RLock()
	defer f.RUnlock()

	if paths, err = clPath.ListAllFiles(f.realpath(path), true); err == nil {
		paths = beFs.PruneRootFrom(f.root, paths)
	}
	return
}

func (f *FileSystem) ReadDir(path string) (paths []fs.DirEntry, err error) {
	f.RLock()
	defer f.RUnlock()

	paths, err = os.ReadDir(f.realpath(path))
	return
}

func (f *FileSystem) ReadFile(path string) (content []byte, err error) {
	f.RLock()
	defer f.RUnlock()

	content, err = os.ReadFile(f.realpath(path))
	return
}

func (f *FileSystem) MimeType(path string) (mime string, err error) {
	f.RLock()
	defer f.RUnlock()
	if mime = clMime.Mime(f.realpath(path)); mime == "" {
		mime = "application/octet-stream"
	}
	return
}

func (f *FileSystem) Shasum(path string) (shasum string, err error) {
	f.RLock()
	defer f.RUnlock()

	shasum, err = sha.BriefFile(f.realpath(path))
	return
}

func (f *FileSystem) Sha256(path string) (shasum string, err error) {
	f.RLock()
	defer f.RUnlock()

	shasum, err = sha.File(f.realpath(path))
	return
}

func (f *FileSystem) FileCreated(path string) (created int64, err error) {
	f.RLock()
	defer f.RUnlock()

	var info times.Timespec
	if info, err = times.Stat(f.realpath(path)); err == nil && info.HasBirthTime() {
		created = info.BirthTime().Unix()
	}
	return
}

func (f *FileSystem) LastModified(path string) (updated int64, err error) {
	f.RLock()
	defer f.RUnlock()

	var info times.Timespec
	if info, err = times.Stat(f.realpath(path)); err == nil && info.HasBirthTime() {
		updated = info.ModTime().Unix()
	}
	return
}

func (f *FileSystem) FileStats(path string) (mime, shasum string, created, updated time.Time, err error) {
	f.RLock()
	defer f.RUnlock()

	realpath := f.realpath(path)
	if mime, err = f.MimeType(realpath); err != nil {
		return
	}
	if shasum, err = f.Shasum(realpath); err != nil {
		return
	}
	var ts times.Timespec
	if ts, err = times.Stat(realpath); err != nil {
		return
	}
	updated = ts.ModTime()
	if ts.HasBirthTime() {
		created = ts.BirthTime()
	}
	return
}

func (f *FileSystem) FindFilePath(prefix string, extensions ...string) (path string, err error) {
	f.RLock()
	defer f.RUnlock()

	realpath := f.realpath(prefix)
	if filepath.Ext(realpath) != "" {
		if clPath.IsFile(realpath) {
			path = beFs.PruneRootFrom(f.root, realpath)
			return
		}
	}

	sort.Sort(clStrings.SortByLength(extensions))

	realpath = strings.TrimSuffix(realpath, "/")
	var paths []string
	for _, extension := range extensions {
		paths = append(paths, realpath+"."+extension)
	}

	for _, p := range paths {
		if clPath.IsFile(p) {
			path = beFs.PruneRootFrom(f.root, p)
			return
		}
	}

	err = os.ErrNotExist
	return
}

func (f *FileSystem) ReadPageMatter(path string) (pm *matter.PageMatter, err error) {

	if f.Exists(path) {
		var data []byte
		if data, err = f.ReadFile(path); err != nil {
			return
		}
		_, _, created, updated, _ := f.FileStats(path)
		if pm, err = matter.ParsePageMatter(f.origin, path, created, updated, data); err != nil {
			log.ErrorF("error parsing page matter: %v - %v", path, err)
		}
		return
	}

	err = os.ErrNotExist
	return
}
