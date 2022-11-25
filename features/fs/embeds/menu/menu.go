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

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/feature"
	beFs "github.com/go-enjin/be/pkg/fs"
	beFsEmbed "github.com/go-enjin/be/pkg/fs/embed"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/menu"
	bePath "github.com/go-enjin/be/pkg/path"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "EmbedMenu"

type Feature interface {
	feature.Feature
	feature.MenuProvider
}

type CFeature struct {
	feature.CFeature

	paths   map[string]string
	setup   map[string]embed.FS
	mounted map[string]beFsEmbed.FileSystem
	menus   map[language.Tag]map[string]menu.Menu
}

type MakeFeature interface {
	MountPathFs(mount, path string, efs embed.FS) MakeFeature

	Make() Feature
}

func New() MakeFeature {
	f := new(CFeature)
	f.Init(f)
	return f
}

func (f *CFeature) MountPathFs(mount, path string, efs embed.FS) MakeFeature {
	f.paths[path] = mount
	f.setup[path] = efs
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.paths = make(map[string]string)
	f.setup = make(map[string]embed.FS)
	f.mounted = make(map[string]beFsEmbed.FileSystem)
}

func (f *CFeature) Tag() (tag feature.Tag) {
	tag = Tag
	return
}

func (f *CFeature) Build(_ feature.Buildable) (err error) {
	for _, path := range maps.SortedKeys(f.setup) {
		if _, ok := f.mounted[path]; ok {
			err = fmt.Errorf(`"%v" already mounted`, path)
			return
		}
		if f.mounted[path], err = beFsEmbed.New(path, f.setup[path]); err != nil {
			log.FatalF(`error mounting filesystem: %v`, err)
			return nil
		}
		mount := f.paths[path]
		beFs.RegisterFileSystem(mount, f.mounted[path])
		log.DebugF("mounted embed menu filesystem %v to %v", path, mount)
	}
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	return f.Reload()
}

func (f *CFeature) addMenuFiles(tag language.Tag, bfs beFs.FileSystem) (err error) {
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
			log.DebugF("added menu %v from: (embed) %v", name, filename)
		}
	}
	return
}

func (f *CFeature) Reload() (err error) {
	f.menus = make(map[language.Tag]map[string]menu.Menu)
	for _, path := range maps.SortedKeys(f.mounted) {

		if ee := f.addMenuFiles(language.Und, f.mounted[path]); ee != nil {
			log.ErrorF("error adding language.Und menu files: %v", ee)
		}

		if dirs, ee := f.mounted[path].ListDirs("."); ee == nil {
			log.DebugF("found menu directories: %#v", dirs)
			for _, dir := range dirs {
				if tag, tpe := language.Parse(dir); tpe == nil {
					if dfs, eee := beFs.Wrap(dir, f.mounted[path]); eee != nil {
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

func (f *CFeature) GetMenus(tag language.Tag) (found map[string]menu.Menu) {
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