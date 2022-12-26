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

package path

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fvbommel/sortorder"
	"github.com/gabriel-vasile/mimetype"
	"github.com/yookoala/realpath"
)

var (
	ErrorDirNotFound = fmt.Errorf(`not found or not an existing directory`)
)

// Base returns the name of the file without the extension
func Base(path string) (name string) {
	name = filepath.Base(path)
	extn := filepath.Ext(name)
	if extn != "" {
		name = name[:len(name)-len(extn)]
	}
	return
}

// Ext returns the extension of the file (without the dot)
func Ext(path string) (extn string) {
	if extn = filepath.Ext(path); extn != "" {
		extn = extn[1:]
	}
	return
}

// ExtExt returns the extension of the file (without the dot) and any secondary
// extension found in the path
func ExtExt(path string) (extn, extra string) {
	extn = Ext(path)
	trimmed := TrimExt(path)
	extra = Ext(trimmed)
	return
}

// TrimExt returns the path without any file extension
func TrimExt(path string) (out string) {
	if extn := filepath.Ext(path); extn != "" {
		out = path[0 : len(path)-len(extn)]
	}
	return
}

// Mime returns the MIME type string of a local file
func Mime(path string) (mime string, err error) {
	if !IsFile(path) {
		err = fmt.Errorf("%v is not a file", path)
		return
	}

	if mime = MimeFromPathOnly(path); mime == "" {
		if mt, err := mimetype.DetectFile(path); err == nil {
			mime = mt.String()
		}
	}

	if mime == "" {
		mime = "application/octet-stream"
	}
	return
}

func MimeFromPathOnly(path string) (mime string) {
	switch Ext(path) {
	case "scss", "sass":
		mime = "text/x-scss; charset=utf-8"
	case "css":
		mime = "text/css; charset=utf-8"
	case "html":
		mime = "text/html; charset=utf-8"
	case "js":
		mime = "text/javascript; charset=utf-8"
	}
	return
}

// Exists returns true if the path is present on the local filesystem
func Exists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	} else if errors.Is(err, os.ErrNotExist) {
		// path does *not* exist
	} else {
		// Schrödinger: file may or may not exist. See err for details.
	}
	return false
}

// IsFile returns true if the path is an existing file
func IsFile(path string) bool {
	if info, err := os.Stat(path); err == nil {
		return info.IsDir() == false
	} else if errors.Is(err, os.ErrNotExist) {
		// path does *not* exist
	} else {
		// Schrodinger: file may or may not exist. See err for details.
		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence
	}
	return false
}

// IsDir returns true if the path is an existing directory
func IsDir(path string) bool {
	if info, err := os.Stat(path); err == nil {
		return info.IsDir()
	} else if errors.Is(err, os.ErrNotExist) {
		// path does *not* exist
	} else {
		// Schrodinger: file may or may not exist. See err for details.
		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence
	}
	return false
}

func ListDirs(path string) (paths []string, err error) {
	var entries []os.DirEntry
	if entries, err = os.ReadDir(path); err == nil {
		for _, info := range entries {
			if info.IsDir() {
				paths = append(paths, TrimSlashes(Join(path, info.Name())))
			}
		}
	}
	sort.Sort(sortorder.Natural(paths))
	return
}

func ListFiles(path string) (paths []string, err error) {
	var entries []os.DirEntry
	if entries, err = os.ReadDir(path); err == nil {
		for _, info := range entries {
			if !info.IsDir() {
				paths = append(paths, TrimSlash(Join(path, info.Name())))
			}
		}
	}
	sort.Sort(sortorder.Natural(paths))
	return
}

func ListAllDirs(path string) (paths []string, err error) {
	var entries []os.DirEntry
	if entries, err = os.ReadDir(path); err == nil {
		for _, info := range entries {
			thisPath := TrimSlashes(Join(path, info.Name()))
			if info.IsDir() {
				paths = append(paths, thisPath)
				if subDirs, err := ListAllDirs(thisPath); err == nil && len(subDirs) > 0 {
					paths = append(paths, subDirs...)
				}
			}
		}
	}
	sort.Sort(sortorder.Natural(paths))
	return
}

func ListAllFiles(path string) (paths []string, err error) {
	var entries []os.DirEntry
	if entries, err = os.ReadDir(path); err == nil {
		for _, info := range entries {
			thisPath := filepath.Clean(Join(path, info.Name()))
			if !info.IsDir() {
				paths = append(paths, thisPath)
				continue
			}
			var moreFiles []string
			if moreFiles, err = ListAllFiles(thisPath); err == nil {
				paths = append(paths, moreFiles...)
			} else {
				err = fmt.Errorf("error listing all files: %v - %v", err, thisPath)
				return
			}
		}
	}
	sort.Sort(sortorder.Natural(paths))
	return
}

func IsHiddenPath(path string) (hidden bool) {
	for _, part := range strings.Split(path, "/") {
		if part != "" && part != "." && part[0] == '.' {
			hidden = true
			return
		}
	}
	return
}

func FindAllDirs(path string, includeHidden bool) (dirs []string, err error) {
	var all []string
	if all, err = ListAllDirs(path); err != nil {
		return
	}
	for _, dir := range all {
		if includeHidden || !IsHiddenPath(dir) {
			dirs = append(dirs, dir)
		}
	}
	return
}

func FindAllFiles(path string, includeHidden bool) (files []string, err error) {
	var all []string
	if all, err = ListAllFiles(path); err != nil {
		return
	}
	for _, file := range all {
		if includeHidden || !IsHiddenPath(file) {
			files = append(files, file)
		}
	}
	return
}

func TrimRelativeToRoot(path, root string) (rel string) {
	rl := len(root)
	if len(path) > rl {
		if path[:rl] == root {
			rel = path[rl+1:]
			rel = TrimSlashes(rel)
		}
	}
	return
}

func FindFileRelativeToPwd(name string) (file string) {
	file = FindFileRelativeToPath(name, ".")
	return
}

func FindFileRelativeToPath(name, path string) (file string) {
	if abs, err := Abs(path); err == nil {
		if IsFile(name) {
			file = abs + "/" + name
			return
		}
		parts := strings.Split(abs, "/")
		parts = parts[1:]
		pl := len(parts)
		for i := pl - 1; i >= 0; i-- {
			combined := "/" + strings.Join(parts[0:i], "/") + "/" + name
			if IsFile(combined) {
				file = combined
				return
			}
		}
	}
	return
}

func PruneEmptyDirs(path string) (err error) {
	var all []string
	if all, err = FindAllDirs(path, true); err != nil {
		return
	}
	for _, dir := range all {
		var files []string
		if files, err = FindAllFiles(dir, true); err != nil {
			return
		}
		if len(files) == 0 {
			if err = os.Remove(dir); err != nil {
				return
			}
		}
	}
	return
}

func CopyFile(src, dst string) (int64, error) {
	// see: https://opensource.com/article/18/6/copying-files-go
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func ChmodAll(src string) {
	_ = Walk(
		src,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				_ = os.Chmod(path, 0770)
			} else {
				_ = os.Chmod(path, 0660)
			}
			return nil
		},
	)
}

func Pwd() (path string) {
	path, _ = os.Getwd()
	return
}

func Mkdir(path string) (err error) {
	if !IsDir(path) {
		if err = os.MkdirAll(path, 0770); err != nil {
			return
		}
	}
	return
}

func Which(name string) (path string) {
	ln := len(name)
	if ln > 1 {
		if name[0] == '/' {
			if rp, err := realpath.Realpath(name); err == nil {
				path = rp
				return
			}
		}
	}
	if ln > 3 {
		if name[0:2] == "./" || name[0:3] == "../" {
			if rp, err := realpath.Realpath(name); err == nil {
				path = rp
				return
			}
			path = name
			return
		}
	}
	envPath := os.Getenv("PATH")
	parts := strings.Split(envPath, ":")
	for _, part := range parts {
		check := part + "/" + name
		if IsFile(check) {
			if rp, err := realpath.Realpath(check); err == nil {
				path = rp
			} else {
				path = check
			}
			return
		}
	}
	path = "" // command not found
	return
}