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

package theme

import (
	"sync"
)

var (
	registerPartialsLock = &sync.RWMutex{}
	registeredPartials   = map[string]map[string]map[string]string{
		"head": {
			"head": {},
			"tail": {},
		},
		"body": {
			"head": {},
			"tail": {},
		},
	}
)

func RegisterPartialHeadHead(name, tmpl string) {
	registerPartialsLock.Lock()
	defer registerPartialsLock.Unlock()
	registeredPartials["head"]["head"][name] = tmpl
}

func RegisterPartialHeadTail(name, tmpl string) {
	registerPartialsLock.Lock()
	defer registerPartialsLock.Unlock()
	registeredPartials["head"]["tail"][name] = tmpl
}

func RegisterPartialBodyHead(name, tmpl string) {
	registerPartialsLock.Lock()
	defer registerPartialsLock.Unlock()
	registeredPartials["body"]["head"][name] = tmpl
}

func RegisterPartialBodyTail(name, tmpl string) {
	registerPartialsLock.Lock()
	defer registerPartialsLock.Unlock()
	registeredPartials["body"]["tail"][name] = tmpl
}