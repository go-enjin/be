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

type Nonce struct {
	// Key is the string to use with nonce factories
	Key string `json:"key"`
	// Name is the form hidden input name
	Name string `json:"name"`
	// Value is the actual nonce to be used with this key/name combination; Value may be omitted in which case the theme
	// must call `Nonce $nonce.Key` to get the correct nonce value to use
	Value string `json:"-"`
}

type Nonces []Nonce