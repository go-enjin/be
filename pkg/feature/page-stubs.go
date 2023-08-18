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

import "math/rand"

type PageStubs []*PageStub

func (s PageStubs) Random() (stub *PageStub) {
	idx := rand.Intn(len(s))
	stub = s[idx]
	return
}

func (s PageStubs) GetSource(source string) (found *PageStub) {
	for _, stub := range s {
		if stub.Source == source {
			found = stub
			return
		}
	}
	return
}

func (s PageStubs) GetShasum(shasum string) (found *PageStub) {
	for _, stub := range s {
		if stub.Shasum == shasum {
			found = stub
			return
		}
	}
	return
}

func (s PageStubs) HasShasum(shasum string) (found bool) {
	for _, stub := range s {
		if found = stub.Shasum == shasum; found {
			return
		}
	}
	return
}

func AnyStubsInStubs(src, tgt PageStubs) (found bool) {
	for _, stub := range src {
		if found = tgt.GetShasum(stub.Shasum) != nil; found {
			return
		}
	}
	return
}