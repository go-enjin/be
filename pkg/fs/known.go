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

	bePath "github.com/go-enjin/be/pkg/path"
)

var _registered = make(map[string][]FileSystem)

func RegisterFileSystem(mount string, f FileSystem) {
	if _, ok := _registered[mount]; !ok {
		_registered[mount] = make([]FileSystem, 0)
	}
	_registered[mount] = append(_registered[mount], f)
}

func FindFileShasum(path string) (shasum string, err error) {
	for mount, systems := range _registered {
		p := bePath.TrimPrefix(path, mount)
		for _, f := range systems {
			if shasum, err = f.Shasum(p); err == nil {
				return
			}
		}
	}
	err = fmt.Errorf("%v not found", path)
	return
}

func FindFileMime(path string) (mime string, err error) {
	for _, systems := range _registered {
		for _, f := range systems {
			if mime, err = f.MimeType(path); err == nil {
				return
			}
		}
	}
	err = fmt.Errorf("%v not found", path)
	return
}