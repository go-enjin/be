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
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-enjin/github-com-djherbis-times"

	beFs "github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/hash/sha"
	bePath "github.com/go-enjin/be/pkg/path"
)

type CFileSystem string

func New(path string) (out beFs.FileSystem, err error) {
	if bePath.IsDir(path) {
		if filepath.IsAbs(path) {
			var relPath string
			if relPath, err = filepath.Rel(bePath.Pwd(), path); err != nil {
				return
			} else {
				path = relPath
			}
		}
		out = CFileSystem(path)
		return
	}
	err = bePath.ErrorDirNotFound
	return
}

func (f CFileSystem) Name() (name string) {
	name = string(f)
	return
}

func (f CFileSystem) realpath(path string) (out string) {
	out = bePath.SafeConcatRelPath(string(f), path)
	return
}

func (f CFileSystem) pruneEntries(paths []string) (pruned []string) {
	rp := strings.TrimPrefix(string(f), "/")
	for _, entry := range paths {
		entry = strings.TrimPrefix(entry, "/")
		entry = strings.TrimPrefix(entry, rp)
		entry = strings.TrimPrefix(entry, "/")
		pruned = append(pruned, entry)
	}
	return
}

func (f CFileSystem) Exists(path string) (exists bool) {
	_, err := os.Stat(f.realpath(path))
	exists = err == nil
	return
}

func (f CFileSystem) Open(path string) (file fs.File, err error) {
	file, err = os.Open(f.realpath(path))
	return
}

func (f CFileSystem) ListDirs(path string) (paths []string, err error) {
	if paths, err = bePath.ListDirs(f.realpath(path)); err == nil {
		paths = f.pruneEntries(paths)
	}
	return
}

func (f CFileSystem) ListFiles(path string) (paths []string, err error) {
	if paths, err = bePath.ListFiles(f.realpath(path)); err == nil {
		paths = f.pruneEntries(paths)
	}
	return
}

func (f CFileSystem) ListAllDirs(path string) (paths []string, err error) {
	if paths, err = bePath.ListAllDirs(f.realpath(path)); err == nil {
		paths = f.pruneEntries(paths)
	}
	return
}

func (f CFileSystem) ListAllFiles(path string) (paths []string, err error) {
	if paths, err = bePath.ListAllFiles(f.realpath(path)); err == nil {
		paths = f.pruneEntries(paths)
	}
	return
}

func (f CFileSystem) ReadDir(path string) (paths []fs.DirEntry, err error) {
	paths, err = os.ReadDir(f.realpath(path))
	return
}

func (f CFileSystem) ReadFile(path string) (content []byte, err error) {
	content, err = os.ReadFile(f.realpath(path))
	return
}

func (f CFileSystem) MimeType(path string) (mime string, err error) {
	if mime, err = bePath.Mime(f.realpath(path)); err != nil {
		mime = "application/octet-stream"
	}
	return
}

func (f CFileSystem) Shasum(path string) (shasum string, err error) {
	shasum, err = sha.FileHash10(f.realpath(path))
	return
}

func (f CFileSystem) FileCreated(path string) (created int64, err error) {
	var info times.Timespec
	if info, err = times.Stat(f.realpath(path)); err == nil && info.HasBirthTime() {
		created = info.BirthTime().Unix()
	}
	return
}

func (f CFileSystem) LastModified(path string) (updated int64, err error) {
	var info times.Timespec
	if info, err = times.Stat(f.realpath(path)); err == nil && info.HasBirthTime() {
		updated = info.ModTime().Unix()
	}
	return
}

func (f CFileSystem) FileStats(path string) (mime, shasum string, created, updated time.Time, err error) {
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

func (f CFileSystem) MakeDir(path string, perm os.FileMode) (err error) {
	err = os.Mkdir(f.realpath(path), perm)
	return
}

func (f CFileSystem) MakeDirAll(path string, perm os.FileMode) (err error) {
	err = os.MkdirAll(f.realpath(path), perm)
	return
}

func (f CFileSystem) WriteFile(path string, data []byte, perm os.FileMode) (err error) {
	err = os.WriteFile(path, data, perm)
	return
}

func (f CFileSystem) Remove(path string) (err error) {
	err = os.Remove(path)
	return
}

func (f CFileSystem) RemoveAll(path string) (err error) {
	err = os.RemoveAll(path)
	return
}