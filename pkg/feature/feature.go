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

package feature

import (
	"sync"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/log"
)

const NotImplemented Tag = "NotImplemented"

type Feature interface {
	Init(this interface{})
	Tag() (tag Tag)
	Self() (f Feature)
	Depends() (deps Tags)
	Context() (ctx context.Context)
	Build(c Buildable) (err error)
	Setup(enjin Internals)
	Startup(ctx *cli.Context) (err error)
	Shutdown()
}

type MakeFeature interface {
	Make() Feature
}

type CFeature struct {
	this interface{}
	ctx  context.Context

	sync.RWMutex
}

func (f *CFeature) Init(this interface{}) {
	f.this = this
	f.ctx = context.New()
}

func (f *CFeature) Tag() (tag Tag) {
	return NotImplemented
}

func (f *CFeature) This() (this interface{}) {
	return f.this
}

func (f *CFeature) Self() (self Feature) {
	var ok bool
	if self, ok = f.this.(Feature); !ok {
		log.FatalF("feature not a feature: %T %+v", f.this, f.this)
	}
	return
}

func (f *CFeature) Make() Feature {
	log.DebugF("making feature: %v", f.Self().Tag())
	return f.Self()
}

func (f *CFeature) Depends() (deps Tags) {
	return
}

func (f *CFeature) Context() (ctx context.Context) {
	ctx = f.ctx.Copy()
	return
}

func (f *CFeature) Build(c Buildable) (err error) {
	return
}

func (f *CFeature) Setup(enjin Internals) {
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	return
}

func (f *CFeature) Shutdown() {
}