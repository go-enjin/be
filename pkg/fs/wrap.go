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

package fs

import (
	"io/fs"

	"github.com/go-enjin/be/pkg/log"
	bePath "github.com/go-enjin/be/pkg/path"
)

type WrapFileSystem struct {
	path string
	fs   FileSystem
}

func Wrap(path string, fs FileSystem) (out FileSystem, err error) {
	out = WrapFileSystem{
		path: path,
		fs:   fs,
	}
	return
}

func (w WrapFileSystem) realpath(path string) (rp string) {
	rp = bePath.SafeConcatRelPath(w.path, path)
	return
}

func (w WrapFileSystem) Name() (name string) {
	name = w.fs.Name()
	return
}

func (w WrapFileSystem) Open(path string) (file fs.File, err error) {
	file, err = w.fs.Open(w.realpath(path))
	return
}

func (w WrapFileSystem) ListDirs(path string) (paths []string, err error) {
	paths, err = w.fs.ListDirs(w.realpath(path))
	return
}

func (w WrapFileSystem) ListFiles(path string) (paths []string, err error) {
	paths, err = w.fs.ListFiles(w.realpath(path))
	return
}

func (w WrapFileSystem) ListAllDirs(path string) (paths []string, err error) {
	paths, err = w.fs.ListAllDirs(w.realpath(path))
	return
}

func (w WrapFileSystem) ListAllFiles(path string) (paths []string, err error) {
	paths, err = w.fs.ListAllFiles(w.realpath(path))
	return
}

func (w WrapFileSystem) ReadDir(path string) (entries []fs.DirEntry, err error) {
	entries, err = w.fs.ReadDir(w.realpath(path))
	return
}

func (w WrapFileSystem) ReadFile(path string) (data []byte, err error) {
	log.WarnF("read file path: %v - %v", w.realpath(path), path)
	data, err = w.fs.ReadFile(w.realpath(path))
	return
}

func (w WrapFileSystem) MimeType(path string) (mime string, err error) {
	if mime, err = w.fs.MimeType(w.realpath(path)); err != nil {
		mime = "application/octet-stream"
	}
	return
}

func (w WrapFileSystem) Shasum(path string) (shasum string, err error) {
	shasum, err = w.fs.Shasum(w.realpath(path))
	return
}