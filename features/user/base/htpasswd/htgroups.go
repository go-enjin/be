//go:build user_base_htpasswd || user_bases || all

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

package htpasswd

import (
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/go-corelibs/slices"
)

type htgroups struct {
	filepath string
	order    []string
	parsed   map[string][]string

	sync.RWMutex
}

var rxHtgroups = regexp.MustCompile(`^\s*([a-zA-Z][a-zA-Z0-9]*):\s*(.+?)\s*$`)

func newHtgroups(filepath string) (htg *htgroups, err error) {
	var data []byte
	if data, err = os.ReadFile(filepath); err != nil {
		return
	}
	htg = &htgroups{
		filepath: filepath,
		order:    make([]string, 0),
		parsed:   make(map[string][]string),
	}
	for _, line := range strings.Split(string(data), "\n") {
		if rxHtgroups.MatchString(line) {
			m := rxHtgroups.FindAllStringSubmatch(line, 1)
			group := m[0][1]
			htg.order = append(htg.order, group)
			for _, user := range strings.Split(m[0][2], " ") {
				if user = strings.TrimSpace(user); user != "" {
					htg.parsed[group] = append(htg.parsed[group], user)
				}
			}
		}
	}
	return
}

func (h *htgroups) AddToGroup(group string, users ...string) {
	h.Lock()
	defer h.Unlock()
	if !slices.Within(group, h.order) {
		h.order = append(h.order, group)
		h.parsed[group] = make([]string, 0)
	}
	for _, user := range users {
		if !slices.Within(user, h.parsed[group]) {
			h.parsed[group] = append(h.parsed[group], user)
		}
	}
}

func (h *htgroups) DeleteGroup(group string) {
	h.Lock()
	defer h.Unlock()
	if _, ok := h.parsed[group]; ok {
		var order []string
		for _, g := range h.order {
			if g != group {
				order = append(order, g)
			}
		}
		h.order = order
	}
}

func (h *htgroups) WriteFile(filepath string, mod os.FileMode) (err error) {
	h.RLock()
	defer h.RUnlock()
	var content string
	for _, group := range h.order {
		content += group + ": " + strings.Join(h.parsed[group], " ") + "\n"
	}
	err = os.WriteFile(filepath, []byte(content), mod)
	return
}
