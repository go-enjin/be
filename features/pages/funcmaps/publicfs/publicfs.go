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
	"strings"

	"github.com/urfave/cli/v2"

	clMime "github.com/go-corelibs/mime"
	clPath "github.com/go-corelibs/path"
	beContext "github.com/go-enjin/be/pkg/context"
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

func (f *CFeature) MakeFuncMap(ctx beContext.Context) (fm feature.FuncMap) {
	if f.Enjin != nil {
		pfs := f.Enjin.PublicFileSystems().Lookup()
		fm = feature.FuncMap{
			"fsHash": func(path string) (shasum string) {
				shasum, _ = pfs.FindFileShasum(path)
				return
			},
			"fsHash256": func(path string) (shasum string) {
				shasum, _ = pfs.FindFileSha256(path)
				return
			},
			"fsUrl": func(path string) (url string) {
				url = path
				if shasum, err := pfs.FindFileShasum(path); err == nil {
					url += "?rev=" + shasum
				}
				return
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
