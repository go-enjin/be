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

package feature

import (
	"sync"
)

type MatcherFn func(path string, pg Page) (found string, ok bool)

var (
	gPageMatcherRegistry = struct {
		known []MatcherFn
		sync.RWMutex
	}{}
)

func RegisterPageMatcherFuncs(matchers ...MatcherFn) {
	gPageMatcherRegistry.Lock()
	defer gPageMatcherRegistry.Unlock()
	gPageMatcherRegistry.known = append(gPageMatcherRegistry.known, matchers...)
}

func GetPageMatcherFuncs() (funcs []MatcherFn) {
	gPageMatcherRegistry.RLock()
	defer gPageMatcherRegistry.RUnlock()
	funcs = append(funcs, gPageMatcherRegistry.known...)
	return
}
