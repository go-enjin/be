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

package nonce

import (
	cryptoRand "crypto/rand"
	"encoding/hex"
	mathRand "math/rand"
	"sync"
	"time"

	"github.com/go-enjin/be/pkg/log"
)

type registry struct {
	nonce map[string]map[string]time.Time

	sync.RWMutex
}

var _known = &registry{
	nonce: make(map[string]map[string]time.Time),
}

func randomValue() (value string) {
	b := make([]byte, 32)
	if _, e := cryptoRand.Read(b); e != nil {
		log.ErrorF("crypto/rand error: %v; falling back to math/rand", e)
		if _, ee := mathRand.Read(b); ee != nil {
			log.ErrorF("math/rand error: %v", ee)
		}
	}
	value = hex.EncodeToString(b)
	return
}

func (r *registry) expire() {
	r.Lock()
	defer r.Unlock()
	now := time.Now()
	prune := make(map[string][]string)
	for key, values := range r.nonce {
		for value, expiration := range values {
			if now.After(expiration) {
				prune[key] = append(prune[key], value)
			}
		}
	}
	for key, values := range prune {
		for _, value := range values {
			delete(r.nonce[key], value)
		}
	}
}

func (r *registry) verify(key, value string) (valid bool) {
	r.expire()
	r.RLock()
	if expiration, ok := r.nonce[key][value]; ok {
		valid = time.Now().Before(expiration)
	}
	r.RUnlock()
	if valid {
		r.Lock()
		delete(r.nonce[key], value)
		r.Unlock()
	}
	return
}

func (r *registry) get(key string) (value string) {
	r.expire()
	r.Lock()
	defer r.Unlock()
	var ok bool
	var values map[string]time.Time
	if values, ok = r.nonce[key]; !ok {
		values = make(map[string]time.Time)
		r.nonce[key] = values
	}
	value = randomValue()
	r.nonce[key][value] = time.Now().Add(time.Hour)
	return
}

// Validate will consume the nonce
func Validate(name, value string) (valid bool) {
	valid = _known.verify(name, value)
	return
}

// Make will create a new nonce
func Make(name string) (value string) {
	value = _known.get(name)
	return
}