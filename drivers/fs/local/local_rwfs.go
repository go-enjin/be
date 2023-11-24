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
	"time"

	"github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/types/page/matter"
)

// TODO: implement {Begin,End}Transaction(), flock-ing a top-level lock file blocking all other operations, this may be thread-safe

func (f *FileSystem) CloneRWFS() (cloned fs.RWFileSystem) {
	cloned = &FileSystem{
		origin: f.origin,
		root:   f.root,
		id:     f.id,
	}
	return
}

func (f *FileSystem) BeginTransaction() {

}

func (f *FileSystem) RollbackTransaction() {

}

func (f *FileSystem) CommitTransaction() {

}

func (f *FileSystem) EndTransaction() {

}

func (f *FileSystem) MakeDir(path string, perm os.FileMode) (err error) {
	f.Lock()
	defer f.Unlock()

	err = os.Mkdir(f.realpath(path), perm)
	return
}

func (f *FileSystem) MakeDirAll(path string, perm os.FileMode) (err error) {
	f.Lock()
	defer f.Unlock()

	err = os.MkdirAll(f.realpath(path), perm)
	return
}

func (f *FileSystem) Remove(path string) (err error) {
	f.Lock()
	defer f.Unlock()

	err = os.Remove(f.realpath(path))
	return
}

func (f *FileSystem) RemoveAll(path string) (err error) {
	f.Lock()
	defer f.Unlock()

	err = os.RemoveAll(f.realpath(path))
	return
}

func (f *FileSystem) WriteFile(path string, data []byte, perm os.FileMode) (err error) {
	f.Lock()
	defer f.Unlock()

	path = f.realpath(path)
	if err = f.ensurePathForWrite(path); err != nil {
		return
	}
	err = os.WriteFile(path, data, perm)
	return
}

func (f *FileSystem) ChangeTimes(path string, created, updated time.Time) (err error) {
	f.Lock()
	defer f.Unlock()
	// TODO: maybe ask Linus someday to encourage support for setting btime values
	path = f.realpath(path)
	err = os.Chtimes(path, updated, updated)
	return
}

func (f *FileSystem) WritePageMatter(pm *matter.PageMatter) (err error) {
	var data []byte
	if data, err = pm.Bytes(); err != nil {
		err = fmt.Errorf("error getting bytes from page matter: %v", err)
		return
	}
	err = f.WriteFile(pm.Path, data, DefaultFileMode)
	return
}

func (f *FileSystem) RemovePageMatter(path string) (err error) {
	err = f.Remove(path)
	return
}
