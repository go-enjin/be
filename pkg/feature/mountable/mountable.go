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

package mountable

import (
	"fmt"
	"strings"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
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

	// Exists returns true if the given URI path is present on any of this
	// feature instance's mounted filesystems (can be a file or directory)
	Exists(path string) (present bool)

	// ReadPageMatter parses the given path file data into a matter.PageMatter
	// structure, suitable for further processing as page.Page or any other
	// type. The first fs.FileSystem which has the given path is used. The order
	// checked is the order features are added during the main enjin build phase
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
}

type CFeature[MakeTypedFeature interface{}] struct {
	feature.CFeature

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
	for _, point := range maps.SortedKeys(f.MountPoints) {
		for _, mp := range f.MountPoints[point] {
			fs.RegisterFileSystem(point, mp.ROFS)
		}
	}
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

	for _, point := range maps.SortedKeys(f.MountPoints) {
		if strings.HasPrefix(uri, point) {
			undSrc := bePath.CleanWithSlash(uri)

			for _, mp := range f.MountPoints[point] {
				switch {

				case tag != language.Und:
					undSrc = "/" + tag.String() + undSrc
					if present = mp.ROFS.Exists(undSrc); present {
						return
					}
					fallthrough

				default:
					defSrc := "/" + defLang.String() + undSrc
					if present = mp.ROFS.Exists(defSrc); present {
						return
					}
					if present = mp.ROFS.Exists(undSrc); present {
						return
					}
				}
			}
		}
	}
	return
}

func (f *CFeature[MakeTypedFeature]) ReadPageMatter(path string) (pm *matter.PageMatter, err error) {
	var ok bool
	var uri, modified string

	tag := language.Und
	defLang := f.Enjin.SiteDefaultLanguage()

	uri = bePath.CleanWithSlash(path)
	if tag, modified, ok = lang.ParseLangPath(uri); ok {
		uri = modified
	}

	processReadPageMatter := func(locale language.Tag, src string, mp *CMountPoint) (pm *matter.PageMatter, err error) {
		if mp.ROFS.Exists(src) {
			var data []byte
			if data, err = mp.ROFS.ReadFile(src); err != nil {
				return
			}
			_, _, created, updated, _ := mp.ROFS.FileStats(src)
			pm, err = matter.ParsePageMatter(src, created, updated, data)
			pm.Locale = locale
			pm.Stub, err = matter.NewPageStub(f.Enjin.Context(), mp.ROFS, mp.Mount, src, pm.Shasum, locale)
			log.TraceF("made page matter and stub for: [%v] %v", tag, uri)
			return
		}
		err = fmt.Errorf("not found")
		return
	}

	for _, point := range maps.SortedKeys(f.MountPoints) {
		if strings.HasPrefix(uri, point) {
			undSrc := bePath.CleanWithSlash(uri)

			for _, mp := range f.MountPoints[point] {
				switch {

				case tag != language.Und:
					tagSrc := "/" + tag.String() + undSrc
					if pm, err = processReadPageMatter(tag, tagSrc, mp); err == nil {
						return
					}
					fallthrough

				default:
					if pm, err = processReadPageMatter(defLang, undSrc, mp); err == nil {
						return
					}
					if pm, err = processReadPageMatter(language.Und, undSrc, mp); err == nil {
						return
					}

				}
			}
		}
	}
	return
}

func (f *CFeature[MakeTypedFeature]) WritePageMatter(pm *matter.PageMatter) (err error) {
	for _, point := range maps.SortedKeys(f.MountPoints) {
		if strings.HasPrefix(pm.Path, point) {
			for _, mp := range f.MountPoints[point] {
				if mp.RWFS != nil {
					var data []byte
					if data, err = pm.Bytes(); err != nil {
						err = fmt.Errorf("error getting bytes from page matter: %v", err)
						return
					}
					err = mp.RWFS.WriteFile(pm.Path, data, 0660)
					return
				}
			}
		}
	}
	err = fmt.Errorf("matching read-write mount point not found for page matter")
	return
}