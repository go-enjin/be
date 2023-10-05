//go:build page_funcmaps || pages || all

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

package dict

import (
	"fmt"
	"strings"

	"github.com/samber/lo"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/maps"
)

type Dictionary map[string]interface{}

// NewDictionary constructs a new Dictionary which is a simple wrapper around a context with method signatures that work
// with the constraints of providing go template functions; uses .SET to handle argv
func NewDictionary(argv ...interface{}) (d Dictionary, err error) {
	if argc := len(argv); argc > 0 && argc%2 != 0 {
		if ctx, ok := argv[0].(Dictionary); ok {
			d, err = WrapDictionary(ctx, argv[1:]...)
			return
		} else if ctx, ok := argv[0].(beContext.Context); ok {
			d, err = WrapDictionary(ctx, argv[1:]...)
			return
		} else if ctx, ok := argv[0].(map[string]interface{}); ok {
			d, err = WrapDictionary(ctx, argv[1:]...)
			return
		}
		err = fmt.Errorf("expected Dictionary, beContext.Context{} or map[string]interface{}, received: %T", argv[0])
		return
	}
	d = Dictionary{}
	_, err = d.SET(argv...)
	return
}

func WrapDictionary(ctx map[string]interface{}, argv ...interface{}) (d Dictionary, err error) {
	d = lo.Assign(Dictionary{}, ctx)
	_, err = d.SET(argv...)
	return
}

// SET
//   - accepts a list of key/value argument pairs, returns error if list is odd
//   - keys can be strings or string slices:
//   - string slices set deep values, ie: ["one","two","three"] produces map[one]map[two]map[three]=value
//   - string keys also support enjin context deep keys, ie: .One.Two.Three is the same as the slice example above
func (d Dictionary) SET(argv ...interface{}) (nop string, err error) {
	var argc int
	if argc = len(argv); argc%2 != 0 {
		err = fmt.Errorf("odd number of dictionary key/value arguments")
		return
	}

	for i := 0; i < argc; i += 2 {

		switch arg := argv[i].(type) {

		case string:
			// set top-level key
			if err = beContext.SetKV(d, arg, argv[i+1]); err != nil {
				return
			}

		case []string:
			// slice is list of keys joined into a Deep-Key
			if err = beContext.SetKV(d, "."+strings.Join(arg, "."), argv[i+1]); err != nil {
				return
			}

		default:
			err = fmt.Errorf("invalid key argument type: %T", arg)
			return

		}

	}

	return
}

func (d Dictionary) GET(key string, def ...interface{}) (value interface{}) {
	if _, value = beContext.GetKV(d, key); value == nil && len(def) > 0 {
		value = def[0]
	}
	return
}

func (d Dictionary) DELETE(key string) (nop string) {
	beContext.DeleteKV(d, key)
	return
}

func (d Dictionary) KEYS() (keys []string) {
	keys = maps.SortedKeys(d)
	return
}