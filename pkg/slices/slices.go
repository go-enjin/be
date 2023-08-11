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
func Copy[V interface{}, S ~[]V](slice S) (copied S) {
	copied = make(S, len(slice))
	copy(copied, slice)
	return
}

// Truncate creates a new slice (array), of specified length, from the given slice
func Truncate[V interface{}, S ~[]V](slice S, length int) (truncated S) {
	truncated = make(S, length)
	copy(truncated, slice)
	return
}

// Insert creates a new slice (array) from the given slice, with additional values inserted at the given index
func Insert[V interface{}, S ~[]V](slice S, at int, values ...V) (modified S) {
	before := slice[:at]
	after := slice[at:]
	modified = make(S, 0)
	modified = append(modified, before...)
	modified = append(modified, values...)
	modified = append(modified, after...)
	return
}

// Prune removes all instances of the specified value from a copy of the given slice
func Prune[V comparable, S ~[]V](slice S, value V) (modified S, rmIndexes []int) {
	modified = make(S, 0)
	for i, v := range slice {
		if v == value {
			rmIndexes = append(rmIndexes, i)
		} else {
			modified = append(modified, v)
		}
	}
	return
}

// Remove creates a new slice (array) from the given slice, with the specified index removed
func Remove[V interface{}, S ~[]V](slice S, at int) (modified S) {
	modified = make(S, 0)
	if at >= 0 && at < len(slice) {
		modified = append(modified, slice[:at]...)
		modified = append(modified, slice[at+1:]...)
	} else {
		modified = append(modified, slice...)
	}
	return
}

// Push appends the given value to a new copy of the given slice
func Push[V interface{}, S ~[]V](slice S, values ...V) (modified S) {
	modified = append(Copy(slice), values...)
	return
}

// Pop removes the last value from a Copy of the slice and returns it
func Pop[V interface{}, S ~[]V](slice S) (modified S, value V) {
	if last := len(slice) - 1; last > -1 {
		value = slice[last]
		modified = Truncate(slice, last)
	}
	return
}

// Shift prepends the given value to a new copy of the given slice
func Shift[V interface{}, S ~[]V](slice S, values ...V) (modified S) {
	modified = make(S, 0)
	modified = append(modified, values...)
	modified = append(modified, slice...)
	return
}

// Unshift removes the first value from a Copy of the slice and returns it
func Unshift[V interface{}, S ~[]V](slice S) (modified S, value V) {
	if len(slice) > 0 {
		value = slice[0]
		modified = Copy(slice[1:])
	}
	return
}

// IndexOf returns the first index matching the value given
func IndexOf[V comparable, S ~[]V](slice S, value V) (index int) {
	index = -1
	for idx, v := range slice {
		if v == value {
			index = idx
			return
		}
	}
	return
}

// IndexesOf returns a list of all indexes matching the value given
func IndexesOf[V comparable, S ~[]V](slice S, value V) (indexes []int) {
	for idx, v := range slice {
		if v == value {
			indexes = append(indexes, idx)
		}
	}
	return
}

// Present returns true if the search value is present in any of the other values given
func Present[V comparable](search V, others ...V) (present bool) {
	present = Within(search, others)
	return
}

// Within return true if the search value is present in any of the other slices of V given
func Within[V comparable, S ~[]V](search V, others ...S) (present bool) {
	for _, other := range others {
		for _, value := range other {
			if present = search == value; present {
				return
			}
		}
	}
	return
}

// AnyWithin returns true if any of the values in the source given are present in any of the other slices given
func AnyWithin[V comparable, S ~[]V](src S, others ...S) (present bool) {
	for _, search := range src {
		if present = Within(search, others...); present {
			return
		}
	}
	return
}