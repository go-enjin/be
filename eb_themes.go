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

package be

import (
	"embed"
	"os"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	bePath "github.com/go-enjin/be/pkg/path"
	"github.com/go-enjin/be/pkg/theme"
)

func (eb *EnjinBuilder) SetTheme(name string) feature.Builder {
	if _, ok := eb.theming[name]; ok {
		eb.theme = name
	} else {
		log.FatalF(`theme not found: "%v"`, name)
	}
	return eb
}

func (eb *EnjinBuilder) AddTheme(t *theme.Theme) feature.Builder {
	eb.theming[t.Name] = t
	if lfs, ok := t.Locales(); ok {
		eb.localeFiles = append(eb.localeFiles, lfs)
		log.DebugF("including %v theme locales", t.Name)
	}
	return eb
}

func (eb *EnjinBuilder) AddThemes(path string) feature.Builder {
	var err error
	var paths []string
	if paths, err = bePath.ListDirs(path); err != nil {
		log.FatalF("error listing path: %v", err)
		return nil
	}
	for _, p := range paths {
		name := bePath.Base(p)
		log.DebugF("loading theme: %v", p)
		if eb.theming[name], err = theme.NewLocal(p); err != nil {
			delete(eb.theming, name)
			log.FatalF("error loading theme: %v", err)
			return nil
		}
		if lfs, ok := eb.theming[name].Locales(); ok {
			eb.localeFiles = append(eb.localeFiles, lfs)
			log.DebugF("including %v theme locales", name)
		}
	}
	return eb
}

func (eb *EnjinBuilder) EmbedTheme(name, path string, tfs embed.FS) feature.Builder {
	var err error
	log.DebugF("embedding theme: %v", name)
	if eb.theming[name], err = theme.NewEmbed(path, tfs); err != nil {
		delete(eb.theming, name)
		log.FatalF("error embedding theme: %v", err)
		return nil
	}
	if lfs, ok := eb.theming[name].Locales(); ok {
		eb.localeFiles = append(eb.localeFiles, lfs)
		log.DebugF("including %v theme locales", name)
	}
	return eb
}

func (eb *EnjinBuilder) EmbedThemes(path string, fs embed.FS) feature.Builder {
	var err error
	var entries []os.DirEntry
	if entries, err = fs.ReadDir(path); err != nil {
		log.FatalF("error reading path: %v", err)
	}
	for _, info := range entries {
		p := bePath.TrimSlashes(path + "/" + info.Name())
		name := bePath.Base(info.Name())
		eb.EmbedTheme(name, p, fs)
	}
	return eb
}