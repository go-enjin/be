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
	"github.com/urfave/cli/v2"

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
	f.FeatureTag = tag
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

func (f *CFeature) Shutdown() {

}

func (f *CFeature) MakeFuncMap(ctx beContext.Context) (fm feature.FuncMap) {
	pfs := f.Enjin.PublicFileSystems().Lookup()
	fm = feature.FuncMap{
		"fsHash": func(path string) (shasum string) {
			shasum, _ = pfs.FindFileShasum(path)
			return
		},
		"fsUrl": func(path string) (url string) {
			url = path
			if shasum, err := pfs.FindFileShasum(path); err == nil {
				url += "?rev=" + shasum
			}
			return
		},
		"fsMime":         pfs.FindFileMime,
		"fsExists":       pfs.FileExists,
		"fsListFiles":    pfs.ListFiles,
		"fsListAllFiles": pfs.ListAllFiles,
		"fsListDirs":     pfs.ListDirs,
		"fsListAllDirs":  pfs.ListAllDirs,
	}
	return
}