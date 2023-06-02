//go:build fs_public || all

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

package public

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/feature/filesystem"
	beFs "github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	bePath "github.com/go-enjin/be/pkg/path"
	"github.com/go-enjin/be/pkg/userbase"
)

var (
	DefaultCacheControl = "public, max-age=604800, no-transform, immutable"
)

const Tag feature.Tag = "fs-public"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type Feature interface {
	filesystem.Feature[MakeFeature]
}

type MakeFeature interface {
	filesystem.MakeFeature[MakeFeature]

	UseDirIndex(indexFileName string) MakeFeature
	SetCacheControl(values string) MakeFeature
	SetMountCacheControl(mount string, value string) MakeFeature
	SetRegexCacheControl(pattern string, value string) MakeFeature

	Make() Feature
}

type CFeature struct {
	filesystem.CFeature[MakeFeature]

	dirIndex string

	cacheControl      string
	mountCacheControl map[string]string
	regexCacheControl map[string]string
	cachedRegexp      map[string]*regexp.Regexp

	uaf feature.Feature
	ubp userbase.AuthProvider
	prh feature.PageRestrictionHandler
	drh feature.DataRestrictionHandler
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.FeatureTag = tag
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.mountCacheControl = make(map[string]string)
	f.regexCacheControl = make(map[string]string)
	f.cachedRegexp = make(map[string]*regexp.Regexp)
}

func (f *CFeature) SetCacheControl(values string) MakeFeature {
	f.cacheControl = values
	return f
}

func (f *CFeature) SetMountCacheControl(mount string, value string) MakeFeature {
	f.mountCacheControl[mount] = value
	return f
}

func (f *CFeature) SetRegexCacheControl(pattern string, value string) MakeFeature {
	if compiled, err := regexp.Compile(pattern); err != nil {
		log.FatalF("error compiling regex cache-control pattern: %v - %v", pattern, err)
	} else {
		f.cachedRegexp[pattern] = compiled
		f.regexCacheControl[pattern] = value
	}
	return f
}

func (f *CFeature) UseDirIndex(indexFileName string) MakeFeature {
	f.dirIndex = filepath.Base(indexFileName)
	return f
}


func (f *CFeature) Make() Feature {
	for point, _ := range f.mountCacheControl {
		if _, found := f.MountPoints[point]; !found {
			log.FatalDF(1, "mount cache control mount-point not found: %v", point)
		}
	}
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

func (f *CFeature) UserActions() (list userbase.Actions) {

	tag := f.Tag().Kebab()
	list = userbase.Actions{
		userbase.NewAction(tag, "view", "public"),
	}

	return
}

func (f *CFeature) FindFile(path string) (data []byte, mime string, err error) {
	if v, ee := url.PathUnescape(path); ee == nil {
		path = v
	}

	if len(path) <= 1 {
		err = os.ErrNotExist
		return
	}

	var ok bool
	for _, point := range maps.SortedKeys(f.MountPoints) {
		for _, mp := range f.MountPoints[point] {
			if data, mime, _, ok = beFs.CheckForFileData(mp.ROFS, path, mp.Mount); ok {
				return
			}
		}
	}

	err = os.ErrNotExist
	return
}

func (f *CFeature) ServePath(path string, s feature.System, w http.ResponseWriter, r *http.Request) (err error) {

	if v, ee := url.PathUnescape(path); ee == nil {
		path = v
	}
	path = bePath.CleanWithSlash(path)

	var cmp *filesystem.CMountPoint
	var data []byte
	var mime string
	var ok bool

	// check for path as-is
	if cmp, data, mime, _, ok = f.findFileData(path); !ok {
		// path as-is not found, have dirIndex configured
		if f.dirIndex == "" {
			err = os.ErrNotExist
			return
		}
		path += "/" + f.dirIndex
		if cmp, data, mime, _, ok = f.findFileData(path); !ok {
			err = os.ErrNotExist
			return
		}
	}

	// if f.drh != nil {
	// 	if modReq, pass := f.drh.RestrictServeData(data, mime, w, r); !pass {
	// 		return
	// 	} else {
	// 		r = modReq
	// 	}
	// }

	var cacheControlValue string
	for _, key := range maps.SortedKeys(f.regexCacheControl) {
		value := f.regexCacheControl[key]
		rx, _ := f.cachedRegexp[key]
		if rx.MatchString(path) {
			cacheControlValue = value
			break
		}
	}
	if cacheControlValue == "" {
		if values, found := f.mountCacheControl[cmp.Mount]; found {
			cacheControlValue = values
		} else if f.cacheControl != "" {
			cacheControlValue = f.cacheControl
		} else {
			cacheControlValue = DefaultCacheControl
		}
	}

	w.Header().Set("Cache-Control", cacheControlValue)
	s.ServeData(data, mime, w, r)
	// log.DebugRDF(r, 1, "%s feature served file: [%v] %v (%v)", f.Tag(), cmp.Mount, fullpath, mime)
	return
}

func (f *CFeature) findFileData(path string) (cmp *filesystem.CMountPoint, data []byte, mime string, fullpath string, ok bool) {
	for _, point := range maps.SortedKeys(f.MountPoints) {
		for _, mp := range f.MountPoints[point] {
			if data, mime, fullpath, ok = beFs.CheckForFileData(mp.ROFS, path, mp.Mount); ok {
				cmp = mp
				return
			}
		}
	}
	return
}