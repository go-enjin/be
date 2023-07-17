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

package cmp

import (
	"fmt"
	"time"
)

func Cmp[T comparable](a, b interface{}) (same bool, err error) {
	if ac, ok := a.(T); ok {
		if bc, ok := b.(T); ok {
			same = ac == bc
			return
		}
	}
	err = fmt.Errorf("error cmp.Cmp - inconsistent types: %T vs %T", a, b)
	return
}

func Compare(a, b interface{}) (same bool, err error) {
	if fmt.Sprintf("%T", a) != fmt.Sprintf("%T", b) {
		err = fmt.Errorf("incompatible types for comparison: %T vs %T", a, b)
		return
	}
	switch ta := a.(type) {
	case nil,
		string,
		float32, float64,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		switch tb := b.(type) {
		case nil,
			string,
			float32, float64,
			int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64:
			same = ta == tb
		}
	case time.Time:
		same = ta.Unix() == b.(time.Time).Unix()
	case time.Duration:
		same = ta.String() == b.(time.Duration).String()
	default:
		err = fmt.Errorf("unsupported type for comparison: %T", a)
	}
	return
}

func Less(a, b interface{}) (less bool, err error) {
	if fmt.Sprintf("%T", a) != fmt.Sprintf("%T", b) {
		err = fmt.Errorf("incompatible types for (less) comparison: %T vs %T", a, b)
		return
	}
	switch ta := a.(type) {
	case nil:
		less = false

	case string:
		less = ta < b.(string)

	case float32:
		less = ta < b.(float32)

	case float64:
		less = ta < b.(float64)

	case int:
		less = ta < b.(int)

	case int8:
		less = ta < b.(int8)

	case int16:
		less = ta < b.(int16)

	case int32:
		less = ta < b.(int32)

	case int64:
		less = ta < b.(int64)

	case uint:
		less = ta < b.(uint)

	case uint8:
		less = ta < b.(uint8)

	case uint16:
		less = ta < b.(uint16)

	case uint32:
		less = ta < b.(uint32)

	case uint64:
		less = ta < b.(uint64)

	case time.Time:
		less = ta.Before(b.(time.Time))

	case time.Duration:
		less = ta.Seconds() < b.(time.Duration).Seconds()

	default:
		err = fmt.Errorf("unsupported type for (less) comparison: %T", a)
	}
	return
}