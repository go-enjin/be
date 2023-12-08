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

package tokens

import (
	"time"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/crypto"
	"github.com/go-enjin/be/pkg/feature"
	uses_kvc "github.com/go-enjin/be/pkg/feature/uses-kvc"
	"github.com/go-enjin/be/pkg/hash/sha"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
)

var (
	DefaultNumRandomBytes int           = 128
	DefaultDuration       time.Duration = time.Minute * 10
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "srv-factory-tokens"

type Feature interface {
	feature.Feature
	feature.TokenFactoryFeature
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

	tokens map[string]map[string]time.Time
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
	f.tokens = make(map[string]map[string]time.Time)
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
	} else if err = f.CUsesKVC.StartupUsesKVC(f.Enjin.Features()); err != nil {
		return
	}
	return
}

func (f *CFeature) Shutdown() {
	f.CFeature.Shutdown()
}

func (f *CFeature) makeKey(k string) (key string) {
	key = f.KebabTag + "__" + k
	return
}

func (f *CFeature) randomValue() (value string) {
	var err error
	if value, err = crypto.RandomValue(f.numRandomBytes); err != nil {
		log.ErrorF("crypto.RandomValue error: %v", err)
	}
	return
}

func (f *CFeature) expire() {
	f.Lock()
	defer f.Unlock()
	now := time.Now()

	prune := make(map[string][]string)
	for key, values := range f.tokens {
		for value, expiration := range values {
			if now.After(expiration) {
				prune[key] = append(prune[key], value)
			}
		}
	}
	for key, values := range prune {
		for _, value := range values {
			delete(f.tokens[key], value)
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
	var tokenBucket, aliasBucket feature.KeyValueStore

	aliasKey := sha.MustDataHash10(key)

	if tokenBucket, err = f.KVC().Bucket(key); err != nil {
		// TODO: is this a case to panic or just trace ignore?
		//panic(err)
		//log.TraceF("error getting token bucket for key: %q - %v", key, err)
		return
	} else if aliasBucket, err = f.KVC().Bucket(aliasKey); err != nil {
		// TODO: is this a case to panic or just trace ignore?
		//panic(err)
		//log.TraceF("error getting alias bucket for key: %q - %v", key, err)
		return
	}

	var shasum string

	if len(value) == 10 {
		shasum = value
		if v, err = aliasBucket.Get(shasum); err != nil {
			// not a valid token alias and nothing to clean up
			return
		} else if value, ok = v.(string); !ok {
			// shasum exists but it's not a string?!
			_ = aliasBucket.Delete(shasum)
			return
		}
	} else {
		shasum = sha.MustDataHash10(value)
	}

	defer func() {
		delete(f.tokens[key], value)
		_ = tokenBucket.Delete(value)
		_ = aliasBucket.Delete(shasum)
	}()

	if v, err = tokenBucket.Get(value); err != nil {
	} else if expiration, ok = v.(time.Time); ok {
		valid = time.Now().Before(expiration)
	}

	return
}

func (f *CFeature) get(key string, duration time.Duration) (value, shasum string) {
	f.expire()
	f.Lock()
	defer f.Unlock()
	if duration < 1 {
		duration = f.duration
	}

	aliasKey := sha.MustDataHash10(key)

	var err error
	var tokenBucket, aliasBucket feature.KeyValueStore
	if tokenBucket, err = f.KVC().Bucket(key); err != nil {
		panic(err)
	} else if aliasBucket, err = f.KVC().Bucket(aliasKey); err != nil {
		panic(err)
	}

	value = f.randomValue()
	shasum = sha.MustDataHash10(value)

	maps.MakeTypedKey(key, f.tokens)
	f.tokens[key][value] = time.Now().Add(f.duration)
	if err = tokenBucket.Set(value, f.tokens[key][value]); err != nil {
		panic(err)
	} else if err = aliasBucket.Set(shasum, value); err != nil {
		panic(err)
	}

	return
}

// VerifyToken will validate and evict the given nonce value
func (f *CFeature) VerifyToken(key, value string) (valid bool) {
	valid = f.verify(f.makeKey(key), value)
	return
}

// CreateToken will add a new nonce associated with the given key
func (f *CFeature) CreateToken(key string) (value, shasum string) {
	value, shasum = f.get(f.makeKey(key), f.duration)
	return
}

// CreateTokenWith will add a new token with the given duration, associated with the given key
func (f *CFeature) CreateTokenWith(key string, duration time.Duration) (value, shasum string) {
	value, shasum = f.get(f.makeKey(key), duration)
	return
}