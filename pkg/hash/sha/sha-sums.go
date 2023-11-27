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

package sha

import (
	"crypto/sha256"
	"crypto/sha512"
)

func Shasum224[V []byte | string](v V) (shasum string, err error) {
	shasum, err = makeShasum(sha256.New224(), []byte(v))
	return
}

func Shasum256[V []byte | string](v V) (shasum string, err error) {
	shasum, err = makeShasum(sha256.New(), []byte(v))
	return
}

func Shasum384[V []byte | string](v V) (shasum string, err error) {
	shasum, err = makeShasum(sha512.New384(), []byte(v))
	return
}

func Shasum512[V []byte | string](v V) (shasum string, err error) {
	shasum, err = makeShasum(sha512.New(), []byte(v))
	return
}
