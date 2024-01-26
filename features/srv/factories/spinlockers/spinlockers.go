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

package spinlockers

import (
	"time"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "srv-factory-spinlockers"

type Feature interface {
	feature.Feature
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
	if err = f.CFeature.Build(b); err != nil {
		return
	}
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	return
}

func (f *CFeature) Shutdown() {
	f.CFeature.Shutdown()
}

func (f *CFeature) NewSyncLocker(tag feature.Tag, key string, store feature.KeyValueStore) (l feature.SyncLocker) {
	l = newSpinLocker(store, tag, key, -1, -1)
	return
}

func (f *CFeature) NewSyncLockerWith(tag feature.Tag, key string, store feature.KeyValueStore, timeout, interval time.Duration) (l feature.SyncLocker) {
	l = newSpinLocker(store, tag, key, timeout, interval)
	return
}

func (f *CFeature) NewSyncRWLocker(tag feature.Tag, key string, readStore, writeStore feature.KeyValueStore) (l feature.SyncRWLocker) {
	l = newSpinRWLocker(readStore, writeStore, tag, key, -1, -1)
	return
}

func (f *CFeature) NewSyncRWLockerWith(tag feature.Tag, key string, readStore, writeStore feature.KeyValueStore, timeout, interval time.Duration) (l feature.SyncRWLocker) {
	l = newSpinRWLocker(readStore, writeStore, tag, key, timeout, interval)
	return
}
