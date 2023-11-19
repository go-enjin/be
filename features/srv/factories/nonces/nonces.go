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

package nonces

import (
	"crypto/rand"
	"encoding/hex"
	rand2 "math/rand"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	uses_kvc "github.com/go-enjin/be/pkg/feature/uses-kvc"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
)

var (
	DefaultNumRandomBytes int           = 32
	DefaultDuration       time.Duration = time.Hour
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "srv-factory-nonces"

type Feature interface {
	feature.Feature
	feature.NonceFactoryFeature
}

type MakeFeature interface {
	uses_kvc.MakeFeature[MakeFeature]

	SetDuration(duration time.Duration) MakeFeature
	SetNumRandomBytes(size int) MakeFeature

	Make() Feature
}

type CFeature struct {
	feature.CFeature
	uses_kvc.CUsesKVC[MakeFeature]

	numRandomBytes int
	duration       time.Duration

	nonces map[string]map[string]time.Time
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
	f.CUsesKVC.InitUsesKVC(this)
	f.nonces = make(map[string]map[string]time.Time)
	f.numRandomBytes = DefaultNumRandomBytes
	f.duration = DefaultDuration
	return
}

func (f *CFeature) SetDuration(duration time.Duration) MakeFeature {
	if duration.Seconds() < 60 {
		log.FatalDF(1, "minimum duration required is 1m, default is %v", DefaultDuration)
	}
	f.duration = duration
	return f
}

func (f *CFeature) SetNumRandomBytes(size int) MakeFeature {
	if size <= 7 {
		log.FatalDF(1, "minimum bytes required is 8, default is %d", DefaultNumRandomBytes)
	}
	f.numRandomBytes = size
	return f
}

func (f *CFeature) Make() (feat Feature) {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	if err = f.CFeature.Build(b); err != nil {
		return
	} else if err = f.CUsesKVC.BuildUsesKVC(); err != nil {
		return
	}
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	} else if ee := f.CUsesKVC.StartupUsesKVC(f.Enjin.Features()); ee != nil {
		return
	}
	return
}

func (f *CFeature) Shutdown() {
	f.CFeature.Shutdown()
}

// VerifyNonce will return true if the nonce is valid and consume the nonce in the process
func (f *CFeature) VerifyNonce(key, value string) (valid bool) {
	valid = f.verify(key, value)
	return
}

// CreateNonce will add a new nonce instance to the given key
func (f *CFeature) CreateNonce(key string) (value string) {
	value = f.create(key)
	return
}

func (f *CFeature) expire() {
	f.Lock()
	defer f.Unlock()
	now := time.Now()

	prune := make(map[string][]string)
	for key, values := range f.nonces {
		for value, expiration := range values {
			if now.After(expiration) {
				prune[key] = append(prune[key], value)
			}
		}
	}
	for key, values := range prune {
		for _, value := range values {
			delete(f.nonces[key], value)
			if bucket, err := f.KVC().Bucket(key); err == nil {
				_ = bucket.Delete(value)
			}
		}
	}
}

func (f *CFeature) verify(key, value string) (valid bool) {
	f.expire()
	f.Lock()
	defer f.Unlock()

	var ok bool
	var err error
	var v interface{}
	var expiration time.Time
	var bucket feature.KeyValueStore

	if bucket, err = f.KVC().Bucket(key); err != nil {
		return
	} else if v, err = bucket.Get(value); err != nil {
		return
	} else if expiration, ok = v.(time.Time); ok {
		valid = time.Now().Before(expiration)
	}

	delete(f.nonces[key], value)
	_ = bucket.Delete(value)
	return
}

func (f *CFeature) create(key string) (value string) {
	f.expire()
	f.Lock()
	defer f.Unlock()

	var err error
	var bucket feature.KeyValueStore
	if bucket, err = f.KVC().Bucket(key); err != nil {
		panic(err)
	}

	value = f.randomValue()
	maps.MakeTypedKey(key, f.nonces)
	f.nonces[key][value] = time.Now().Add(f.duration)
	if err = bucket.Set(value, f.nonces[key][value]); err != nil {
		panic(err)
	}

	return
}

func (f *CFeature) randomValue() (value string) {
	b := make([]byte, f.numRandomBytes)
	if _, e := rand.Read(b); e != nil {
		log.ErrorF("crypto/rand error: %v; falling back to math/rand", e)
		if _, ee := rand2.Read(b); ee != nil {
			log.ErrorF("math/rand error: %v", ee)
		}
	}
	value = hex.EncodeToString(b)
	return
}