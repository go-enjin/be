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

	"github.com/go-enjin/be/pkg/page/matter"
)

type WrapRWFileSystem struct {
	WrapFileSystem
	rw RWFileSystem
}

func WrapRW(path string, fs RWFileSystem) (out RWFileSystem, err error) {
	out = WrapRWFileSystem{
		WrapFileSystem: WrapFileSystem{
			path: path,
			fs:   fs,
		},
		rw: fs,
	}
	return
}

func (w WrapRWFileSystem) MakeDir(path string, perm os.FileMode) (err error) {
	err = w.rw.MakeDir(w.realpath(path), perm)
	return
}

func (w WrapRWFileSystem) MakeDirAll(path string, perm os.FileMode) (err error) {
	err = w.rw.MakeDirAll(w.realpath(path), perm)
	return
}

func (w WrapRWFileSystem) WriteFile(path string, data []byte, perm os.FileMode) (err error) {
	err = w.rw.WriteFile(w.realpath(path), data, perm)
	return
}

func (w WrapRWFileSystem) Remove(path string) (err error) {
	err = w.rw.Remove(w.realpath(path))
	return
}

func (w WrapRWFileSystem) RemoveAll(path string) (err error) {
	err = w.rw.RemoveAll(w.realpath(path))
	return
}

func (w WrapRWFileSystem) WritePageMatter(pm *matter.PageMatter) (err error) {
	err = w.rw.WritePageMatter(pm)
	return
}

func (w WrapRWFileSystem) RemovePageMatter(path string) (err error) {
	err = w.rw.RemovePageMatter(path)
	return
}