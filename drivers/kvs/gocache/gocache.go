//go:build driver_kvs_gocache || drivers_kvs || drivers || all

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

package gocache

import (
	"fmt"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
)

const (
	NoExpiration time.Duration = -1
)

const Tag feature.Tag = "drivers-kvs-gocache"

var (
	BucketNotFound = fmt.Errorf("bucket not found")
	BucketExists   = fmt.Errorf("bucket exists")
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type Feature interface {
	feature.Feature
	feature.KeyValueCaches
}

type MakeFeature interface {
	BigCacheSupport
	IMCacheSupport
	MemorySupport
	RedisSupport
	RistrettoSupport

	Make() Feature
}

type CFeature struct {
	feature.CFeature

	order  []string
	caches map[string]feature.KeyValueCache
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
	f.caches = make(map[string]feature.KeyValueCache)
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) addCache(key string, kvc feature.KeyValueCache) {
	f.order = append(f.order, key)
	f.caches[key] = kvc
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	for _, key := range f.order {
		kvc, _ := f.caches[key]
		if kvcf, ok := kvc.(feature.KeyValueCacheFeature); ok {
			if err = kvcf.Build(b); err != nil {
				return
			}
		}
	}
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	for _, key := range f.order {
		kvc, _ := f.caches[key]
		if kvcf, ok := kvc.(feature.KeyValueCacheFeature); ok {
			if err = kvcf.Startup(ctx); err != nil {
				return
			}
		}
	}
	return
}

func (f *CFeature) Shutdown() {
	for _, key := range f.order {
		kvc, _ := f.caches[key]
		if kvcf, ok := kvc.(feature.KeyValueCacheFeature); ok {
			kvcf.Shutdown()
		}
	}
	return
}

func (f *CFeature) Get(name string) (kvs feature.KeyValueCache, err error) {
	if c, ok := f.caches[name]; ok {
		kvs = c
		return
	}
	err = BucketNotFound
	return
}
