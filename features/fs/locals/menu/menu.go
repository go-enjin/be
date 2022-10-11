//go:build local_menu || locals || all

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
	"fmt"
	"sort"

	"github.com/fvbommel/sortorder"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	beFs "github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/fs/local"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/menu"
	bePath "github.com/go-enjin/be/pkg/path"
)

var _localMenu *Feature

var _ feature.Feature = (*Feature)(nil)
var _ feature.MenuProvider = (*Feature)(nil)

const Tag feature.Tag = "LocalMenu"

type Feature struct {
	feature.CFeature

	setup   map[string]string
	mounted map[string]beFs.FileSystem
	menus   map[string]menu.Menu
}

type MakeFeature interface {
	feature.MakeFeature

	MountPath(mount, path string) MakeFeature
}

func New() MakeFeature {
	if _localMenu == nil {
		_localMenu = new(Feature)
		_localMenu.Init(_localMenu)
	}
	return _localMenu
}

func (f *Feature) MountPath(mount, path string) MakeFeature {
	f.setup[mount] = path
	return f
}

func (f *Feature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.setup = make(map[string]string)
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
		path := f.setup[mount]
		var lfs beFs.FileSystem
		if lfs, err = local.New(path); err != nil {
			log.FatalF(`error mounting filesystem: %v`, err)
			return nil
		}
		f.mounted[mount] = lfs
		beFs.RegisterFileSystem(mount, f.mounted[mount])
		log.DebugF("mounted local menu filesystem on %v to %v", mount, path)
	}
	return
}

func (f *Feature) Startup(ctx *cli.Context) (err error) {
	return f.Reload()
}

func (f *Feature) Reload() (err error) {
	f.menus = make(map[string]menu.Menu)
	for _, mount := range maps.SortedKeys(f.mounted) {
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
					log.DebugF("added menu %v from: (local) %v", name, filename)
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