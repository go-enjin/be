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

func MustDataHash64[V []byte | string](data V) (shasum string) {
	shasum = mustShasum(data, DataHash64[V])
	return
}

func DataHash64[V []byte | string](data V) (shasum string, err error) {
	shasum, err = Shasum256(data)
	return
}

func MustDataHash10[V []byte | string](data V) (shasum string) {
	shasum = mustShasum(data, DataHash10[V])
	return
}

func DataHash10[V []byte | string](data V) (shasum string, err error) {
	if shasum, err = Shasum256(data); err == nil {
		shasum = shasum[0:10]
	}
	return
}

func VerifyData64[V []byte | string](data V, shasum string) (verified bool) {
	if actual, err := Shasum256(data); err != nil {
	} else if verified = actual == shasum; verified {
	}
	return
}