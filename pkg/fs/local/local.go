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

	"github.com/go-enjin/be/pkg/hash/sha"
	bePath "github.com/go-enjin/be/pkg/path"
)

type FileSystem string

func New(path string) (out FileSystem, err error) {
	if bePath.IsDir(path) {
		out = FileSystem(path)
		return
	}
	err = bePath.ErrorDirNotFound
	return
}

func (f FileSystem) Name() (name string) {
	name = string(f)
	return
}

func (f FileSystem) realpath(path string) (out string) {
	out = bePath.SafeConcatRelPath(string(f), path)
	return
}

func (f FileSystem) Open(path string) (file fs.File, err error) {
	file, err = os.Open(f.realpath(path))
	return
}

func (f FileSystem) ListDirs(path string) (paths []string, err error) {
	paths, err = bePath.ListDirs(f.realpath(path))
	return
}

func (f FileSystem) ListFiles(path string) (paths []string, err error) {
	paths, err = bePath.ListFiles(f.realpath(path))
	return
}

func (f FileSystem) ListAllDirs(path string) (paths []string, err error) {
	paths, err = bePath.ListAllDirs(f.realpath(path))
	return
}

func (f FileSystem) ListAllFiles(path string) (paths []string, err error) {
	paths, err = bePath.ListAllFiles(f.realpath(path))
	return
}

func (f FileSystem) ReadDir(path string) (paths []fs.DirEntry, err error) {
	paths, err = os.ReadDir(f.realpath(path))
	return
}

func (f FileSystem) ReadFile(path string) (content []byte, err error) {
	content, err = os.ReadFile(f.realpath(path))
	return
}

func (f FileSystem) MimeType(path string) (mime string, err error) {
	if mime, err = bePath.Mime(f.realpath(path)); err != nil {
		mime = "application/octet-stream"
	}
	return
}

func (f FileSystem) Shasum(path string) (shasum string, err error) {
	shasum, err = sha.FileHash10(f.realpath(path))
	return
}

func (f FileSystem) LastModified(path string) (modTime int64, err error) {
	var info os.FileInfo
	if info, err = os.Stat(path); err != nil {
		return
	}
	modTime = info.ModTime().Unix()
	return
}