// Copyright (c) 2023  The Go-Enjin Authors
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
	"os"
	"sync"
)

var (
	gRegistries = struct {
		order []string
		known map[string]*fsRegistry
		sync.RWMutex
	}{
		order: []string{},
		known: make(map[string]*fsRegistry),
	}
)

func GetFileSystem(id string) (f FileSystem, ok bool) {
	gRegistries.RLock()
	defer gRegistries.RUnlock()
	for _, rid := range gRegistries.order {
		if f, ok = gRegistries.known[rid].GetFileSystem(id); ok {
			return
		}
	}
	return
}

func FileExists(path string) (exists bool) {
	gRegistries.RLock()
	defer gRegistries.RUnlock()
	for _, rid := range gRegistries.order {
		if exists = gRegistries.known[rid].FileExists(path); exists {
			return
		}
	}
	return
}

func FindFileShasum(path string) (shasum string, err error) {
	gRegistries.RLock()
	defer gRegistries.RUnlock()
	for _, rid := range gRegistries.order {
		if shasum, err = gRegistries.known[rid].FindFileShasum(path); err == nil {
			return
		}
	}
	err = os.ErrNotExist
	return
}

func FindFileMime(path string) (mime string, err error) {
	gRegistries.RLock()
	defer gRegistries.RUnlock()
	for _, rid := range gRegistries.order {
		if mime, err = gRegistries.known[rid].FindFileMime(path); err == nil {
			return
		}
	}
	err = os.ErrNotExist
	return
}

func ListFiles(path string) (files []string) {
	gRegistries.RLock()
	defer gRegistries.RUnlock()
	for _, rid := range gRegistries.order {
		if found := gRegistries.known[rid].ListFiles(path); len(found) > 0 {
			files = append(files, found...)
		}
	}
	return
}

func ListAllFiles(path string) (files []string) {
	gRegistries.RLock()
	defer gRegistries.RUnlock()
	for _, rid := range gRegistries.order {
		if found := gRegistries.known[rid].ListAllFiles(path); len(found) > 0 {
			files = append(files, found...)
		}
	}
	return
}

func ListDirs(path string) (dirs []string, err error) {
	gRegistries.RLock()
	defer gRegistries.RUnlock()
	for _, rid := range gRegistries.order {
		if found := gRegistries.known[rid].ListDirs(path); len(found) > 0 {
			dirs = append(dirs, found...)
		}
	}
	return
}

func ListAllDirs(path string) (dirs []string, err error) {
	gRegistries.RLock()
	defer gRegistries.RUnlock()
	for _, rid := range gRegistries.order {
		if found := gRegistries.known[rid].ListAllDirs(path); len(found) > 0 {
			dirs = append(dirs, found...)
		}
	}
	return
}
