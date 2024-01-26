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

package sha

import (
	"encoding/base64"
	"fmt"
	"hash"
	"regexp"
)

var (
	RxShasum64 = regexp.MustCompile(`^([a-f0-9]{64})$`)
	RxShasum10 = regexp.MustCompile(`^([a-f0-9]{10})$`)
)

// Valid returns true if the string given is 10 or 64 hexadecimal characters [a-f0-9]
func Valid(shasum string) (valid bool) {
	size := len(shasum)
	if valid = size == 10 || size == 64; valid {
		for _, char := range shasum {
			switch char {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f':
			default:
				valid = false
				return
			}
		}
	}
	return
}

func makeHash(h hash.Hash, data []byte) (shasum string, err error) {
	if _, err = h.Write(data); err != nil {
		return
	}
	shasum = base64.StdEncoding.EncodeToString(h.Sum(nil))
	return
}

func makeShasum(h hash.Hash, data []byte) (shasum string, err error) {
	if _, err = h.Write(data); err != nil {
		return
	}
	shasum = fmt.Sprintf("%x", h.Sum(nil))
	return
}

func mustFn[V []byte | string](data V, fn func(data V) (shasum string, err error)) (shasum string) {
	var err error
	if shasum, err = fn(data); err != nil {
		panic(err)
	}
	return
}
