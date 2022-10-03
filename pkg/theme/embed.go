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

package theme

import (
	"embed"

	beFs "github.com/go-enjin/be/pkg/fs"
	beFsEmbed "github.com/go-enjin/be/pkg/fs/embed"
	"github.com/go-enjin/be/pkg/log"
	bePath "github.com/go-enjin/be/pkg/path"
)

func NewEmbed(path string, efs embed.FS) (t *Theme, err error) {
	path = bePath.TrimSlashes(path)
	t = new(Theme)
	t.Path = path
	if t.FileSystem, err = beFsEmbed.New(path, efs); err != nil {
		return
	}
	if staticFs, e := beFsEmbed.Wrap(path, "static", efs); e == nil {
		beFs.RegisterFileSystem("/", staticFs)
		log.DebugF("wrapping static fs: %v/static", path)
	}
	err = t.init()
	return
}