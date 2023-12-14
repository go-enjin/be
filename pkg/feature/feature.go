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
	"fmt"
	"sync"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/log"
)

const NotImplemented Tag = "not-implemented"

type Feature interface {
	// Construct is used as the last call within the standard NewTagged(tag) constructor and is used by extended types
	// to perform initializations which require the .FeatureTag set
	Construct(this interface{})
	// Init is used as the first call within the standard NewTagged(tag) constructor to perform initializations which
	// do not require the .FeatureTag set
	Init(this interface{})
	// Tag is the feature.Tag for this particular feature instance
	Tag() (tag Tag)
	// BaseTag is the stock tag common to all feature instances of this type
	BaseTag() (pkg Tag)
	// This returns an interface{} reference to the underlying structure instance
	This() (this interface{})
	// Self returns f.this, typed as a Feature, in such a way that calling f.Self().Thing() from a base type will invoke
	// the type's overloaded method; for example:
	//  - feature A implements `.Thing()` method and calls `f.Thing()` from some `.OtherThing()` method
	//  - feature B embeds feature A and overloads the `.Thing()` method
	//  - when feature A calls `f.Thing()` within `.OtherThing()`, feature A's method is invoked
	//  - if feature A instead calls `f.Self().Thing()` instead, feature B's method is invoked
	//
	// The example above is of course contrived as this only works with Feature methods, however, the design pattern can
	// be re-used in other systems to achieve the same effect, see EditorFeature.SelfEditor for another example.
	Self() (f Feature)
	State() (state LifeCycleState)
	Depends() (deps Tags)
	UsageNotes() (notes []string)
	Build(c Buildable) (err error)
	Setup(enjin Internals)
	Startup(ctx *cli.Context) (err error)
	Shutdown()
}

type ReloadableFeature interface {
	Reload()
}

type MakeFeature interface {
	Make() Feature
}

type CFeature struct {
	this  interface{}
	ctx   context.Context
	state LifeCycleState

	KebabTag   string
	PackageTag Tag
	FeatureTag Tag

	Enjin Internals

	sync.RWMutex
}

func (f *CFeature) UsageNotes() (notes []string) {
	return
}

func (f *CFeature) Construct(this interface{}) {
	f.KebabTag = f.Tag().Kebab()
	f.SetState(StateConstructed)
	return
}

func (f *CFeature) Init(this interface{}) {
	f.this = this
	f.ctx = context.New()
	f.PackageTag = NotImplemented
	f.FeatureTag = NotImplemented
	f.SetState(StateInitialized)
}

func (f *CFeature) Make() Feature {
	f.SetState(StateMade)
	return f.Self()
}

func (f *CFeature) Build(b Buildable) (err error) {
	f.SetState(StateBuilt)
	return
}

func (f *CFeature) Setup(enjin Internals) {
	f.Enjin = enjin
	f.SetState(StateSetup)
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if f.Tag() == NotImplemented {
		panic("not implemented")
	}
	log.DebugDF(1, "%v starting up", f.Tag())
	f.SetState(StateStarted)
	return
}

func (f *CFeature) Shutdown() {
	log.DebugDF(1, "%v shutting down", f.FeatureTag)
	f.SetState(StateShutdown)
}

func (f *CFeature) BaseTag() (pkg Tag) {
	if f.PackageTag == NotImplemented {
		panic("not implemented")
	}
	return f.PackageTag
}

func (f *CFeature) Tag() (tag Tag) {
	if f.FeatureTag == NotImplemented {
		panic("not implemented")
	}
	return f.FeatureTag
}

func (f *CFeature) SetState(state LifeCycleState) {
	if !state.Valid() {
		panic(fmt.Errorf("invalid state: %d", state))
	}
	f.Lock()
	defer f.Unlock()
	f.state = state
	return
}

func (f *CFeature) State() (state LifeCycleState) {
	f.RLock()
	defer f.RUnlock()
	state = f.state
	return
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

func (f *CFeature) Depends() (deps Tags) {
	return
}

func (f *CFeature) CloneBaseFeature() (cloned CFeature) {
	cloned = CFeature{
		this:       f.this,
		ctx:        f.ctx.Copy(),
		state:      f.state,
		KebabTag:   f.KebabTag,
		PackageTag: f.PackageTag,
		FeatureTag: f.FeatureTag,
		Enjin:      f.Enjin,
	}
	return
}