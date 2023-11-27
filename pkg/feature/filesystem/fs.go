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

package filesystem

import (
	"github.com/go-enjin/be/pkg/fs"
)

type Support[MakeTypedFeature interface{}] interface {
	// MountFS maps the local filesystem `path` to the enjin URL `point`
	//
	// The `point` is pruned from the URL during an HTTP request and the `path`
	// prefixes the file's real path. For example, it's common for the following
	// pattern:
	//
	//   f.MountLocalPath("/", "public")
	//
	// This configuration means to provide everything within the local path of
	// `./public/*` (recursively) at the root point of the URL, so for example
	// the URL `/favicon.ico` would translate to the local filesystem path of
	// `./public/favicon.ico`
	MountFS(point, path string, ifs fs.FileSystem) MakeTypedFeature
}

func (f *CFeature[MakeTypedFeature]) MountFS(mount, path string, ifs fs.FileSystem) MakeTypedFeature {
	if rwfs, ok := ifs.(fs.RWFileSystem); ok {
		f.MountPathRWFS(path, mount, rwfs)
	} else {
		f.MountPathROFS(path, mount, ifs)
	}
	v, _ := f.This().(MakeTypedFeature)
	return v
}
