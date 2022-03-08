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
	"io/fs"
)

type FileSystem interface {
	Name() (name string)
	Open(path string) (fh fs.File, err error)
	ListDirs(path string) (paths []string, err error)
	ListFiles(path string) (paths []string, err error)
	ListAllDirs(path string) (paths []string, err error)
	ListAllFiles(path string) (paths []string, err error)
	ReadDir(path string) (paths []fs.DirEntry, err error)
	ReadFile(path string) (content []byte, err error)
	MimeType(path string) (mime string, err error)
	Shasum(path string) (shasum string, err error)
}