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

package kvm

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"
)

type Value struct {
	Time     *time.Time     `json:"time,omitempty"`
	Duration *time.Duration `json:"duration,omitempty"`

	Bool    *bool    `json:"bool,omitempty"`
	String  *string  `json:"string,omitempty"`
	Int     *int     `json:"int,omitempty"`
	Int8    *int8    `json:"int8,omitempty"`
	Int16   *int16   `json:"int16,omitempty"`
	Int32   *int32   `json:"int32,omitempty"`
	Int64   *int64   `json:"int64,omitempty"`
	Uint    *uint    `json:"uint,omitempty"`
	Uint8   *uint8   `json:"uint8,omitempty"`
	Uint16  *uint16  `json:"uint16,omitempty"`
	Uint32  *uint32  `json:"uint32,omitempty"`
	Uint64  *uint64  `json:"uint64,omitempty"`
	Float32 *float32 `json:"float32,omitempty"`
	Float64 *float64 `json:"float64,omitempty"`

	SliceBool    *[]bool    `json:"slice-bool,omitempty"`
	SliceString  *[]string  `json:"slice-string,omitempty"`
	SliceInt     *[]int     `json:"slice-int,omitempty"`
	SliceInt8    *[]int8    `json:"slice-int8,omitempty"`
	SliceInt16   *[]int16   `json:"slice-int16,omitempty"`
	SliceInt32   *[]int32   `json:"slice-int32,omitempty"`
	SliceInt64   *[]int64   `json:"slice-int64,omitempty"`
	SliceUint    *[]uint    `json:"slice-uint,omitempty"`
	SliceUint8   *[]uint8   `json:"slice-uint8,omitempty"`
	SliceUint16  *[]uint16  `json:"slice-uint16,omitempty"`
	SliceUint32  *[]uint32  `json:"slice-uint32,omitempty"`
	SliceUint64  *[]uint64  `json:"slice-uint64,omitempty"`
	SliceFloat32 *[]float32 `json:"slice-float32,omitempty"`
	SliceFloat64 *[]float64 `json:"slice-float64,omitempty"`

	SliceInterface *[]interface{} `json:"slice-interface,omitempty"`
	Interface      *interface{}   `json:"interface,omitempty"`
}

func NewValue(from interface{}) (value *Value) {
	value = new(Value)
	switch t := from.(type) {

	case time.Time:
		value.Time = &t
	case time.Duration:
		value.Duration = &t

	case string:
		value.String = &t
	case int:
		value.Int = &t
	case int8:
		value.Int8 = &t
	case int16:
		value.Int16 = &t
	case int32:
		value.Int32 = &t
	case int64:
		value.Int64 = &t
	case uint:
		value.Uint = &t
	case uint8:
		value.Uint8 = &t
	case uint16:
		value.Uint16 = &t
	case uint32:
		value.Uint32 = &t
	case uint64:
		value.Uint64 = &t
	case float32:
		value.Float32 = &t
	case float64:
		value.Float64 = &t

	case []string:
		value.SliceString = &t
	case []int:
		value.SliceInt = &t
	case []int8:
		value.SliceInt8 = &t
	case []int16:
		value.SliceInt16 = &t
	case []int32:
		value.SliceInt32 = &t
	case []int64:
		value.SliceInt64 = &t
	case []uint:
		value.SliceUint = &t
	case []uint8:
		value.SliceUint8 = &t
	case []uint16:
		value.SliceUint16 = &t
	case []uint32:
		value.SliceUint32 = &t
	case []uint64:
		value.SliceUint64 = &t
	case []float32:
		value.SliceFloat32 = &t
	case []float64:
		value.SliceFloat64 = &t

	case []interface{}:
		value.SliceInterface = &t
	case interface{}:
		value.Interface = &t
	}
	return
}

var rxIsNumber = regexp.MustCompile(`^\s*([.\d]+)\s*$`)

func NewValueFromTypeData(vType, vData string) (value *Value, err error) {
	var encoded string
	if rxIsNumber.MatchString(vData) {
		encoded = fmt.Sprintf(`{"%v": %v}`, vType, vData)
	} else {
		encoded = fmt.Sprintf(`{"%v": "%v"}`, vType, vData)
	}
	var v Value
	err = v.UnmarshalBinary([]byte(encoded))
	value = &v
	return
}

func (v Value) MarshalBinary() ([]byte, error) {
	return json.Marshal(v)
}

func (v *Value) UnmarshalBinary(b []byte) error {
	return json.Unmarshal(b, v)
}

func (v *Value) Get() (value interface{}) {
	switch {

	case v.Time != nil:
		return *v.Time
	case v.Duration != nil:
		return *v.Duration

	case v.Bool != nil:
		return *v.Bool
	case v.String != nil:
		return *v.String
	case v.Int != nil:
		return *v.Int
	case v.Int8 != nil:
		return *v.Int8
	case v.Int16 != nil:
		return *v.Int16
	case v.Int32 != nil:
		return *v.Int32
	case v.Int64 != nil:
		return *v.Int64
	case v.Uint != nil:
		return *v.Uint
	case v.Uint8 != nil:
		return *v.Uint8
	case v.Uint16 != nil:
		return *v.Uint16
	case v.Uint32 != nil:
		return *v.Uint32
	case v.Uint64 != nil:
		return *v.Uint64
	case v.Float32 != nil:
		return *v.Float32
	case v.Float64 != nil:
		return *v.Float64

	case v.SliceBool != nil:
		return *v.SliceBool
	case v.SliceString != nil:
		return *v.SliceString
	case v.SliceInt != nil:
		return *v.SliceInt
	case v.SliceInt8 != nil:
		return *v.SliceInt8
	case v.SliceInt16 != nil:
		return *v.SliceInt16
	case v.SliceInt32 != nil:
		return *v.SliceInt32
	case v.SliceInt64 != nil:
		return *v.SliceInt64
	case v.SliceUint != nil:
		return *v.SliceUint
	case v.SliceUint8 != nil:
		return *v.SliceUint8
	case v.SliceUint16 != nil:
		return *v.SliceUint16
	case v.SliceUint32 != nil:
		return *v.SliceUint32
	case v.SliceUint64 != nil:
		return *v.SliceUint64
	case v.SliceFloat32 != nil:
		return *v.SliceFloat32
	case v.SliceFloat64 != nil:
		return *v.SliceFloat64

	case v.SliceInterface != nil:
		return *v.SliceInterface
	case v.Interface != nil:
		return *v.Interface

	}
	return
}

func (v *Value) GetKV() (key string, value interface{}) {
	switch {
	case v.Time != nil:
		return "time", *v.Time
	case v.Duration != nil:
		return "duration", *v.Duration
	case v.Bool != nil:
		return "bool", *v.Bool
	case v.String != nil:
		return "string", *v.String
	case v.Int != nil:
		return "int", *v.Int
	case v.Int8 != nil:
		return "int8", *v.Int8
	case v.Int16 != nil:
		return "int16", *v.Int16
	case v.Int32 != nil:
		return "int32", *v.Int32
	case v.Int64 != nil:
		return "int64", *v.Int64
	case v.Uint != nil:
		return "uint", *v.Uint
	case v.Uint8 != nil:
		return "uint8", *v.Uint8
	case v.Uint16 != nil:
		return "uint16", *v.Uint16
	case v.Uint32 != nil:
		return "uint32", *v.Uint32
	case v.Uint64 != nil:
		return "uint64", *v.Uint64
	case v.Float32 != nil:
		return "float32", *v.Float32
	case v.Float64 != nil:
		return "float64", *v.Float64
	case v.SliceBool != nil:
		return "slice-bool", *v.SliceBool
	case v.SliceString != nil:
		return "slice-string", *v.SliceString
	case v.SliceInt != nil:
		return "slice-int", *v.SliceInt
	case v.SliceInt8 != nil:
		return "slice-int8", *v.SliceInt8
	case v.SliceInt16 != nil:
		return "slice-int16", *v.SliceInt16
	case v.SliceInt32 != nil:
		return "slice-int32", *v.SliceInt32
	case v.SliceInt64 != nil:
		return "slice-int64", *v.SliceInt64
	case v.SliceUint != nil:
		return "slice-uint", *v.SliceUint
	case v.SliceUint8 != nil:
		return "slice-uint8", *v.SliceUint8
	case v.SliceUint16 != nil:
		return "slice-uint16", *v.SliceUint16
	case v.SliceUint32 != nil:
		return "slice-uint32", *v.SliceUint32
	case v.SliceUint64 != nil:
		return "slice-uint64", *v.SliceUint64
	case v.SliceFloat32 != nil:
		return "slice-float32", *v.SliceFloat32
	case v.SliceFloat64 != nil:
		return "slice-float64", *v.SliceFloat64
	case v.SliceInterface != nil:
		return "slice-interface", *v.SliceInterface
	case v.Interface != nil:
		return "interface", *v.Interface
	}

	return
}