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

package gob

import (
	"bytes"
	"encoding/gob"

	"github.com/go-enjin/golang-org-x-text/language"
)

type wrapper struct {
	Value interface{}
}

func Register(v interface{}) {
	gob.Register(v)
	gob.Register(language.Tag{})
}

func Encode(v interface{}) (data []byte, err error) {
	var b bytes.Buffer
	e := gob.NewEncoder(&b)
	w := wrapper{
		Value: v,
	}
	if err = e.Encode(&w); err != nil {
		return
	}
	data = b.Bytes()
	return
}

func Decode(data []byte) (v interface{}, err error) {
	var b bytes.Buffer
	e := gob.NewDecoder(&b)
	b.Write(data)
	var w wrapper
	if err = e.Decode(&w); err != nil {
		return
	}
	v = w.Value
	return
}