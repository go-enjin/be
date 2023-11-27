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
	"io/fs"
	"os"
	"path/filepath"

	times "github.com/go-enjin/github-com-djherbis-times"
)

// Abs is a convenience wrapper around filepath.Abs
func Abs(path string) (absolute string, err error) {
	absolute, err = filepath.Abs(path)
	return
}

// Dir is a convenience wrapper around filepath.Dir
func Dir(path string) (name string) {
	name = filepath.Dir(path)
	return
}

// Walk is a convenience wrapper around filepath.Walk
func Walk(root string, fn filepath.WalkFunc) (err error) {
	err = filepath.Walk(root, fn)
	return
}

// ReadDir is a convenience wrapper around os.ReadDir
func ReadDir(path string) (paths []fs.DirEntry, err error) {
	paths, err = os.ReadDir(path)
	return
}

// ReadFile is a convenience wrapper around os.ReadFile
func ReadFile(path string) (content []byte, err error) {
	content, err = os.ReadFile(path)
	return
}

// Stat is a convenience wrapper around github-com-djherbis-times.Stat
func Stat(path string) (spec times.Timespec, err error) {
	spec, err = times.Stat(path)
	return
}
