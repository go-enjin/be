//go:build page_shortcodes || pages || all

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

package shortcodes

import (
	"strings"

	"github.com/go-enjin/be/pkg/slices"
)

type Attributes struct {
	Keys   []string
	Lookup map[string]string
}

func newAttributes() (attrs *Attributes) {
	attrs = &Attributes{
		Keys:   make([]string, 0),
		Lookup: make(map[string]string),
	}
	return
}

func (a *Attributes) Set(key string, value string) {
	a.Keys = append(a.Keys, key)
	a.Lookup[key] = value
}

func (a *Attributes) Append(key string, value string) {
	if _, present := a.Lookup[key]; present {
		a.Lookup[key] += value
	} else {
		a.Set(key, value)
	}
}

func (a *Attributes) String() (output string) {
	for _, key := range a.Keys {
		output += " " + key + `="` + strings.ReplaceAll(a.Lookup[key], `"`, `\"`) + `"`
	}
	return
}

func (a *Attributes) Apply(other *Attributes) {
	for _, k := range other.Keys {
		if !slices.Within(k, a.Keys) {
			a.Keys = append(a.Keys, k)
		}
		a.Lookup[k] = other.Lookup[k]
	}
}

func (a *Attributes) Clone() (cloned *Attributes) {
	cloned = newAttributes()
	cloned.Keys = append(cloned.Keys, a.Keys...)
	for k, v := range a.Lookup {
		cloned.Lookup[k] = v
	}
	return
}
