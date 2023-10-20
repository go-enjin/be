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

package maps

import (
	"github.com/go-enjin/be/pkg/deepcopy"
)

type BaseTypes interface {
	~bool |
		~string | ~[]byte |
		~float32 | ~float64 |
		~complex64 | ~complex128 |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

func CopyBaseSlice[V BaseTypes](src []V) (dst []V) {
	dst = make([]V, len(src))
	copy(src, dst)
	return
}

func CopyBaseMap[T BaseTypes](src map[string]T) (dst map[string]T) {
	dst = make(map[string]T)
	for k, v := range src {
		dst[k] = v
	}
	return
}

func DeepCopy(src map[string]interface{}) (dst map[string]interface{}) {
	dst, _ = deepcopy.DeepCopy(src).(map[string]interface{})
	return
}