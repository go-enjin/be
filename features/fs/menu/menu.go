//go:build fs_menu || all

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
	"strings"

	"github.com/maruel/natural"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/feature/filesystem"
	"github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/menu"
	bePath "github.com/go-enjin/be/pkg/path"
)

const Tag feature.Tag = "fs-menu"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type Feature interface {
	filesystem.Feature[MakeFeature]
	feature.MenuProvider
}

type MakeFeature interface {
	filesystem.MakeFeature[MakeFeature]

	Make() Feature
}

type CFeature struct {
	filesystem.CFeature[MakeFeature]
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	err = f.CFeature.Startup(ctx)
	return
}

func (f *CFeature) Shutdown() {
	return
}

func (f *CFeature) GetMenus(tag language.Tag) (found map[string]menu.Menu) {
	found = make(map[string]menu.Menu)

	for _, point := range maps.SortedKeys(f.MountPoints) {
		for _, mp := range f.MountPoints[point] {
			if foundMenus, err := f.findAllMenus(tag, mp.ROFS); err == nil {
				for name, menus := range foundMenus {
					found[name] = append(found[name], menus...)
				}
			} else {
				log.ErrorF("error finding all menus: [%v] %v - %v", tag, point, err)
			}
		}
	}

	return
}

func (f *CFeature) GetAllMenus() (menus map[language.Tag]map[string]menu.Menu) {
	menus = make(map[language.Tag]map[string]menu.Menu)
	tags := f.Enjin.SiteLocales()

	for _, tag := range tags {
		menus[tag] = make(map[string]menu.Menu)
	}

	for _, point := range maps.SortedKeys(f.MountPoints) {
		for _, mp := range f.MountPoints[point] {
			for _, tag := range tags {
				if foundMenus, err := f.findAllMenus(tag, mp.ROFS); err == nil {
					for name, found := range foundMenus {
						menus[tag][name] = found
					}
				} else {
					log.ErrorF("error finding all menus: [%v] %v - %v", tag, point, err)
				}
			}
		}
	}

	return
}

func (f *CFeature) findAllMenus(tag language.Tag, bfs fs.FileSystem) (menus map[string]menu.Menu, err error) {
	menus = make(map[string]menu.Menu)

	var updateItemLangs func(tag language.Tag, in menu.Menu) (out menu.Menu)
	updateItemLangs = func(tag language.Tag, in menu.Menu) (out menu.Menu) {
		for _, item := range in {
			if item.Lang == "" {
				item.Lang = tag.String()
			}
			// log.WarnF("rewrite: [%v] - %v - %v", tag, item.Href, item.Text)
			item.SubMenu = updateItemLangs(tag, item.SubMenu)
			out = append(out, item)
		}
		return
	}

	log.TraceF("checking [%v] %v menu files", tag.String(), bfs.Name())
	var filenames []string
	if filenames, err = bfs.ListAllFiles("/"); err != nil {
		err = fmt.Errorf("error listing files: [%v] %v", bfs.Name(), err)
		return
	}

	log.TraceF("found [%v] %v menu files: %v", tag.String(), bfs.Name(), filenames)
	sort.Sort(natural.StringSlice(filenames))
	for _, filename := range filenames {
		if !strings.HasSuffix(filename, ".json") {
			log.WarnF("not a <menu-name>.json filename: %v", filename)
			continue
		}

		fileLang := language.Und
		if foundLang, _, ok := lang.ParseLangPath(filename); ok {
			if foundLang != tag {
				log.TraceF("expected %v, received %v", tag, foundLang)
				continue
			}
			fileLang = foundLang
			log.TraceF("parsed %v lang from: %v", foundLang, filename)
		}

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
		}
		var pruned menu.Menu
		for _, item := range parsed {
			if fileLang == language.Und {
				if item.Lang != "" && item.Lang != tag.String() {
					continue
				}
			} else if fileLang != tag {
				continue
			}
			pruned = append(pruned, item)
		}
		menus[name] = updateItemLangs(tag, pruned)
		log.TraceF("added menu %v from: (%v) %v", name, f.Tag(), filename)
	}
	return
}