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
	"sort"

	"github.com/fvbommel/sortorder"
	"github.com/gabriel-vasile/mimetype"

	"github.com/go-enjin/be/pkg/hash/sha"
	bePath "github.com/go-enjin/be/pkg/path"
)

func Shasum(path string, efs embed.FS) (shasum string, err error) {
	var data []byte
	if data, err = efs.ReadFile(path); err != nil {
		return
	}
	shasum, err = sha.DataHash10(data)
	return
}

func Mime(path string, efs embed.FS) (mime string, err error) {
	var data []byte
	if data, err = efs.ReadFile(path); err != nil {
		return
	}

	if mime = bePath.MimeFromPathOnly(path); mime == "" {
		mime = mimetype.Detect(data).String()
	}

	if mime == "" {
		mime = "application/octet-stream"
	}
	return
}

func ListDirs(path string, efs embed.FS) (paths []string, err error) {
	var entries []fs.DirEntry
	if entries, err = efs.ReadDir(path); err != nil {
		return
	}
	for _, info := range entries {
		if info.IsDir() {
			paths = append(paths, path+string(os.PathSeparator)+info.Name())
		}
	}
	sort.Sort(sortorder.Natural(paths))
	return
}

func ListFiles(path string, efs embed.FS) (paths []string, err error) {
	var entries []fs.DirEntry
	if entries, err = efs.ReadDir(path); err != nil {
		return
	}
	for _, info := range entries {
		if !info.IsDir() {
			paths = append(paths, path+string(os.PathSeparator)+info.Name())
		}
	}
	sort.Sort(sortorder.Natural(paths))
	return
}

func ListAllDirs(path string, efs embed.FS) (paths []string, err error) {
	var entries []os.DirEntry
	if entries, err = efs.ReadDir(path); err == nil {
		for _, info := range entries {
			thisPath := path + string(os.PathSeparator) + info.Name()
			if info.IsDir() {
				paths = append(paths, thisPath)
				if subDirs, err := ListAllDirs(thisPath, efs); err == nil && len(subDirs) > 0 {
					paths = append(paths, subDirs...)
				}
			}
		}
	}
	sort.Sort(sortorder.Natural(paths))
	return
}

func ListAllFiles(path string, efs embed.FS) (paths []string, err error) {
	var entries []os.DirEntry
	if entries, err = efs.ReadDir(path); err == nil {
		for _, info := range entries {
			thisPath := path + string(os.PathSeparator) + info.Name()
			if !info.IsDir() {
				paths = append(paths, thisPath)
				continue
			}
			if moreFiles, err := ListAllFiles(thisPath, efs); err == nil && len(moreFiles) > 0 {
				paths = append(paths, moreFiles...)
			}
		}
	}
	sort.Sort(sortorder.Natural(paths))
	return
}