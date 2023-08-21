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

package funcmaps

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/features/pages/funcmaps/casting"
	"github.com/go-enjin/be/features/pages/funcmaps/dates"
	"github.com/go-enjin/be/features/pages/funcmaps/dict"
	"github.com/go-enjin/be/features/pages/funcmaps/elements"
	"github.com/go-enjin/be/features/pages/funcmaps/forms"
	"github.com/go-enjin/be/features/pages/funcmaps/gtf"
	"github.com/go-enjin/be/features/pages/funcmaps/lang"
	"github.com/go-enjin/be/features/pages/funcmaps/logging"
	"github.com/go-enjin/be/features/pages/funcmaps/math"
	"github.com/go-enjin/be/features/pages/funcmaps/publicfs"
	"github.com/go-enjin/be/features/pages/funcmaps/slices"
	"github.com/go-enjin/be/features/pages/funcmaps/strcase"
	"github.com/go-enjin/be/features/pages/funcmaps/strings"
	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
)

const Tag feature.Tag = "pages-funcmaps"

type MakeFuncMapFn = func(ctx beContext.Context) (fn interface{})

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

	Include(others ...feature.FuncMapProvider) MakeFeature

	Defaults() MakeFeature
	SetMakerFunc(name string, maker MakeFuncMapFn) MakeFeature
	SetStaticFunc(name string, fn interface{}) MakeFeature
}

type CFeature struct {
	feature.CFeature

	static feature.FuncMap
	makers map[string]MakeFuncMapFn
	others map[feature.Tag]feature.FuncMapProvider
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
	f.static = feature.FuncMap{}
	f.makers = make(map[string]MakeFuncMapFn)
	f.others = make(map[feature.Tag]feature.FuncMapProvider)
}

func (f *CFeature) Include(others ...feature.FuncMapProvider) MakeFeature {
	for _, other := range others {
		if _, exists := f.others[other.Tag()]; exists {
			log.FatalDF(1, "%v FuncMap feature exists already", other.Tag())
		}
		f.others[other.Tag()] = other
	}
	return f
}

func (f *CFeature) Defaults() MakeFeature {
	f.Include(
		casting.New().Make(),
		dates.New().Make(),
		dict.New().Make(),
		elements.New().Make(),
		forms.New().Make(),
		gtf.New().Make(),
		lang.New().Make(),
		logging.New().Make(),
		math.New().Make(),
		publicfs.New().Make(),
		slices.New().Make(),
		strcase.New().Make(),
		strings.New().Make(),
	)
	return f
}

func (f *CFeature) SetMakerFunc(name string, maker MakeFuncMapFn) MakeFeature {
	if maker == nil {
		delete(f.makers, name)
	} else {
		f.makers[name] = maker
	}
	return f
}

func (f *CFeature) SetStaticFunc(name string, fn interface{}) MakeFeature {
	if fn == nil {
		delete(f.static, name)
	} else {
		f.static[name] = fn
	}
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	for _, name := range maps.OrderedKeys(f.others) {
		if err = f.others[name].Build(b); err != nil {
			err = fmt.Errorf("error building %s funcmap: %v", name, err)
			return
		}
	}
	return
}

func (f *CFeature) Setup(enjin feature.Internals) {
	f.CFeature.Setup(enjin)
	for _, name := range maps.OrderedKeys(f.others) {
		f.others[name].Setup(enjin)
	}
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	for _, name := range maps.OrderedKeys(f.others) {
		if err = f.others[name].Startup(ctx); err != nil {
			err = fmt.Errorf("error starting %s funcmap: %v", name, err)
			return
		}
	}
	return
}

func (f *CFeature) Shutdown() {
	f.CFeature.Shutdown()
	for _, name := range maps.OrderedKeys(f.others) {
		f.others[name].Shutdown()
	}
}

func (f *CFeature) MakeFuncMap(ctx beContext.Context) (fm feature.FuncMap) {
	fm = feature.FuncMap{}
	for _, name := range maps.SortedKeys(f.static) {
		fm[name] = f.static[name]
	}
	for _, tag := range maps.OrderedKeys(f.others) {
		if more := f.others[tag].MakeFuncMap(ctx); len(more) > 0 {
			fm.Apply(more)
		}
	}
	if len(ctx) > 0 {
		// only add makers when context present
		for _, name := range maps.SortedKeys(f.makers) {
			fm[name] = f.makers[name](ctx)
		}
	}
	return
}