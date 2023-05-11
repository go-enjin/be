//go:build fs_drivers_zip || fs_drivers || all

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
	"io/fs"
	"strings"
	"time"

	"github.com/go-enjin/github-com-djherbis-times"
	"github.com/spkg/zipfs"

	beFs "github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/globals"
	bePath "github.com/go-enjin/be/pkg/path"
	bePathZip "github.com/go-enjin/be/pkg/path/zip"
)

type CFileSystem struct {
	path string
	wrap string
	zip  *zipfs.FileSystem
}

func New(path string, zfs *zipfs.FileSystem) (out beFs.FileSystem, err error) {
	out = CFileSystem{
		path: path,
		wrap: "",
		zip:  zfs,
	}
	return
}

func Wrap(path, wrap string, zfs *zipfs.FileSystem) (out CFileSystem, err error) {
	out = CFileSystem{
		path: path,
		zip:  zfs,
	}
	return
}

func (f CFileSystem) Name() (name string) {
	name = f.path
	return
}

func (f CFileSystem) realpath(path string) (out string) {
	out = bePath.SafeConcatRelPath(f.path, path)
	return
}

func (f CFileSystem) pruneEntries(paths []string) (pruned []string) {
	rp := f.path
	for _, entry := range paths {
		if strings.HasPrefix(entry, rp) {
			if entry = entry[len(rp):]; entry != "" && entry[0] == '/' {
				entry = entry[1:]
			}
		}
		pruned = append(pruned, entry)
	}
	return
}

func (f CFileSystem) Exists(path string) (exists bool) {
	if fh, err := f.zip.Open(f.realpath(path)); err == nil {
		_ = fh.Close()
		exists = true
	}
	return
}

func (f CFileSystem) Open(path string) (file fs.File, err error) {
	file, err = f.zip.Open(f.realpath(path))
	return
}

func (f CFileSystem) ListDirs(path string) (paths []string, err error) {
	if paths, err = bePathZip.ListDirs(f.realpath(path), f.zip); err == nil {
		paths = f.pruneEntries(paths)
	}
	return
}

func (f CFileSystem) ListFiles(path string) (paths []string, err error) {
	if paths, err = bePathZip.ListFiles(f.realpath(path), f.zip); err == nil {
		paths = f.pruneEntries(paths)
	}
	return
}

func (f CFileSystem) ListAllDirs(path string) (paths []string, err error) {
	if paths, err = bePathZip.ListAllDirs(f.realpath(path), f.zip); err == nil {
		paths = f.pruneEntries(paths)
	}
	return
}

func (f CFileSystem) ListAllFiles(path string) (paths []string, err error) {
	if paths, err = bePathZip.ListAllFiles(f.realpath(path), f.zip); err == nil {
		paths = f.pruneEntries(paths)
	}
	return
}

func (f CFileSystem) ReadDir(path string) (paths []fs.DirEntry, err error) {
	paths, err = bePathZip.ReadDir(f.realpath(path), f.zip)
	return
}

func (f CFileSystem) ReadFile(path string) (content []byte, err error) {
	content, err = bePathZip.ReadFile(f.realpath(path), f.zip)
	return
}

func (f CFileSystem) MimeType(path string) (mime string, err error) {
	if mime, err = bePathZip.Mime(f.realpath(path), f.zip); err != nil {
		mime = "application/octet-stream"
	}
	return
}

func (f CFileSystem) Shasum(path string) (shasum string, err error) {
	shasum, err = bePathZip.Shasum(f.realpath(path), f.zip)
	return
}

func (f CFileSystem) FileCreated(_ string) (created int64, err error) {
	var info times.Timespec
	if info, err = globals.BuildFileInfo(); err == nil && info.HasBirthTime() {
		created = info.BirthTime().Unix()
	}
	return
}

func (f CFileSystem) LastModified(_ string) (modTime int64, err error) {
	var info times.Timespec
	if info, err = globals.BuildFileInfo(); err == nil {
		modTime = info.ModTime().Unix()
	}
	return
}

func (f CFileSystem) FileStats(path string) (mime, shasum string, created, updated time.Time, err error) {
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