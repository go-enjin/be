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

package maths

import (
	"github.com/go-enjin/be/pkg/maps"
)

func Clamp[T maps.Number](value, min, max T) T {
	if value >= min && value <= max {
		return value
	}
	if value > max {
		return max
	}
	return min
}

func Floor[T maps.Number](value, min T) T {
	if value < min {
		return min
	}
	return value
}

func Ceil[T maps.Number](value, max T) T {
	if value > max {
		return max
	}
	return value
}