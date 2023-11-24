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
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/spkg/zipfs"

	"github.com/go-enjin/be/pkg/hash/sha"
	beMime "github.com/go-enjin/be/pkg/mime"
)

func ReadDir(path string, zfs *zipfs.FileSystem) (entries []fs.DirEntry, err error) {
	var hf http.File
	if hf, err = zfs.Open(path); err != nil {
		return
	}
	defer hf.Close()
	if s, e := hf.Stat(); e == nil && !s.IsDir() {
		err = fmt.Errorf("not a directory: %v", path)
		return
	}
	var infos []fs.FileInfo
	if infos, err = hf.Readdir(0); err != nil {
		return
	}
	for _, info := range infos {
		entries = append(entries, &dirEntry{info: info})
	}
	return
}

func ReadFile(path string, zfs *zipfs.FileSystem) (data []byte, err error) {
	var hf http.File
	if hf, err = zfs.Open(path); err != nil {
		return
	}
	defer hf.Close()
	if s, e := hf.Stat(); e == nil && s.IsDir() {
		err = fmt.Errorf("is a directory")
		return
	}
	if data, err = io.ReadAll(hf); err != nil {
		return
	}
	return
}

func Shasum(path string, zfs *zipfs.FileSystem) (shasum string, err error) {
	var data []byte
	if data, err = ReadFile(path, zfs); err != nil {
		return
	}
	shasum, err = sha.DataHash10(data)
	return
}

func Sha256(path string, zfs *zipfs.FileSystem) (shasum string, err error) {
	var data []byte
	if data, err = ReadFile(path, zfs); err != nil {
		return
	}
	shasum, err = sha.Hash256(data)
	return
}

func Mime(path string, zfs *zipfs.FileSystem) (mime string, err error) {
	var data []byte
	if data, err = ReadFile(path, zfs); err != nil {
		if err.Error() == "is a directory" {
			err = nil
			mime = beMime.DirectoryMimeType
		}
		return
	}

	if mime = beMime.FromPathOnly(path); mime == "" {
		mime = mimetype.Detect(data).String()
	}

	if mime == "" {
		mime = beMime.BinaryMimeType
	}
	return
}

func ListDirs(path string, zfs *zipfs.FileSystem) (dirs []string, err error) {
	var hf http.File
	if hf, err = zfs.Open(path); err != nil {
		return
	}
	defer hf.Close()
	var infos []fs.FileInfo
	if infos, err = hf.Readdir(0); err != nil {
		return
	} else {
		if !strings.HasSuffix(path, "/") {
			path += "/"
		}
		for _, info := range infos {
			if info.IsDir() {
				dirs = append(dirs, path+info.Name())
			}
		}
	}
	// sort.Sort(natural.StringSlice(dirs))
	return
}

func ListFiles(path string, zfs *zipfs.FileSystem) (files []string, err error) {
	var hf http.File
	if hf, err = zfs.Open(path); err != nil {
		return
	}
	defer hf.Close()
	var infos []fs.FileInfo
	if infos, err = hf.Readdir(0); err != nil {
		return
	} else {
		if !strings.HasSuffix(path, "/") {
			path += "/"
		}
		for _, info := range infos {
			if !info.IsDir() {
				files = append(files, path+info.Name())
			}
		}
	}
	// sort.Sort(natural.StringSlice(files))
	return
}

func ListAllDirs(path string, zfs *zipfs.FileSystem) (all []string, err error) {
	var dirs []string
	if dirs, err = ListDirs(path, zfs); err != nil {
		return
	}
	for _, dir := range dirs {
		all = append(all, dir)
		if more, e := ListAllDirs(dir, zfs); e != nil {
			err = e
			return
		} else if len(more) > 0 {
			all = append(all, more...)
		}
	}
	// sort.Sort(natural.StringSlice(all))
	return
}

func ListAllFiles(path string, zfs *zipfs.FileSystem) (all []string, err error) {
	if all, err = ListFiles(path, zfs); err != nil {
		return
	}
	var dirs []string
	if dirs, err = ListDirs(path, zfs); err != nil {
		return
	}
	for _, dir := range dirs {
		if more, e := ListAllFiles(dir, zfs); e != nil {
			err = e
			return
		} else if len(more) > 0 {
			all = append(all, more...)
		}
	}
	// sort.Sort(natural.StringSlice(all))
	return
}