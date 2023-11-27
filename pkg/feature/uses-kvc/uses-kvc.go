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

package uses_kvc

import (
	"fmt"
	"time"

	"github.com/go-enjin/be/pkg/feature"
)

type MakeFeature[M interface{}] interface {
	SetKeyValueCache(tag feature.Tag, name string) M
	SetExpirationInterval(expiration, interval time.Duration) M
}

type CUsesKVC[M interface{}] struct {
	_kvcTag  feature.Tag
	_kvcName string
	_kvc     feature.KeyValueCache

	_interval   time.Duration
	_expiration time.Duration

	this interface{}
}

func (f *CUsesKVC[M]) InitUsesKVC(this interface{}) {
	f.this = this
}

func (f *CUsesKVC[M]) SetKeyValueCache(tag feature.Tag, name string) M {
	f._kvcTag = tag
	f._kvcName = name
	t, _ := f.this.(M)
	return t
}

func (f *CUsesKVC[M]) SetExpirationInterval(expiration, interval time.Duration) M {
	f._expiration = expiration
	f._interval = interval
	t, _ := f.this.(M)
	return t
}

func (f *CUsesKVC[M]) BuildUsesKVC() (err error) {
	if f._kvcTag.IsNil() {
		err = fmt.Errorf("calling .SetKeyValueCache on this feature is required")
		return
	}
	return
}

func (f *CUsesKVC[M]) StartupUsesKVC(features *feature.FeaturesCache) (err error) {
	if kvf, ok := features.Get(f._kvcTag); ok {
		if kvcs, ok := kvf.This().(feature.KeyValueCaches); !ok {
			err = fmt.Errorf("%v does not implement kvs.KeyValueCaches", kvf.Tag())
			return
		} else if f._kvc, err = kvcs.Get(f._kvcName); err != nil {
			err = fmt.Errorf("%v has no cache named: %q", kvf.Tag(), f._kvcName)
			return
		}
	} else {
		err = fmt.Errorf("%v feature not found", f._kvcTag)
		return
	}
	return
}

func (f *CUsesKVC[M]) KVC() (kvc feature.KeyValueCache) {
	kvc = f._kvc
	return
}
