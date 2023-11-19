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

func AsInt[V Number](v V) int {
	return int(v)
}

func AsUint[V Number](v V) uint {
	return uint(v)
}

func AsInt64[V Number](v V) int64 {
	return int64(v)
}

func AsUint64[V Number](v V) uint64 {
	return uint64(v)
}

func ToInt(v interface{}, d int) int {
	switch t := v.(type) {
	case int:
		return t
	case int8:
		return int(t)
	case int16:
		return int(t)
	case int32:
		return int(t)
	case int64:
		return int(t)
	case uint:
		return int(t)
	case uint8:
		return int(t)
	case uint16:
		return int(t)
	case uint32:
		return int(t)
	case uint64:
		return int(t)
	case float32:
		return int(t)
	case float64:
		return int(t)
	default:
		return d
	}
}

func ToInt64(v interface{}, d int64) int64 {
	switch t := v.(type) {
	case int:
		return int64(t)
	case int8:
		return int64(t)
	case int16:
		return int64(t)
	case int32:
		return int64(t)
	case int64:
		return t
	case uint:
		return int64(t)
	case uint8:
		return int64(t)
	case uint16:
		return int64(t)
	case uint32:
		return int64(t)
	case uint64:
		return int64(t)
	case float32:
		return int64(t)
	case float64:
		return int64(t)
	default:
		return d
	}
}

func ToUint(v interface{}, d uint) uint {
	switch t := v.(type) {
	case int:
		return uint(t)
	case int8:
		return uint(t)
	case int16:
		return uint(t)
	case int32:
		return uint(t)
	case int64:
		return uint(t)
	case uint:
		return t
	case uint8:
		return uint(t)
	case uint16:
		return uint(t)
	case uint32:
		return uint(t)
	case uint64:
		return uint(t)
	case float32:
		return uint(t)
	case float64:
		return uint(t)
	default:
		return d
	}
}

func ToUint64(v interface{}, d uint64) uint64 {
	switch t := v.(type) {
	case int:
		return uint64(t)
	case int8:
		return uint64(t)
	case int16:
		return uint64(t)
	case int32:
		return uint64(t)
	case int64:
		return uint64(t)
	case uint:
		return uint64(t)
	case uint8:
		return uint64(t)
	case uint16:
		return uint64(t)
	case uint32:
		return uint64(t)
	case uint64:
		return t
	case float32:
		return uint64(t)
	case float64:
		return uint64(t)
	default:
		return d
	}
}

func ToFloat64(v interface{}, d float64) float64 {
	switch t := v.(type) {
	case int:
		return float64(t)
	case int8:
		return float64(t)
	case int16:
		return float64(t)
	case int32:
		return float64(t)
	case int64:
		return float64(t)
	case uint:
		return float64(t)
	case uint8:
		return float64(t)
	case uint16:
		return float64(t)
	case uint32:
		return float64(t)
	case uint64:
		return float64(t)
	case float32:
		return float64(t)
	case float64:
		return t
	default:
		return d
	}
}