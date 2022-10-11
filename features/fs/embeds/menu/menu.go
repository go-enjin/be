//go:build embed_menu || embeds || all

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

package menu

import (
	"embed"
	"fmt"
	"sort"

	"github.com/fvbommel/sortorder"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	beFs "github.com/go-enjin/be/pkg/fs"
	beFsEmbed "github.com/go-enjin/be/pkg/fs/embed"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/menu"
	bePath "github.com/go-enjin/be/pkg/path"
)

var _embedMenu *Feature

var _ feature.Feature = (*Feature)(nil)
var _ feature.MenuProvider = (*Feature)(nil)

const Tag feature.Tag = "EmbedMenu"

type Feature struct {
	feature.CFeature

	paths   map[string]string
	setup   map[string]embed.FS
	mounted map[string]beFs.FileSystem
	menus   map[string]menu.Menu
}

type MakeFeature interface {
	feature.MakeFeature

	MountPathFs(mount, path string, efs embed.FS) MakeFeature
}

func New() MakeFeature {
	if _embedMenu == nil {
		_embedMenu = new(Feature)
		_embedMenu.Init(_embedMenu)
	}
	return _embedMenu
}

func (f *Feature) MountPathFs(mount, path string, efs embed.FS) MakeFeature {
	f.paths[mount] = path
	f.setup[mount] = efs
	return f
}

func (f *Feature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.paths = make(map[string]string)
	f.setup = make(map[string]embed.FS)
	f.mounted = make(map[string]beFs.FileSystem)
}

func (f *Feature) Tag() (tag feature.Tag) {
	tag = Tag
	return
}

func (f *Feature) Build(_ feature.Buildable) (err error) {
	for _, mount := range maps.SortedKeys(f.setup) {
		if _, ok := f.mounted[mount]; ok {
			err = fmt.Errorf(`"%v" already mounted`, mount)
			return
		}
		if f.mounted[mount], err = beFsEmbed.New(f.paths[mount], f.setup[mount]); err != nil {
			log.FatalF(`error mounting filesystem: %v`, err)
			return nil
		}
		beFs.RegisterFileSystem(mount, f.mounted[mount])
		log.DebugF("mounted embed menu filesystem on %v to %v", mount, f.paths[mount])
	}
	return
}

func (f *Feature) Startup(ctx *cli.Context) (err error) {
	return f.Reload()
}

func (f *Feature) Reload() (err error) {
	f.menus = make(map[string]menu.Menu)
	for _, mount := range maps.SortedKeys[beFs.FileSystem](f.mounted) {
		var filenames []string
		if filenames, err = f.mounted[mount].ListAllFiles("/"); err != nil {
			err = fmt.Errorf("error listing filesystem: %v - %v", mount, err)
		} else {
			sort.Sort(sortorder.Natural(filenames))
			for _, filename := range filenames {
				name := bePath.Base(filename)
				var data []byte
				if data, err = f.mounted[mount].ReadFile(filename); err != nil {
					err = fmt.Errorf("error reading filesystem: %v - %v", mount, err)
					return
				}
				var parsed menu.Menu
				if parsed, err = menu.NewMenuFromJson(data); err != nil {
					err = fmt.Errorf("error loading menu from file: [%v] %v - %v", mount, filename, err)
					return
				} else {
					f.menus[name] = parsed
					log.DebugF("added menu %v from: (embed) %v", name, filename)
				}
			}
		}
	}
	return
}

func (f *Feature) GetMenu(name string) (parsed menu.Menu, ok bool) {
	parsed, ok = f.menus[name]
	return
}

func (f *Feature) GetMenus() (all map[string]menu.Menu) {
	all = make(map[string]menu.Menu)
	for _, name := range maps.SortedKeys(f.menus) {
		all[name] = f.menus[name]
	}
	return
}