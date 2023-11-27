//go:build driver_fs_zip || drivers_fs || zips || all

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
	"github.com/spkg/zipfs"

	"github.com/go-enjin/be/drivers/fs/zip"
	"github.com/go-enjin/be/pkg/log"
)

type ZipPathSupport[MakeTypedFeature interface{}] interface {
	// MountZipPath maps the embed filesystem `path` to the enjin URL `point`
	//
	// The `point` is pruned from the URL during an HTTP request and the `path`
	// prefixes the file's real path. For example, it's common for the following
	// pattern:
	//
	//   /* prepare zip of local public directory; in this example, the zip file
	//      does in fact contain a top-level directory of "public" which is
	//      important for the f.MountZipPath call
	//   */
	//   $ zip -r public.zip public
	//
	//   zipFS, err := zipfs.New("./public.zip")
	//   f.MountZipPath("/", "public", zipFS)
	//
	// This configuration means to provide everything within the zip file of
	// `./public/*` (recursively) at the root point of the URL, so for example
	// the URL `/favicon.ico` would translate to the embed filesystem path of
	// `./public/favicon.ico` within the zip file
	MountZipPath(mount, path string, zfs *zipfs.FileSystem) MakeTypedFeature
}

func (f *CFeature[MakeTypedFeature]) MountZipPath(mount, path string, zfs *zipfs.FileSystem) MakeTypedFeature {
	if lfs, err := zip.New(f.Tag().String(), path, zfs); err != nil {
		log.FatalDF(1, "error mounting path: %v", err)
	} else {
		f.MountPathROFS(path, mount, lfs)
	}
	v, _ := f.This().(MakeTypedFeature)
	return v
}
