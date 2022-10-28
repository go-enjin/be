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
	"fmt"
	"sync"

	bePath "github.com/go-enjin/be/pkg/path"
)

type registry struct {
	registered map[string][]FileSystem

	sync.RWMutex
}

var _registry = &registry{
	registered: make(map[string][]FileSystem),
}

func RegisteredFileSystems() (registered map[string][]FileSystem) {
	_registry.RLock()
	defer _registry.RUnlock()
	registered = _registry.registered
	return
}

func RegisterFileSystem(mount string, f FileSystem) {
	_registry.Lock()
	defer _registry.Unlock()
	_registry.registered[mount] = append(_registry.registered[mount], f)
	// log.DebugDF(1, "registered fs: %v (%d)", mount, len(_registry.registered[mount]))
}

func FileExists(path string) (exists bool) {
	_registry.RLock()
	defer _registry.RUnlock()
	for mount, systems := range _registry.registered {
		p := bePath.TrimPrefix(path, mount)
		for _, f := range systems {
			// log.DebugF("checking for file existence: %v - %v", f.Name(), p)
			if exists = f.Exists(p); exists {
				return
			}
		}
	}
	return
}

func FindFileShasum(path string) (shasum string, err error) {
	_registry.RLock()
	defer _registry.RUnlock()
	for mount, systems := range _registry.registered {
		p := bePath.TrimPrefix(path, mount)
		for _, f := range systems {
			// log.DebugF("checking for file shasum: %v - %v", f.Name(), p)
			if shasum, err = f.Shasum(p); err == nil {
				return
			}
		}
	}
	err = fmt.Errorf("%v not found", path)
	return
}

func FindFileMime(path string) (mime string, err error) {
	_registry.RLock()
	defer _registry.RUnlock()
	for _, systems := range _registry.registered {
		for _, f := range systems {
			if mime, err = f.MimeType(path); err == nil {
				return
			}
		}
	}
	err = fmt.Errorf("%v not found", path)
	return
}