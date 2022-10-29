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

	"github.com/go-enjin/golang-org-x-text/language"

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
	menus   map[language.Tag]map[string]menu.Menu
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

func (f *Feature) addMenuFiles(tag language.Tag, bfs beFs.FileSystem) (err error) {
	var rewrite func(tag language.Tag, in menu.Menu) (out menu.Menu)
	rewrite = func(tag language.Tag, in menu.Menu) (out menu.Menu) {
		for _, item := range in {
			if item.Lang == "" {
				item.Lang = tag.String()
			}
			// log.WarnF("rewrite: [%v] - %v - %v", tag, item.Href, item.Text)
			item.SubMenu = rewrite(tag, item.SubMenu)
			out = append(out, item)
		}
		return
	}

	log.DebugF("checking [%v] %v menu files", tag.String(), bfs.Name())
	var filenames []string
	if filenames, err = bfs.ListFiles("."); err != nil {
		err = fmt.Errorf("error listing files: [%v] %v", bfs.Name(), err)
		return
	}
	log.DebugF("found [%v] %v menu files: %v", tag.String(), bfs.Name(), filenames)
	sort.Sort(sortorder.Natural(filenames))
	for _, filename := range filenames {
		name := bePath.Base(filename)
		var data []byte
		if data, err = bfs.ReadFile(filename); err != nil {
			err = fmt.Errorf("error reading filesystem: %v - %v", bfs.Name(), err)
			return
		}
		var parsed menu.Menu
		if parsed, err = menu.NewMenuFromJson(data); err != nil {
			err = fmt.Errorf("error loading menu from file: [%v] %v - %v", bfs.Name(), filename, err)
			return
		} else {
			if _, ok := f.menus[tag]; !ok {
				f.menus[tag] = make(map[string]menu.Menu)
			}
			// rewrite(tag, &parsed)
			f.menus[tag][name] = rewrite(tag, parsed)
			// f.menus[tag][name] = parsed
			log.DebugF("added menu %v from: (local) %v", name, filename)
		}
	}
	return
}

func (f *Feature) Reload() (err error) {
	f.menus = make(map[language.Tag]map[string]menu.Menu)
	for _, mount := range maps.SortedKeys(f.mounted) {

		if ee := f.addMenuFiles(language.Und, f.mounted[mount]); ee != nil {
			log.ErrorF("error adding language.Und menu files: %v", ee)
		}

		if dirs, ee := f.mounted[mount].ListDirs("."); ee == nil {
			log.DebugF("found menu directories: %#v", dirs)
			for _, dir := range dirs {
				if tag, tpe := language.Parse(dir); tpe == nil {
					if dfs, eee := beFs.Wrap(dir, f.mounted[mount]); eee != nil {
						log.ErrorF("error wrapping menu directory: [%v] %v", dir, eee)
					} else if eeee := f.addMenuFiles(tag, dfs); eeee != nil {
						log.ErrorF("error adding menu directory: [%v] %v", dir, eeee)
					}
				}
			}
		}
	}
	return
}

func (f *Feature) GetMenus(tag language.Tag) (found map[string]menu.Menu) {
	found = make(map[string]menu.Menu)

	// undefined first so that actual lang can overwrite, leaving Und fallbacks
	if undMenus, ok := f.menus[language.Und]; ok {
		for _, name := range maps.SortedKeys(undMenus) {
			found[name] = undMenus[name]
			log.DebugF("found %v menu: %v", language.Und.String(), name)
		}
	}

	if localeMenus, ok := f.menus[tag]; ok {
		for _, name := range maps.SortedKeys(localeMenus) {
			found[name] = localeMenus[name]
			log.DebugF("found %v menu: %v", tag.String(), name)
		}
	}
	return
}
