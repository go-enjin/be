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
	"fmt"
	"io/fs"
	"os"
	"sort"

	"github.com/gabriel-vasile/mimetype"
	"github.com/maruel/natural"

	clMime "github.com/go-corelibs/mime"
	clPath "github.com/go-corelibs/path"
	sha "github.com/go-corelibs/shasum"
)

func Sha256(path string, efs embed.FS) (shasum string, err error) {
	var data []byte
	if data, err = efs.ReadFile(path); err != nil {
		return
	}
	shasum, err = sha.Sum(data)
	return
}

func Shasum(path string, efs embed.FS) (shasum string, err error) {
	var data []byte
	if data, err = efs.ReadFile(path); err != nil {
		return
	}
	shasum, err = sha.BriefSum(data)
	return
}

func Mime(path string, efs embed.FS) (mime string, err error) {
	var data []byte
	if data, err = efs.ReadFile(path); err != nil {
		if err.Error() == "is a directory" {
			err = nil
			mime = clMime.DirectoryMimeType
		}
		return
	}

	if mime = clMime.FromPathOnly(path); mime == "" {
		mime = mimetype.Detect(data).String()
	}

	if mime == "" {
		mime = clMime.BinaryMimeType
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
			paths = append(paths, clPath.TrimSlashes(clPath.Join(path, info.Name())))
		}
	}
	sort.Sort(natural.StringSlice(paths))
	return
}

func ListFiles(path string, efs embed.FS) (paths []string, err error) {
	var entries []fs.DirEntry
	if entries, err = efs.ReadDir(path); err != nil {
		return
	}
	for _, info := range entries {
		if !info.IsDir() {
			paths = append(paths, clPath.TrimSlashes(clPath.Join(path, info.Name())))
		}
	}
	sort.Sort(natural.StringSlice(paths))
	return
}

func ListAllDirs(path string, efs embed.FS) (paths []string, err error) {
	var entries []os.DirEntry
	if entries, err = efs.ReadDir(path); err == nil {
		for _, info := range entries {
			thisPath := clPath.TrimSlashes(clPath.Join(path, info.Name()))
			if info.IsDir() {
				paths = append(paths, thisPath)
				if subDirs, err := ListAllDirs(thisPath, efs); err == nil && len(subDirs) > 0 {
					paths = append(paths, subDirs...)
				}
			}
		}
	}
	sort.Sort(natural.StringSlice(paths))
	return
}

func ListAllFiles(path string, efs embed.FS) (paths []string, err error) {
	var entries []os.DirEntry
	if entries, err = efs.ReadDir(path); err == nil {
		for _, info := range entries {
			thisPath := clPath.TrimSlashes(clPath.Join(path, info.Name()))
			if !info.IsDir() {
				paths = append(paths, thisPath)
				continue
			}
			var moreFiles []string
			if moreFiles, err = ListAllFiles(thisPath, efs); err == nil {
				paths = append(paths, moreFiles...)
			} else {
				err = fmt.Errorf("error listing all files: %v - %v", err, thisPath)
				return
			}
		}
	}
	sort.Sort(natural.StringSlice(paths))
	return
}
