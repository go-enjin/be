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

package embed

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"

	bePath "github.com/go-enjin/be/pkg/path"
	bePathEmbed "github.com/go-enjin/be/pkg/path/embed"
)

type FileSystem struct {
	path  string
	wrap  string
	embed embed.FS
}

func New(path string, efs embed.FS) (out FileSystem, err error) {
	out = FileSystem{
		path:  path,
		wrap:  "",
		embed: efs,
	}
	return
}

func Wrap(path, wrap string, efs embed.FS) (out FileSystem, err error) {
	out = FileSystem{
		path:  path,
		wrap:  wrap,
		embed: efs,
	}
	return
}

func (f FileSystem) Name() (name string) {
	name = "embed"
	return
}

func (f FileSystem) realpath(path string) (out string) {
	if f.wrap != "" {
		out = bePath.SafeConcatRelPath(f.path, f.wrap, path)
	} else {
		out = bePath.SafeConcatRelPath(f.path, path)
	}
	return
}

func (f FileSystem) Open(path string) (file fs.File, err error) {
	file, err = f.embed.Open(f.realpath(path))
	return
}

func (f FileSystem) ListDirs(path string) (paths []string, err error) {
	paths, err = bePathEmbed.ListDirs(f.realpath(path), f.embed)
	return
}

func (f FileSystem) ListFiles(path string) (paths []string, err error) {
	paths, err = bePathEmbed.ListFiles(f.realpath(path), f.embed)
	return
}

func (f FileSystem) ListAllDirs(path string) (paths []string, err error) {
	paths, err = bePathEmbed.ListAllDirs(f.realpath(path), f.embed)
	return
}

func (f FileSystem) ListAllFiles(path string) (paths []string, err error) {
	paths, err = bePathEmbed.ListAllFiles(f.realpath(path), f.embed)
	return
}

func (f FileSystem) ReadDir(path string) (paths []fs.DirEntry, err error) {
	paths, err = f.embed.ReadDir(f.realpath(path))
	return
}

func (f FileSystem) ReadFile(path string) (content []byte, err error) {
	content, err = f.embed.ReadFile(f.realpath(path))
	return
}

func (f FileSystem) MimeType(path string) (mime string, err error) {
	if mime, err = bePathEmbed.Mime(f.realpath(path), f.embed); err != nil {
		mime = "application/octet-stream"
	}
	return
}

func (f FileSystem) Shasum(path string) (shasum string, err error) {
	shasum, err = bePathEmbed.Shasum(f.realpath(path), f.embed)
	return
}

func (f FileSystem) LastModified(path string) (modTime int64, err error) {
	var info os.FileInfo
	var name, tgt string
	if name, err = os.Executable(); err == nil {
		if tgt, err = filepath.EvalSymlinks(name); err == nil {
			if info, err = os.Stat(tgt); err == nil {
				modTime = info.ModTime().Unix()
			}
		}
	}
	return
}