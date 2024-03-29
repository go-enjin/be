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

type Features []Feature

func (f Features) Len() (count int) {
	return len(f)
}

func (f Features) Has(tag Tag) (present bool) {
	for _, ef := range f {
		if present = ef.Tag() == tag; present {
			return
		}
	}
	return
}

func (f Features) Get(tag Tag) (found Feature) {
	for _, ef := range f {
		if ef.Tag() == tag {
			found = ef
			return
		}
	}
	return
}
