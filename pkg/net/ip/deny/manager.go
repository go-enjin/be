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

package deny

import (
	"regexp"
	"sync"
	"time"
)

var (
	DenyDuration int64 = 60 * 60 * 24

	_manager *manager = &manager{
		rw:   &sync.RWMutex{},
		deny: make(map[string]int64),
		path: make(map[string]*regexp.Regexp),
	}
)

type manager struct {
	rw   *sync.RWMutex
	deny map[string]int64
	path map[string]*regexp.Regexp
}

func (m *manager) Deny(address string) int64 {
	m.expiration()
	m.rw.Lock()
	defer m.rw.Unlock()
	if _, ok := m.deny[address]; !ok {
		m.deny[address] = time.Now().Unix() + DenyDuration
	}
	return m.deny[address]
}

func (m *manager) Denied(address string) (denied bool) {
	m.expiration()
	m.rw.RLock()
	defer m.rw.RUnlock()
	_, denied = m.deny[address]
	return
}

func (m *manager) Restrict(pattern string) (err error) {
	m.rw.Lock()
	defer m.rw.Unlock()
	if _, ok := m.path[pattern]; !ok {
		var rx *regexp.Regexp
		if rx, err = regexp.Compile(pattern); err == nil {
			m.path[pattern] = rx
		}
	}
	return
}

func (m *manager) Restricted(path string) (restricted bool) {
	m.rw.RLock()
	defer m.rw.RUnlock()
	for _, rx := range m.path {
		if rx.MatchString(path) {
			return true
		}
	}
	return
}

func (m *manager) expiration() {
	m.rw.Lock()
	defer m.rw.Unlock()
	now := time.Now().Unix()
	expired := []string{}
	for address, expiry := range m.deny {
		if expiry <= now {
			expired = append(expired, address)
		}
	}
	for _, address := range expired {
		delete(m.deny, address)
	}
}