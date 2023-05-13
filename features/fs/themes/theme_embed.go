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
	"os"

	"github.com/go-enjin/be/pkg/log"
	bePath "github.com/go-enjin/be/pkg/path"
	"github.com/go-enjin/be/pkg/theme"
)

type EmbedSupport interface {
	// EmbedTheme constructs an embedded theme.Theme instance and adds it to the
	// enjin during the build phase
	EmbedTheme(name, path string, tfs embed.FS) MakeFeature

	// EmbedThemes constructs all embedded theme.Theme instances and adds them
	// to the enjin during the build phase
	EmbedThemes(path string, fs embed.FS) MakeFeature
}

func (f *CFeature) EmbedTheme(name, path string, tfs embed.FS) MakeFeature {
	var err error
	log.DebugDF(1, "embedding theme: %v", name)
	if f.themes[name], err = theme.NewEmbed(path, tfs); err != nil {
		delete(f.themes, name)
		log.FatalDF(1, "error embedding theme: %v", err)
	}
	return f
}

func (f *CFeature) EmbedThemes(path string, fs embed.FS) MakeFeature {
	var err error
	var entries []os.DirEntry
	if entries, err = fs.ReadDir(path); err != nil {
		log.FatalF("error reading path: %v", err)
	}
	for _, info := range entries {
		p := bePath.TrimSlashes(path + "/" + info.Name())
		name := bePath.Base(info.Name())
		f.EmbedTheme(name, p, fs)
	}
	return f
}