//go:build driver_fs_embed || drivers_fs || embeds || all

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
	"embed"

	beFsEmbed "github.com/go-enjin/be/drivers/fs/embed"
	"github.com/go-enjin/be/pkg/log"
)

type EmbedPathSupport[MakeTypedFeature interface{}] interface {
	// MountEmbedPath maps the embed filesystem `path` to the enjin URL `point`
	//
	// The `point` is pruned from the URL during an HTTP request and the `path`
	// prefixes the file's real path. For example, it's common for the following
	// pattern:
	//
	//   //go:embed public/**
	//   var publicFS embed.FS
	//
	//   f.MountEmbedPath("/", "public", publicFS)
	//
	// This configuration means to provide everything within the embed path of
	// `./public/*` (recursively) at the root point of the URL, so for example
	// the URL `/favicon.ico` would translate to the embed filesystem path of
	// `./public/favicon.ico`
	MountEmbedPath(mount, path string, rofs embed.FS) MakeTypedFeature
}

func (f *CFeature[MakeTypedFeature]) MountEmbedPath(mount, path string, efs embed.FS) MakeTypedFeature {
	if lfs, err := beFsEmbed.New(f.Tag().String(), path, efs); err != nil {
		log.FatalDF(1, "error mounting path: %v", err)
	} else {
		f.MountPathROFS(path, mount, lfs)
	}
	v, _ := f.This().(MakeTypedFeature)
	return v
}