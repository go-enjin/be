//go:build (fs_theme && (drivers_fs_embed || drivers_fs || drivers || embeds)) || all

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

package themes

import (
	"embed"
	"fmt"
	"os"

	beFsEmbed "github.com/go-enjin/be/drivers/fs/embed"
	"github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/log"
)

type ThemeEmbedSupport interface {
	// EmbedTheme constructs an embedded theme.Theme instance and adds it to the
	// enjin during the build phase
	EmbedTheme(path string, tfs embed.FS) MakeFeature

	// EmbedThemes constructs all embedded theme.Theme instances and adds them
	// to the enjin during the build phase
	EmbedThemes(path string, fs embed.FS) MakeFeature
}

func (f *CFeature) loadEmbedTheme(path string, efs embed.FS) (err error) {
	var themeFs, staticFs fs.FileSystem
	if themeFs, err = beFsEmbed.New(f.Tag().String(), path, efs); err != nil {
		err = fmt.Errorf("error mounting local filesystem: %v - %v", path, err)
		return
	}
	if staticFs, err = beFsEmbed.New(f.Tag().String(), path+"/static", efs); err != nil {
		staticFs = nil
		err = nil
	}

	f.loading = append(f.loading, &loadTheme{
		support:  "embed",
		path:     path,
		themeFs:  themeFs,
		staticFs: staticFs,
	})

	return
}

func (f *CFeature) EmbedTheme(path string, tfs embed.FS) MakeFeature {

	if err := f.loadEmbedTheme(path, tfs); err != nil {
		log.FatalDF(1, "%v", err)
	}

	return f
}

func (f *CFeature) EmbedThemes(path string, tfs embed.FS) MakeFeature {
	var err error
	var entries []os.DirEntry
	if entries, err = tfs.ReadDir(path); err != nil {
		log.FatalF("error reading path: %v", err)
	}
	for _, info := range entries {
		if e := f.loadEmbedTheme(path+"/"+info.Name(), tfs); e != nil {
			log.FatalDF(1, "%s", err)
		}
	}
	return f
}