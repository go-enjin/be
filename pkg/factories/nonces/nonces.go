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

package nonces

import (
	"sync"
	"time"

	"github.com/go-enjin/be/pkg/crypto"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
)

var (
	DefaultRandomBytes               = 32
	DefaultDuration    time.Duration = time.Hour * 24
)

type nonces struct {
	cache map[string]map[string]time.Time

	duration time.Duration

	sync.RWMutex
}

// New creates a new feature.NonceFactory with the given expiration duration, if the expires value is less than 1s, the
// package default of 24h is used
func New(expires time.Duration) (factory feature.NonceFactory) {
	if expires < 1 {
		expires = DefaultDuration
	}
	factory = &nonces{
		duration: expires,
		cache:    make(map[string]map[string]time.Time),
	}
	return
}

func (r *nonces) randomValue() (value string) {
	var err error
	if value, err = crypto.RandomValue(DefaultRandomBytes); err != nil {
		log.ErrorF("crypto.RandomValue error: %v", err)
	}
	return
}

func (r *nonces) expire() {
	r.Lock()
	defer r.Unlock()
	now := time.Now()
	prune := make(map[string][]string)
	for key, values := range r.cache {
		for value, expiration := range values {
			if now.After(expiration) {
				prune[key] = append(prune[key], value)
			}
		}
	}
	for key, values := range prune {
		for _, value := range values {
			delete(r.cache[key], value)
		}
	}
}

func (r *nonces) verify(key, value string) (valid bool) {
	r.expire()
	r.Lock()
	defer r.Unlock()
	if expiration, ok := r.cache[key][value]; ok {
		valid = time.Now().Before(expiration)
	}
	delete(r.cache[key], value)
	return
}

func (r *nonces) get(key string) (value string) {
	r.expire()
	r.Lock()
	defer r.Unlock()
	value = r.randomValue()
	maps.MakeTypedKey(key, r.cache)
	r.cache[key][value] = time.Now().Add(r.duration)
	return
}

// VerifyNonce will validate and evict the given nonce value
func (r *nonces) VerifyNonce(key, value string) (valid bool) {
	valid = r.verify(key, value)
	return
}

// CreateNonce will add a new nonce associated with the given key
func (r *nonces) CreateNonce(key string) (value string) {
	value = r.get(key)
	return
}
