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

//go:build driver_fs_embed || drivers_fs || drivers || embeds || all

package catalog

import (
	"embed"

	"github.com/go-enjin/golang-org-x-text/language"

	beFsEmbed "github.com/go-enjin/be/drivers/fs/embed"
)

func (c *Catalog) IncludeEmbed(tag language.Tag, path string, efs embed.FS) {
	if bfs, err := beFsEmbed.New("lang-catalog", path, efs); err == nil {
		c.AddLocalesFromFS(tag, bfs)
	}
}