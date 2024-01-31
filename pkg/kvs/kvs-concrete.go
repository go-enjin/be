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

package kvs

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strconv"
	"time"

	"github.com/go-corelibs/slices"
	beErrors "github.com/go-enjin/be/pkg/errors"
)

type ConcreteTag uint8

const (
	StartConcreteTags ConcreteTag = iota + 1
	CString
	CFloat32
	CFloat64
	CInt
	CInt8
	CInt16
	CInt32
	CInt64
	CUint
	CUint8
	CUint16
	CUint32
	CUint64
	CTimeTime
	CTimeDuration
	CStringSlice
	CInterfaceSlice
	EndConcreteTags
)

func (t ConcreteTag) Valid() (valid bool) {
	valid = StartConcreteTags < t && t < EndConcreteTags
	return
}

func (t ConcreteTag) String() (value string) {
	value = strconv.Itoa(int(t))
	return
}

func GetConcreteTag(value interface{}) (tag ConcreteTag) {
	switch t := value.(type) {
	case *string, string:
		tag = CString
	case float32:
		tag = CFloat32
	case float64:
		tag = CFloat64
	case int:
		tag = CInt
	case int8:
		tag = CInt8
	case int16:
		tag = CInt16
	case int32:
		tag = CInt32
	case int64:
		tag = CInt64
	case uint:
		tag = CUint
	case uint8:
		tag = CUint8
	case uint16:
		tag = CUint16
	case uint32:
		tag = CUint32
	case uint64:
		tag = CUint64
	case time.Time:
		tag = CTimeTime
	case time.Duration:
		tag = CTimeDuration
	case []string:
		tag = CStringSlice
	case []interface{}:
		var ok bool
		if _, ok = slices.Retype[string](t); ok {
		} else if _, ok = slices.Retype[int](t); ok {
		} else if _, ok = slices.Retype[uint64](t); ok {
		} else if _, ok = slices.Retype[float64](t); ok {
		}
		if ok {
			tag = CInterfaceSlice
		}
	}
	return
}

func MarshalConcrete(value interface{}) (data string, err error) {
	var tag ConcreteTag
	if tag = GetConcreteTag(value); !tag.Valid() {
		err = fmt.Errorf("not a supported concrete type: %T", value)
		return
	}
	var bData []byte
	switch t := value.(type) {
	case string, *string,
		float32, float64,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		time.Time, time.Duration, []string, []interface{}:
		if bData, err = Marshal(t); err != nil {
			return
		}
	default:
		err = fmt.Errorf("unexpected type: %T %q", t, t)
		return
	}
	data = string(append([]byte("c\t"+tag.String()+"\n"), bData...))
	return
}

func decoder[T interface{}](data []byte) (value interface{}, err error) {
	var v T
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	if err = dec.Decode(&v); err != nil {
		return
	}
	value = v
	return
}

func UnmarshalConcrete[V string | []byte](input V) (value interface{}, err error) {
	data := []byte(input)
	size := len(data)
	if size == 0 {
		return
	} else if size < 3 || data[0] != 'c' || data[1] != '\t' {
		err = beErrors.ErrDataTypeNotSupported
		return
	}

	var end int
	var tagString string
	for idx, char := range data {
		if idx < 2 {
			continue
		} else if char == '\n' {
			end = idx
			break
		}
		tagString += string(char)
	}
	if end+1 < size {
		data = data[end+1:]
	} else {
		data = []byte{}
	}
	var tag ConcreteTag
	if i, ee := strconv.Atoi(tagString); ee != nil {
		err = beErrors.ErrDataTypeNotSupported
		return
	} else if tag = ConcreteTag(i); !tag.Valid() {
		err = beErrors.ErrDataTypeNotSupported
		return
	}

	switch tag {
	case CString:
		value, err = decoder[string](data)
	case CFloat32:
		value, err = decoder[float32](data)
	case CFloat64:
		value, err = decoder[float64](data)
	case CInt:
		value, err = decoder[int](data)
	case CInt8:
		value, err = decoder[int8](data)
	case CInt16:
		value, err = decoder[int16](data)
	case CInt32:
		value, err = decoder[int32](data)
	case CInt64:
		value, err = decoder[int64](data)
	case CUint:
		value, err = decoder[uint](data)
	case CUint8:
		value, err = decoder[uint8](data)
	case CUint16:
		value, err = decoder[uint16](data)
	case CUint32:
		value, err = decoder[uint32](data)
	case CUint64:
		value, err = decoder[uint64](data)
	case CTimeTime:
		value, err = decoder[time.Time](data)
	case CTimeDuration:
		value, err = decoder[time.Duration](data)
	case CStringSlice:
		value, err = decoder[[]string](data)
	case CInterfaceSlice:
		value, err = decoder[[]interface{}](data)
	default:
		err = beErrors.ErrDataTypeNotSupported
		return
	}

	return
}
