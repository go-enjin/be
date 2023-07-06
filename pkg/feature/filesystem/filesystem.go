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

package filesystem

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-enjin/golang-org-x-text/language"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/page/matter"
	bePath "github.com/go-enjin/be/pkg/path"
)

var (
	_ Feature[MakeFeature[interface{}]] = &CFeature[MakeFeature[interface{}]]{}
)

type Feature[MakeTypedFeature interface{}] interface {
	feature.Feature

	/*
		MountPath methods for use during enjin Make phase
	*/

	// MountPathROFS mounts the given fs.FileSystem path to the virtual URI point
	MountPathROFS(path, point string, rofs fs.FileSystem)

	// MountPathRWFS mounts the given fs.RWFileSystem path to the virtual URI point
	MountPathRWFS(path, point string, rwfs fs.RWFileSystem)

	/*
		Runtime methods for reading and writing matter.PageMatter fs.FileSystem
		content, supporting local path prefixes.
	*/

	// FindPathMountPoint returns the available mounts matching the given path.
	// For example if there are mounted points of "/stuff" and "/", calling:
	//
	//   f.FindPathMountPoint("/thing")
	//
	// would return the "/" mounts and calling:
	//
	//   f.FindPathMountPoint("/stuff/thing")
	//
	// would return the "/stuff" mounts
	FindPathMountPoint(path string) (mps []*CMountPoint)

	// Exists returns true if the given URI path is present on any of this
	// feature instance's mounted filesystems (can be a file or directory)
	Exists(path string) (present bool)

	// FindPageMatterPath looks for the actual path by checking for the prefix
	// with each of the enjin provided page format extensions, this allows for
	// finding PageMatter without knowing the page's actual format or language
	// first. Uses FindPathMountPoint on the prefix to find the correct
	// filesystem
	FindPageMatterPath(prefix string) (path string, err error)

	// FindReadPageMatter parses the given path file data into a matter.PageMatter
	// structure, suitable for further processing as page.Page or any other
	// type. The first fs.FileSystem which has the given path is used. The order
	// checked is the order features are added during the main enjin build
	// phase. ReadPageMatter uses FindPageMatterPath to find the matter for
	// reading
	FindReadPageMatter(path string) (pm *matter.PageMatter, err error)

	// ReadPageMatter is like FindReadPageMatter except that it uses the path
	// without using FindPageMatterPath
	ReadPageMatter(path string) (pm *matter.PageMatter, err error)

	// WritePageMatter constructs new file data from the existing pm.Body with
	// pm.Matter (using pm.FrontMatterType) and writes it to the first
	// fs.RWFileSystem that matches the PageMatter's path
	WritePageMatter(pm *matter.PageMatter) (err error)
}

type MakeFeature[MakeTypedFeature interface{}] interface {
	LocalPathSupport[MakeTypedFeature]
	EmbedPathSupport[MakeTypedFeature]
	ZipPathSupport[MakeTypedFeature]
	GormDBPathSupport[MakeTypedFeature]
}

type CFeature[MakeTypedFeature interface{}] struct {
	feature.CFeature
	CGormDBPathSupport[MakeTypedFeature]

	MountPoints map[string][]*CMountPoint
}

type CMountPoint struct {
	// Path is the actual filesystem path
	Path string
	// Mount is the URL path prefix
	Mount string
	// ROFS is the read-only filesystem, always non-nil
	ROFS fs.FileSystem
	// RWFS is the write-only filesystem, nil when fs is read-only
	RWFS fs.RWFileSystem
}

func (f *CFeature[MakeTypedFeature]) Init(this interface{}) {
	f.CFeature.Init(this)
	f.CGormDBPathSupport.initGormDBPathSupport(f)
	f.FeatureTag = feature.NotImplemented
	f.MountPoints = make(map[string][]*CMountPoint)
}

func (f *CFeature[MakeTypedFeature]) MountPathROFS(path, point string, rofs fs.FileSystem) {
	f.MountPoints[point] = append(f.MountPoints[point], &CMountPoint{
		Path:  path,
		Mount: point,
		ROFS:  rofs,
		RWFS:  nil,
	})
	return
}

func (f *CFeature[MakeTypedFeature]) MountPathRWFS(path, point string, rwfs fs.RWFileSystem) {
	f.MountPoints[point] = append(f.MountPoints[point], &CMountPoint{
		Path:  path,
		Mount: point,
		ROFS:  rwfs,
		RWFS:  rwfs,
	})
	return
}

func (f *CFeature[MakeTypedFeature]) Setup(enjin feature.Internals) {
	f.CFeature.Setup(enjin)
	for _, point := range maps.SortedKeyLengths(f.MountPoints) {
		for _, mp := range f.MountPoints[point] {
			fs.RegisterFileSystem(point, mp.ROFS)
		}
	}
}

func (f *CFeature[MakeTypedFeature]) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	if err = f.CGormDBPathSupport.startupGormDBPathSupport(f, ctx); err != nil {
		return
	}
	return
}

func (f *CFeature[MakeTypedFeature]) FindPathMountPoint(path string) (mps []*CMountPoint) {
	path = bePath.CleanWithSlash(path)
	// log.WarnDF(1, "finding path mount for: %v", path)
	for _, point := range maps.SortedKeyLengths(f.MountPoints) {
		if path == point || strings.HasPrefix(path, bePath.CleanWithSlashes(point)) {
			// log.WarnDF(1, "%v has %v prefix", path, point)
			mps = append(mps, f.MountPoints[point]...)
		}
		// log.WarnDF(1, "%v does not have %v prefix", path, point)
	}
	return
}

func (f *CFeature[MakeTypedFeature]) Exists(path string) (present bool) {
	var ok bool
	var uri, modified string

	tag := language.Und
	defLang := f.Enjin.SiteDefaultLanguage()

	uri = bePath.CleanWithSlash(path)
	if tag, modified, ok = lang.ParseLangPath(uri); ok {
		uri = modified
	}

	undSrc := bePath.CleanWithSlash(uri)

	switch {

	case tag != language.Und:
		// check for specific language
		undSrc = "/" + tag.String() + undSrc
		for _, mp := range f.FindPathMountPoint(undSrc) {
			if present = mp.ROFS.Exists(undSrc); present {
				return
			}
		}
		fallthrough

	default:
		// check for default language
		defSrc := "/" + defLang.String() + undSrc
		for _, mp := range f.FindPathMountPoint(undSrc) {
			if present = mp.ROFS.Exists(defSrc); present {
				return
			}
		}
		// check for undefined language
		for _, mp := range f.FindPathMountPoint(undSrc) {
			if present = mp.ROFS.Exists(undSrc); present {
				return
			}
		}
	}

	return
}

func (f *CFeature[MakeTypedFeature]) FindPathsWithContextKey(path, key string) (found []string, err error) {
	for _, mp := range f.FindPathMountPoint(path) {
		if v, ok := mp.RWFS.(fs.QueryFileSystem); ok {
			if more, ee := v.FindPathsWithContextKey(path, key); ee == nil {
				found = append(found, more...)
			}
		}
	}
	if len(found) == 0 {
		err = os.ErrNotExist
	}
	return
}

func (f *CFeature[MakeTypedFeature]) FindPathsWhereContextKeyEquals(path, key string, value interface{}) (found []string, err error) {
	for _, mp := range f.FindPathMountPoint(path) {
		if v, ok := mp.RWFS.(fs.QueryFileSystem); ok {
			if more, ee := v.FindPathsWhereContextKeyEquals(path, key, value); ee == nil {
				found = append(found, more...)
			}
		}
	}
	if len(found) == 0 {
		err = os.ErrNotExist
	}
	return
}

func (f *CFeature[MakeTypedFeature]) FindPathsWhereContextEquals(path string, conditions map[string]interface{}) (found []string, err error) {
	for _, mp := range f.FindPathMountPoint(path) {
		if v, ok := mp.RWFS.(fs.QueryFileSystem); ok {
			if more, ee := v.FindPathsWhereContextEquals(path, conditions); ee == nil {
				found = append(found, more...)
			}
		}
	}
	if len(found) == 0 {
		err = os.ErrNotExist
	}
	return
}

func (f *CFeature[MakeTypedFeature]) FindPathsWhereContext(path string, orJsonConditions ...map[string]interface{}) (found []string, err error) {
	for _, mp := range f.FindPathMountPoint(path) {
		if v, ok := mp.RWFS.(fs.QueryFileSystem); ok {
			if more, ee := v.FindPathsWhereContext(path, orJsonConditions...); ee == nil {
				found = append(found, more...)
			}
		}
	}
	if len(found) == 0 {
		err = os.ErrNotExist
	}
	return
}

func (f *CFeature[MakeTypedFeature]) FindPageMatterPath(path string) (realpath string, err error) {
	realpath, _, _, err = f.findPageMatterPathMount(path)
	return
}

func (f *CFeature[MakeTypedFeature]) FindReadPageMatter(path string) (pm *matter.PageMatter, err error) {
	var realpath string
	var mp *CMountPoint
	var locale language.Tag
	if realpath, mp, locale, err = f.findPageMatterPathMount(path); err != nil {
		if err != os.ErrNotExist {
			err = fmt.Errorf("error finding page matter path: %v - %v", path, err)
		}
		return
	}
	if mp == nil {
		panic("nil mp")
	}
	if pm, err = mp.ROFS.ReadPageMatter(realpath); err == nil {
		pm.Locale = locale
		pm.Stub = fs.NewPageStub(f.Tag().String(), f.Enjin.Context(), mp.ROFS, mp.Mount, realpath, pm.Shasum, locale)
		return
	}
	err = os.ErrNotExist
	return
}

func (f *CFeature[MakeTypedFeature]) ReadPageMatter(path string) (pm *matter.PageMatter, err error) {
	for _, mp := range f.FindPathMountPoint(path) {
		if pm, err = mp.ROFS.ReadPageMatter(path); err == nil {
			pm.Stub = fs.NewPageStub(f.Tag().String(), f.Enjin.Context(), mp.ROFS, mp.Mount, path, pm.Shasum, pm.Locale)
			return
		}
	}
	err = os.ErrNotExist
	return
}

func (f *CFeature[MakeTypedFeature]) WritePageMatter(pm *matter.PageMatter) (err error) {
	for _, mp := range f.FindPathMountPoint(pm.Path) {
		if mp.RWFS != nil {
			err = mp.RWFS.WritePageMatter(pm)
			return
		}
	}
	err = fmt.Errorf("read/write mount point for [%v] not found", pm.Path)
	return
}

func (f *CFeature[MakeTypedFeature]) RemovePageMatter(path string) (err error) {
	for _, mp := range f.FindPathMountPoint(path) {
		if mp.RWFS != nil {
			err = mp.RWFS.RemovePageMatter(path)
			return
		}
	}
	err = fmt.Errorf("read/write mount point for [%v] not found", path)
	return
}