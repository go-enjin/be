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

package funcs

import (
	"strconv"

	"github.com/go-enjin/be/pkg/log"
)

func Int(v interface{}) (i int64) {
	switch t := v.(type) {
	case int:
		i = int64(t)
	case int8:
		i = int64(t)
	case int16:
		i = int64(t)
	case int32:
		i = int64(t)
	case int64:
		i = t
	case uint:
		i = int64(t)
	case uint8:
		i = int64(t)
	case uint16:
		i = int64(t)
	case uint32:
		i = int64(t)
	case uint64:
		i = int64(t)
	case float32:
		i = int64(t)
	case float64:
		i = int64(t)
	case string:
		if vv, err := strconv.Atoi(t); err != nil {
			i = int64(vv)
		}
	}
	return
}

func Float(v interface{}) (i float64) {
	switch t := v.(type) {
	case int:
		i = float64(t)
	case int8:
		i = float64(t)
	case int16:
		i = float64(t)
	case int32:
		i = float64(t)
	case int64:
		i = float64(t)
	case uint:
		i = float64(t)
	case uint8:
		i = float64(t)
	case uint16:
		i = float64(t)
	case uint32:
		i = float64(t)
	case uint64:
		i = float64(t)
	case float32:
		i = float64(t)
	case float64:
		i = t
	case string:
		if vv, err := strconv.ParseFloat(t, 64); err != nil {
			i = vv
		}
	}
	return
}

func Add(values ...interface{}) (result int64) {
	for _, v := range values {
		value := Int(v)
		result += value
	}
	return
}

func Sub(values ...interface{}) (result int64) {
	if len(values) == 0 {
		return
	}
	for idx, v := range values {
		value := Int(v)
		if idx == 0 {
			result = value
		} else {
			result -= value
		}
	}
	return
}

func Mul(a, b interface{}) (result int64) {
	result = Int(a) * Int(b)
	return
}

func Div(a, b interface{}) (result int64) {
	ia, ib := Int(a), Int(b)
	if ib == 0 {
		log.WarnF("caught template divide by zero: %d / %d (%v / %v)", ia, ib, a, b)
		return
	}
	result = ia / ib
	return
}

func Mod(a, b interface{}) (result int64) {
	result = Int(a) % Int(b)
	return
}

func AddFloat(values ...interface{}) (result float64) {
	for _, v := range values {
		value := Float(v)
		result += value
	}
	return
}

func SubFloat(values ...interface{}) (result float64) {
	if len(values) == 0 {
		return
	}
	for idx, v := range values {
		value := Float(v)
		if idx == 0 {
			result = value
		} else {
			result -= value
		}
	}
	return
}

func MulFloat(a, b interface{}) (result float64) {
	result = Float(a) * Float(b)
	return
}

func DivFloat(a, b interface{}) (result float64) {
	ia, ib := Float(a), Float(b)
	if ib == 0 {
		log.WarnF("caught template divide by zero: %d / %d (%v / %v)", ia, ib, a, b)
		return
	}
	result = ia / ib
	return
}