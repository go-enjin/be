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

package slices

// Copy creates a new slice (array) from the given slice
func Copy[T interface{}, S ~[]T](slice S) (copied S) {
	copied = make(S, len(slice))
	copy(copied, slice)
	return
}

// Truncate creates a new slice (array), of specified length, from the given slice
func Truncate[T interface{}, S ~[]T](slice S, length int) (truncated S) {
	truncated = make(S, length)
	copy(truncated, slice)
	return
}

// Insert creates a new slice (array) from the given slice, with additional values inserted at the given index
func Insert[T interface{}, S ~[]T](slice S, at int, values ...T) (modified S) {
	before := slice[:at]
	after := slice[at:]
	modified = make(S, 0)
	modified = append(modified, before...)
	modified = append(modified, values...)
	modified = append(modified, after...)
	return
}