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

package context

import (
	"fmt"
	"sort"

	"github.com/iancoleman/strcase"
)

type Context map[string]interface{}

func New() (ctx Context) {
	ctx = make(Context)
	return
}

func NewFromMap(m map[string]interface{}) Context {
	return Context(m)
}

func (c Context) Keys() (keys []string) {
	for k, _ := range c {
		keys = append(keys, k)
	}
	return
}

func (c Context) Copy() (ctx Context) {
	ctx = New()
	for k, v := range c {
		if sm, ok := v.(map[string]interface{}); ok {
			ctx.Set(k, Context(sm).Copy())
		} else {
			ctx.Set(k, v)
		}
	}
	return
}

func (c Context) Apply(contexts ...Context) {
	for _, cc := range contexts {
		for k, v := range cc {
			c.Set(k, v)
		}
	}
	return
}

func (c Context) Has(key string) (present bool) {
	present = c.Get(key) != nil
	return
}

func (c Context) Set(key string, value interface{}) Context {
	key = strcase.ToCamel(key)
	c[key] = value
	return c
}

func (c Context) SetSpecific(key string, value interface{}) Context {
	c[key] = value
	return c
}

func (c Context) Get(key string) interface{} {
	if v, ok := c[key]; ok {
		return v
	}
	camel := strcase.ToCamel(key)
	if v, ok := c[camel]; ok {
		return v
	}
	snake := strcase.ToSnake(key)
	if v, ok := c[snake]; ok {
		return v
	}
	return nil
}

func (c Context) Delete(key string) (deleted bool) {
	if _, ok := c[key]; ok {
		delete(c, key)
		return true
	}
	camel := strcase.ToCamel(key)
	if _, ok := c[camel]; ok {
		delete(c, camel)
		return true
	}
	snake := strcase.ToSnake(key)
	if _, ok := c[snake]; ok {
		delete(c, snake)
		return true
	}
	return true
}

func (c Context) DeleteKeys(keys ...string) {
	for _, key := range keys {
		_ = c.Delete(key)
	}
}

func (c Context) String(key, def string) string {
	if v := c.Get(key); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

func (c Context) Bool(key string, def bool) bool {
	if v := c.Get(key); v != nil {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return def
}

func (c Context) Int(key string, def int) int {
	if v := c.Get(key); v != nil {
		if i, ok := v.(int); ok {
			return i
		}
	}
	return def
}

func (c Context) Int64(key string, def int64) int64 {
	if v := c.Get(key); v != nil {
		if i, ok := v.(int64); ok {
			return i
		}
	}
	return def
}

func (c Context) Float64(key string, def float64) float64 {
	if v := c.Get(key); v != nil {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return def
}

func (c Context) AsMap() (out map[string]interface{}) {
	out = make(map[string]interface{})
	for k, v := range c {
		out[k] = v
	}
	return
}

func (c Context) AsMapStrings() (out map[string]string) {
	out = make(map[string]string)
	for k, v := range c {
		switch t := v.(type) {
		case string:
			out[k] = t
		default:
			out[k] = fmt.Sprintf("%v", t)
		}
	}
	return
}

func (c Context) AsLogString() (out string) {
	keys := c.Keys()
	sort.Strings(keys)
	out = "["
	for idx, key := range keys {
		if idx > 0 {
			out += ", "
		}
		if s, ok := c[key].(string); ok {
			out += key + ": \"" + s + "\""
			continue
		}
		if s, ok := c[key].(fmt.Stringer); ok {
			out += key + ": \"" + s.String() + "\""
			continue
		}
		out += fmt.Sprintf("%v: %#v", key, c[key])
	}
	out += "]"
	return
}