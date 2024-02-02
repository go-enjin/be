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

	clPath "github.com/go-corelibs/path"
	"github.com/go-enjin/be/pkg/log"
)

type RegistryLookup interface {
	// FileExists returns true if any of the known filesystems have a file matching the given path
	FileExists(path string) (exists bool)

	// ReadFile returns the file data
	ReadFile(path string) (data []byte, err error)

	// FindFileShasum returns the 10-character, hex encoded, shasum of the given path
	FindFileShasum(path string) (shasum string, err error)

	// FindFileSha256 returns the given path's sha256 byte hash, base64 encoded
	FindFileSha256(path string) (shasum string, err error)

	// FindFileMime returns the mime type of the given path
	FindFileMime(path string) (mime string, err error)

	// ListFiles returns the list of files in the specified directory path
	ListFiles(path string) (files []string)

	// ListAllFiles returns the recursive list of all files at the specified directory path
	ListAllFiles(path string) (files []string)

	// ListDirs returns the list of directories in the specified directory path
	ListDirs(path string) (dirs []string)

	// ListAllDirs returns the recursive list of all files at the specified directory path
	ListAllDirs(path string) (dirs []string)
}

type Registry interface {
	// ID returns the identifier used when constructing the registry
	ID() (id string)

	// Register adds the given filesystems to the specified mount point
	Register(mount string, f ...FileSystem)

	// GetFileSystem returns the first FileSystem with the given identifier (not the same as ID)
	GetFileSystem(id string) (f FileSystem, ok bool)

	Lookup() (rl RegistryLookup)
}

type fsRegistry struct {
	id string

	known map[string]FileSystems

	sync.RWMutex
}

func NewRegistry(id string) (r Registry) {
	gRegistries.Lock()
	defer gRegistries.Unlock()
	if _, exists := gRegistries.known[id]; exists {
		log.FatalDF(1, "registry ID exists already")
	}
	gRegistries.order = append(gRegistries.order, id)
	gRegistries.known[id] = &fsRegistry{
		id:    id,
		known: make(map[string]FileSystems),
	}
	r = gRegistries.known[id]
	return
}

func (r *fsRegistry) ID() (id string) {
	id = r.id
	return
}

func (r *fsRegistry) Lookup() (rl RegistryLookup) {
	return r
}

func (r *fsRegistry) Register(mount string, f ...FileSystem) {
	r.Lock()
	defer r.Unlock()
	r.known[mount] = append(r.known[mount], f...)
}

func (r *fsRegistry) GetFileSystem(id string) (f FileSystem, ok bool) {
	r.RLock()
	defer r.RUnlock()
	for _, systems := range r.known {
		for _, system := range systems {
			if ok = system.ID() == id; ok {
				f = system
				return
			}
		}
	}
	return
}

func (r *fsRegistry) ReadFile(path string) (data []byte, err error) {
	r.RLock()
	defer r.RUnlock()
	for mount, systems := range r.known {
		p := clPath.TrimPrefix(path, mount)
		for _, f := range systems {
			if f.Exists(p) {
				data, err = f.ReadFile(p)
				return
			}
		}
	}
	return
}

func (r *fsRegistry) FileExists(path string) (exists bool) {
	r.RLock()
	defer r.RUnlock()
	for mount, systems := range r.known {
		p := clPath.TrimPrefix(path, mount)
		for _, f := range systems {
			if exists = f.Exists(p); exists {
				return
			}
		}
	}
	return
}

func (r *fsRegistry) FindFileShasum(path string) (shasum string, err error) {
	r.RLock()
	defer r.RUnlock()
	for mount, systems := range r.known {
		p := clPath.TrimPrefix(path, mount)
		for _, f := range systems {
			if shasum, err = f.Shasum(p); err == nil {
				return
			}
		}
	}
	err = os.ErrNotExist
	return
}

func (r *fsRegistry) FindFileSha256(path string) (shasum string, err error) {
	r.RLock()
	defer r.RUnlock()
	for mount, systems := range r.known {
		p := clPath.TrimPrefix(path, mount)
		for _, f := range systems {
			if shasum, err = f.Sha256(p); err == nil {
				return
			}
		}
	}
	err = os.ErrNotExist
	return
}

func (r *fsRegistry) FindFileMime(path string) (mime string, err error) {
	r.RLock()
	defer r.RUnlock()
	for _, systems := range r.known {
		for _, f := range systems {
			if mime, err = f.MimeType(path); err == nil {
				return
			}
		}
	}
	err = os.ErrNotExist
	return
}

func (r *fsRegistry) ListFiles(path string) (files []string) {
	r.RLock()
	defer r.RUnlock()
	for mount, systems := range r.known {
		p := clPath.TrimPrefix(path, mount)
		for _, f := range systems {
			if found, ee := f.ListFiles(p); ee == nil {
				files = append(files, found...)
			}
		}
	}
	return
}

func (r *fsRegistry) ListAllFiles(path string) (files []string) {
	r.RLock()
	defer r.RUnlock()
	for mount, systems := range r.known {
		p := clPath.TrimPrefix(path, mount)
		for _, f := range systems {
			if found, ee := f.ListAllFiles(p); ee == nil {
				files = append(files, found...)
			}
		}
	}
	return
}

func (r *fsRegistry) ListDirs(path string) (dirs []string) {
	r.RLock()
	defer r.RUnlock()
	for mount, systems := range r.known {
		p := clPath.TrimPrefix(path, mount)
		for _, f := range systems {
			if found, ee := f.ListDirs(p); ee == nil {
				dirs = append(dirs, found...)
			}
		}
	}
	return
}

func (r *fsRegistry) ListAllDirs(path string) (dirs []string) {
	r.RLock()
	defer r.RUnlock()
	for mount, systems := range r.known {
		p := clPath.TrimPrefix(path, mount)
		for _, f := range systems {
			if found, ee := f.ListAllDirs(p); ee == nil {
				dirs = append(dirs, found...)
			}
		}
	}
	return
}
