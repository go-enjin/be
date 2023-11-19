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
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/editor"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/feature/filesystem"
	uses_actions "github.com/go-enjin/be/pkg/feature/uses-actions"
	beFs "github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/net/serve"
	bePath "github.com/go-enjin/be/pkg/path"
)

var (
	DefaultCacheControl        = "public, max-age=604800, no-transform, immutable"
	DefaultVirtualCacheControl = "no-store"
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
	ServeBasePath(prefix, index string) MakeFeature
	SetCacheControl(values string) MakeFeature
	SetVirtualCacheControl(values string) MakeFeature
	SetMountCacheControl(mount string, value string) MakeFeature
	SetRegexCacheControl(pattern string, value string) MakeFeature

	Make() Feature
}

type CFeature struct {
	filesystem.CFeature[MakeFeature]
	uses_actions.CUsesActions

	dirIndex string

	cacheControl      string
	mountCacheControl map[string]string
	regexCacheControl map[string]string
	cachedRegexp      map[string]*regexp.Regexp

	basePaths map[string]string

	virtualPathCacheControl string

	uaf feature.Feature
	ubp feature.AuthProvider
	prh feature.PageRestrictionHandler
	drh feature.DataRestrictionHandler
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.CUsesActions.ConstructUsesActions(f)
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.CFeature.Localized = false
	f.mountCacheControl = make(map[string]string)
	f.regexCacheControl = make(map[string]string)
	f.cachedRegexp = make(map[string]*regexp.Regexp)
	f.basePaths = make(map[string]string)
}

func (f *CFeature) SetCacheControl(values string) MakeFeature {
	f.cacheControl = values
	return f
}

func (f *CFeature) SetVirtualCacheControl(values string) MakeFeature {
	f.virtualPathCacheControl = values
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

func (f *CFeature) ServeBasePath(prefix, index string) MakeFeature {
	prefix = bePath.CleanWithSlashes(prefix)
	f.basePaths[prefix] = bePath.CleanWithSlash(index)
	return f
}

func (f *CFeature) Make() Feature {
	for point, _ := range f.mountCacheControl {
		if _, found := f.MountPoints[point]; !found {
			log.FatalDF(1, "mount cache control mount-point not found: %v", point)
		}
	}
	for _, bp := range maps.SortedKeyLengths(f.basePaths) {
		if _, _, e := f.FindFile(f.basePaths[bp]); e != nil {
			log.FatalDF(1, "base path %v index file not found: %v", bp, f.basePaths[bp])
		}
	}
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	if f.virtualPathCacheControl == "" {
		f.virtualPathCacheControl = DefaultVirtualCacheControl
	}
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	err = f.CFeature.Startup(ctx)
	return
}

func (f *CFeature) Shutdown() {
	return
}

func (f *CFeature) UserActions() (list feature.Actions) {
	list = feature.Actions{
		f.Action("view", "public"),
	}
	return
}

func (f *CFeature) FindFile(path string) (data []byte, mime string, err error) {
	_, _, data, mime, err = f.findFileUnsafe(path)
	return
}

func (f *CFeature) ServePath(path string, s feature.System, w http.ResponseWriter, r *http.Request) (err error) {

	if v, ee := url.PathUnescape(path); ee == nil {
		path = v
	}
	path = bePath.CleanWithSlash(path)

	var data []byte
	var mime string
	var cmp *feature.CMountPoint
	var isVirtualBasePath bool

	for _, basePath := range maps.SortedKeyLengths(f.basePaths) {
		if path+"/" == basePath {
			// serve base path index file
			if cmp, _, data, mime, err = f.findFileUnsafe(f.basePaths[basePath]); err != nil {
				err = fmt.Errorf("error finding base path %v index file: %v - %v", basePath, f.basePaths[basePath], err)
				return
			} else {
				isVirtualBasePath = true
			}
		} else if strings.HasPrefix(path, basePath) {
			// check if it is an actual file
			if cmp, _, data, mime, err = f.findFileUnsafe(path); err != nil {
				err = nil
				// not a file, serve index without redirecting
				if cmp, _, data, mime, err = f.findFileUnsafe(f.basePaths[basePath]); err != nil {
					err = fmt.Errorf("error finding base path %v index file: %v - %v", basePath, f.basePaths[basePath], err)
					return
				} else {
					isVirtualBasePath = true
				}
			}
		}
	}

	if cmp == nil {
		if cmp, _, data, mime, err = f.findFileUnsafe(path); err == os.ErrNotExist {
			if f.dirIndex == "" {
				return
			}
			if cmp, _, data, mime, err = f.findFileUnsafe(path + "/" + f.dirIndex); err != nil {
				return
			}
		}
	}

	var cacheControlValue string
	if isVirtualBasePath {
		cacheControlValue = f.virtualPathCacheControl
	} else {
		for _, key := range maps.SortedKeys(f.regexCacheControl) {
			value := f.regexCacheControl[key]
			rx, _ := f.cachedRegexp[key]
			if rx.MatchString(path) {
				cacheControlValue = value
				break
			}
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

	if cacheControlValue != "" {
		r = serve.SetCacheControl(cacheControlValue, w, r)
	}
	s.ServeData(data, mime, w, r)
	return
}

func (f *CFeature) findFileUnsafe(path string) (cmp *feature.CMountPoint, realpath string, data []byte, mime string, err error) {
	if v, ee := url.PathUnescape(path); ee == nil {
		path = v
	}

	if _, _, ok := editor.ParseEditorWorkFile(path); ok {
		err = os.ErrNotExist
		return
	}

	if len(path) <= 1 {
		err = os.ErrNotExist
		return
	}

	var ok bool
	for _, mp := range f.FindPathMountPoint(path) {
		if data, mime, _, ok = beFs.CheckForFileData(mp.ROFS, path, mp.Mount); ok {
			cmp = mp
			return
		}
	}

	cmp = nil
	realpath = ""
	data = nil
	mime = ""

	err = os.ErrNotExist
	return
}