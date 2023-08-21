//go:build srv_fastcgi || srv || all

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

package fastcgi

import (
	"net/http"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	bePath "github.com/go-enjin/be/pkg/path"
	"github.com/go-enjin/be/pkg/slices"
)

var (
	DefaultMount   = "/"
	DefaultDocRoot = "./docroot"
	DefaultNetwork = "auto"
)

const Tag feature.Tag = "srv-fastcgi"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type Feature interface {
	feature.Feature
	feature.ApplyMiddleware
}

type MakeFeature interface {
	Make() Feature

	SetMount(path string) MakeFeature
	SetDocRoot(path string) MakeFeature
	SetFastCGI(source string) MakeFeature
	SetNetwork(network string) MakeFeature
	SetDirIndex(filename string) MakeFeature
	UseEnv(keys ...string) MakeFeature
}

type CFeature struct {
	feature.CFeature

	mount    string
	source   string
	docroot  string
	network  string
	dirIndex string

	envKeys []string
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
	f.mount = DefaultMount
	f.docroot = DefaultDocRoot
	f.network = DefaultNetwork
	f.dirIndex = DefaultDirIndex
}

func (f *CFeature) SetMount(path string) MakeFeature {
	f.mount = path
	return f
}

func (f *CFeature) SetDocRoot(path string) MakeFeature {
	if bePath.IsDir(path) {
		f.docroot = path
	} else {
		log.FatalDF(1, "path not found or not a directory: %v", path)
	}
	return f
}

func (f *CFeature) SetFastCGI(source string) MakeFeature {
	f.source = source
	return f
}

func (f *CFeature) SetNetwork(network string) MakeFeature {
	f.network = network
	return f
}

func (f *CFeature) SetDirIndex(filename string) MakeFeature {
	f.dirIndex = filename
	return f
}

func (f *CFeature) UseEnv(keys ...string) MakeFeature {
	for _, key := range keys {
		if !slices.Within(key, f.envKeys) {
			f.envKeys = append(f.envKeys, key)
		}
	}
	return f
}

func (f *CFeature) Make() Feature {
	if f.source == "" {
		log.FatalDF(1, "%v feature requires .SetFastCGI", f.Tag())
	} else if f.network == "" {
		log.FatalDF(1, "%v feature requires .SetNetwork", f.Tag())
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

func (f *CFeature) Apply(s feature.System) (err error) {
	var handler http.Handler
	if handler, err = newHandler(f.dirIndex, f.docroot, f.network, f.source, f.envKeys); err != nil {
		return
	}
	s.Router().Mount(f.mount, handler)
	return
}