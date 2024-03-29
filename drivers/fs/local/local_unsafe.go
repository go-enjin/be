//go:build driver_fs_local || drivers_fs || drivers || locals || all

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

package local

import (
	"fmt"
	"os"
	"path/filepath"

	clPath "github.com/go-corelibs/path"
	"github.com/go-corelibs/slices"
)

func (f *FileSystem) realpath(path string) (out string) {
	out = clPath.SafeConcatRelPath(f.root, path)
	return
}

func (f *FileSystem) ensurePathForWrite(path string) (err error) {
	if dir := filepath.Dir(path); !slices.Present(dir, "", ".", "/", "./") {
		if err = os.MkdirAll(dir, DefaultDirMode); err != nil {
			err = fmt.Errorf("error making directory: %v", dir)
			return
		}
	}
	return
}
