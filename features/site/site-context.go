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

package site

import (
	"fmt"

	"github.com/go-enjin/be/pkg/context"
)

func (f *CFeature) GetContextUnsafe(eid string) (ctx context.Context) {
	var ok bool
	if v, err := f.userContextBucket.Get(eid); err != nil || v == nil {
		ctx = context.Context{}
	} else if ctx, ok = v.(context.Context); ok {
	} else if ctx, ok = v.(map[string]interface{}); ok {
	} else {
		panic(fmt.Errorf("value is neither a beContext.Context nor a map[string]interface{}: %T", v))
	}
	return
}

func (f *CFeature) GetContext(eid string) (ctx context.Context) {
	f.userContextLocker.Lock(eid)
	defer f.userContextLocker.Unlock(eid)
	ctx = f.GetContextUnsafe(eid)
	return
}

func (f *CFeature) SetContextUnsafe(eid string, ctx context.Context) {
	if err := f.userContextBucket.Set(eid, ctx); err != nil {
		panic(err)
	}
	return
}

func (f *CFeature) SetContext(eid string, ctx context.Context) {
	f.userContextLocker.Lock(eid)
	defer f.userContextLocker.Unlock(eid)
	f.SetContextUnsafe(eid, ctx)
	return
}

func (f *CFeature) ApplyContext(eid string, changes context.Context) {
	f.userContextLocker.Lock(eid)
	defer f.userContextLocker.Unlock(eid)
	ctx := f.GetContextUnsafe(eid)
	ctx.ApplySpecific(changes)
	f.SetContextUnsafe(eid, ctx)
	return
}

func (f *CFeature) DeleteContextKeys(eid string, keys ...string) {
	f.userContextLocker.Lock(eid)
	defer f.userContextLocker.Unlock(eid)
	ctx := f.GetContextUnsafe(eid)
	ctx.DeleteKeys(keys...)
	f.SetContextUnsafe(eid, ctx)
	return
}