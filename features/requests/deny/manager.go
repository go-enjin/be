//go:build requests_deny || requests || all

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
	DefaultDuration int64 = 60 * 60 * 24
)

type manager struct {
	rw     *sync.RWMutex
	deny   map[string]int64
	path   map[string]*regexp.Regexp
	period int64
}

func newManager(period int64) (mgr *manager) {
	if period <= 0 {
		period = DefaultDuration
	}
	mgr = &manager{
		rw:     &sync.RWMutex{},
		deny:   make(map[string]int64),
		path:   make(map[string]*regexp.Regexp),
		period: period,
	}
	return
}

func (m *manager) SetPeriod(deny int64) {
	m.period = deny
}

func (m *manager) Block(address string) {
	m.expiration()
	m.rw.Lock()
	defer m.rw.Unlock()
	m.deny[address] = 0
}

func (m *manager) Unblock(address string) {
	m.expiration()
	m.rw.Lock()
	defer m.rw.Unlock()
	delete(m.deny, address)
}

func (m *manager) Deny(address string) int64 {
	m.expiration()
	m.rw.Lock()
	defer m.rw.Unlock()
	if _, ok := m.deny[address]; !ok {
		m.deny[address] = time.Now().Unix() + m.period
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
	var expired []string
	for address, expiry := range m.deny {
		if expiry > 0 && expiry <= now {
			expired = append(expired, address)
		}
	}
	for _, address := range expired {
		delete(m.deny, address)
	}
}
