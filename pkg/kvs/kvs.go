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

	"github.com/go-corelibs/maths"
	"github.com/go-enjin/be/pkg/feature"
)

type typedNil struct {
	IS string
}

var (
	gNilValue = typedNil{IS: "nil"}
	gNilBytes = "\x1c\x7f\x03\x01\x01\btypedNil\x01\xff\x80\x00\x01\x01\x01\x02IS\x01\f\x00\x00\x00\b\xff\x80\x01\x03nil\x00"
)

type Variables interface {
	maths.Number | byte | string
}

func IsSet(store feature.KeyValueStore, key string) (present bool) {
	if v, err := store.Get(key); err == nil {
		present = len(v) > 0 && string(v) != gNilBytes
	}
	return
}

func Marshal(value interface{}) (data []byte, err error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err = enc.Encode(value); err != nil {
		return
	}
	data = buf.Bytes()
	return
}

func SetMarshal(store feature.KeyValueStore, key string, value interface{}) (err error) {
	var data []byte
	if data, err = Marshal(value); err != nil {
		return
	}
	err = store.Set(key, data)
	return
}

func Unmarshal[T interface{}](data []byte, value *T) (err error) {
	dec := gob.NewDecoder(bytes.NewBuffer(data))
	err = dec.Decode(value)
	return
}

func GetUnmarshal[T interface{}](store feature.KeyValueStore, key string, value *T) (err error) {
	var data []byte
	if data, err = store.Get(key); err != nil {
		return
	}
	err = Unmarshal(data, value)
	return
}

func GetValue[T interface{}](store feature.KeyValueStore, key string) (value T) {
	_ = GetUnmarshal(store, key, &value)
	return
}

func AddToNumber[T maths.Number](store feature.KeyValueStore, key string, increment T) (updated T, err error) {
	var current T
	_ = GetUnmarshal(store, key, &current)
	err = SetMarshal(store, key, current+increment)
	return
}

func GetIsNil(store feature.KeyValueStore, key string) (isNil bool) {
	var tnv typedNil
	if e := GetUnmarshal(store, key, &tnv); e == nil {
		isNil = tnv == gNilValue
	}
	return
}
