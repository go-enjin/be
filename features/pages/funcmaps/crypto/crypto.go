//go:build page_funcmaps || pages || all

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

package crypto

import (
	"github.com/gofrs/uuid"
	"github.com/urfave/cli/v2"

	sha "github.com/go-corelibs/shasum"
	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
)

const Tag feature.Tag = "pages-funcmaps-crypto"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

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
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	return
}

func (f *CFeature) Setup(enjin feature.Internals) {
	f.CFeature.Setup(enjin)

}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	return
}

func (f *CFeature) MakeFuncMap(ctx beContext.Context) (fm feature.FuncMap) {
	fm = feature.FuncMap{
		"newUUID":        uuid.NewV4,
		"uuidFromString": uuid.FromString,
		"byteHash10":     sha.BriefSum[[]byte],
		"byteHash64":     sha.Sum256[[]byte],
		"dataHash10":     sha.BriefSum[string],
		"dataHash64":     sha.Sum256[string],
		"shasum224":      sha.Sum224[string],
		"shasum256":      sha.Sum256[string],
		"shasum384":      sha.Sum384[string],
		"shasum512":      sha.Sum512[string],
		"shasum224b":     sha.Sum224[[]byte],
		"shasum256b":     sha.Sum256[[]byte],
		"shasum384b":     sha.Sum384[[]byte],
		"shasum512b":     sha.Sum512[[]byte],
	}
	return
}
