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
	"sync"
	"time"

	"github.com/go-enjin/be/pkg/crypto"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/hash/sha"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
)

var (
	DefaultRandomBytes               = 128
	DefaultDuration    time.Duration = time.Minute * 10
)

type instance struct {
	value   string
	shasum  string
	expires time.Time
}

type registry struct {
	alias map[string]string
	cache map[string]map[string]instance

	numBytes int
	duration time.Duration

	sync.RWMutex
}

// New creates a new feature.TokenFactory with the given expiration duration, if the expires value is less than 1s, the
// package default of 10m is used
func New(expires time.Duration) (factory feature.TokenFactory) {
	if expires < 1 {
		expires = DefaultDuration
	}
	factory = &registry{
		numBytes: DefaultRandomBytes,
		duration: expires,
		alias:    make(map[string]string),
		cache:    make(map[string]map[string]instance),
	}
	return
}

func (r *registry) randomValue() (value string) {
	var err error
	if value, err = crypto.RandomValue(r.numBytes); err != nil {
		log.ErrorF("crypto.RandomValue error: %v", err)
	}
	return
}

func (r *registry) expire() {
	r.Lock()
	defer r.Unlock()
	now := time.Now()
	prune := make(map[string][]instance)
	for key, values := range r.cache {
		for _, token := range values {
			if now.After(token.expires) {
				prune[key] = append(prune[key], token)
			}
		}
	}
	for key, values := range prune {
		for _, token := range values {
			delete(r.cache[key], token.value)
			delete(r.alias, token.shasum)
		}
	}
}

func (r *registry) verify(key, value string) (valid bool) {
	r.expire()
	r.Lock()
	defer r.Unlock()
	var present bool
	var token instance
	if len(value) == 10 {
		if v, aliased := r.alias[value]; aliased {
			value = v
		}
	}
	if token, present = r.cache[key][value]; present {
		valid = time.Now().Before(token.expires)
	}
	if valid {
		// purge all tokens for key
		delete(r.cache, key)
	} else if present {
		// delete just the value
		delete(r.cache[key], value)
		delete(r.alias, token.shasum)
	}
	return
}

func (r *registry) get(key string, duration time.Duration) (value, shasum string) {
	r.expire()
	r.Lock()
	defer r.Unlock()
	if duration < 1 {
		duration = r.duration
	}
	value = r.randomValue()
	shasum, _ = sha.DataHash10(value)
	maps.MakeTypedKey(key, r.cache)
	r.cache[key][value] = instance{
		value:   value,
		shasum:  shasum,
		expires: time.Now().Add(duration),
	}
	r.alias[shasum] = value
	return
}

// VerifyToken will validate and evict the given nonce value
func (r *registry) VerifyToken(key, value string) (valid bool) {
	valid = r.verify(key, value)
	return
}

// CreateToken will add a new nonce associated with the given key
func (r *registry) CreateToken(key string) (value, shasum string) {
	value, shasum = r.get(key, r.duration)
	return
}

// CreateTokenWith will add a new token with the given duration, associated with the given key
func (r *registry) CreateTokenWith(key string, duration time.Duration) (value, shasum string) {
	value, shasum = r.get(key, duration)
	return
}