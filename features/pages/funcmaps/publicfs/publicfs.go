//go:build page_funcmaps || pages || all

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

package publicfs

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"strings"

	"github.com/urfave/cli/v2"

	clContext "github.com/go-corelibs/context"
	clMime "github.com/go-corelibs/mime"
	clPath "github.com/go-corelibs/path"
	"github.com/go-enjin/be/pkg/feature"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "pages-funcmaps-public-fs"

type Feature interface {
	feature.Feature
	feature.FuncMapProvider
}

type MakeFeature interface {
	Make() Feature
}

type CFeature struct {
	feature.CFeature
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.CFeature.Construct(f)
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	return
}

func (f *CFeature) Make() (feat Feature) {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	return
}

func (f *CFeature) MakeFuncMap(ctx clContext.Context) (fm feature.FuncMap) {
	if f.Enjin != nil {
		pfs := f.Enjin.PublicFileSystems().Lookup()
		preloads := getPfsCache(ctx, "_fs_preloads")
		fsUrlCache := getPfsCache(ctx, "_fs_url")
		fm = feature.FuncMap{
			"fsHash": func(path string) (shasum string) {
				if item := fsUrlCache.get(path); item != nil {
					return item.shasum
				} else if shasum, _ = pfs.FindFileShasum(path); shasum != "" {
					fsUrlCache.set(path, shasum, "")
				}
				return
			},
			"fsHash256": func(path string) (shasum string) {
				shasum, _ = pfs.FindFileSha256(path)
				return
			},
			"fsUrl": func(path string) string {
				if item := fsUrlCache.get(path); item != nil {
					return item.revUrl()
				}
				if shasum, err := pfs.FindFileShasum(path); err == nil {
					fsUrlCache.set(path, shasum, "")
					return path + "?rev=" + shasum
				}
				return path
			},
			"fsPreloadUrl": func(path string) string {
				if item := preloads.get(path); item != nil {
					return item.revUrl()
				} else if shasum, err := pfs.FindFileShasum(path); err == nil {
					if mimeType, err := pfs.FindFileMime(path); err == nil {
						preloads.set(path, shasum, mimeType)
						return path + "?rev=" + shasum
					}
				}
				return path
			},
			"fsPreloadLinks": func() template.HTML {
				var outputs []string
				preloads.walk(func(item *pfsCacheItem) bool {
					// only images for now?
					if strings.HasPrefix(item.mimeType, "image/") {
						outputs = append(
							outputs,
							fmt.Sprintf(
								`<link rel="preload" as="image" type=%q href=%q />`,
								item.mimeType,
								item.revUrl(),
							),
						)
					}
					return true
				})
				return template.HTML(strings.Join(outputs, "\n\t"))
			},
			"fsDataUri":      f.DataUri,
			"fsMime":         pfs.FindFileMime,
			"fsExists":       pfs.FileExists,
			"fsListFiles":    pfs.ListFiles,
			"fsListAllFiles": pfs.ListAllFiles,
			"fsListDirs":     pfs.ListDirs,
			"fsListAllDirs":  pfs.ListAllDirs,
			"trimPageFmt":    f.TrimPageFormat,
			"parsePageFmt":   f.PageFormat,
			"pageFormats":    f.ListPageFormats,
			"basename":       clPath.Base,
			"basepath":       clPath.BasePath,
			"ext":            clPath.Ext,
		}
	}
	return
}

func (f *CFeature) DataUri(path string) (dataUri string) {
	pfs := f.Enjin.PublicFileSystems().Lookup()
	var err error
	var mime string
	var data []byte
	var encoded string
	if mime, err = pfs.FindFileMime(path); err != nil {
		return
	} else if data, err = pfs.ReadFile(path); err != nil || len(data) == 0 {
		return
	} else if encoded = base64.StdEncoding.EncodeToString(data); encoded == "" {
		return
	}
	switch clMime.PruneCharset(mime) {
	case "image/png":
		dataUri = "data:image/png;base64," + encoded
	case "image/jpg":
		dataUri = "data:image/jpg;base64," + encoded
	case "image/gif":
		dataUri = "data:image/gif;base64," + encoded
	case "image/apng":
		dataUri = "data:image/apng;base64," + encoded
	case "image/avif":
		dataUri = "data:image/avif;base64," + encoded
	case "image/webp":
		dataUri = "data:image/webp;base64," + encoded
	}
	return
}

func (f *CFeature) PageFormat(filename string) (match string) {
	t := f.Enjin.MustGetTheme()
	_, match = t.MatchFormat(filename)
	return
}

func (f *CFeature) ListPageFormats() (names []string) {
	t := f.Enjin.MustGetTheme()
	names = t.ListFormats()
	return
}

func (f *CFeature) TrimPageFormat(filename string) (basename string) {
	if match := f.PageFormat(filename); match != "" {
		basename = strings.TrimSuffix(filename, "."+match)
	} else {
		basename = filename
	}
	return
}
