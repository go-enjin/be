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

	"github.com/fvbommel/sortorder"
	"github.com/iancoleman/strcase"
)

type RequestKey string

// Context is a wrapper around a map[string]interface{} structure which is used
// throughout Go-Enjin for parsing configurations and contents.
type Context map[string]interface{}

// New constructs a new Context instance
func New() (ctx Context) {
	ctx = make(Context)
	return
}

// NewFromMap casts an existing map[string]interface{} as a Context
func NewFromMap(m map[string]interface{}) Context {
	return Context(m)
}

// Keys returns a list of all the map keys in the Context, sorted in natural
// order for consistency
func (c Context) Keys() (keys []string) {
	for k, _ := range c {
		keys = append(keys, k)
	}
	sort.Sort(sortorder.Natural(keys))
	return
}

// TODO: make context.Copy a deeper Copy() or maybe literal types only

// Copy makes a duplicate of this Context
func (c Context) Copy() (ctx Context) {
	ctx = New()
	for k, v := range c {
		if sm, ok := v.(map[string]interface{}); ok {
			ctx.SetSpecific(k, Context(sm).Copy())
		} else {
			ctx.SetSpecific(k, v)
		}
	}
	return
}

// Apply takes a list of contexts and merges their contents into this one
func (c Context) Apply(contexts ...Context) {
	for _, cc := range contexts {
		for k, v := range cc {
			c.Set(k, v)
		}
	}
	return
}

// Has returns true if the given Context key exists
func (c Context) Has(key string) (present bool) {
	present = c.Get(key) != nil
	return
}

// Set CamelCases the given key and sets that within this Context
func (c Context) Set(key string, value interface{}) Context {
	key = strcase.ToCamel(key)
	c[key] = value
	return c
}

// SetSpecific is like Set(), without CamelCasing the key
func (c Context) SetSpecific(key string, value interface{}) Context {
	c[key] = value
	return c
}

// Get returns the given key's value as an interface{} and returns nil if not
// found. Get looks for the key as given first and if not found looks for a
// CamelCased version of the key and if still not found looks for a kebab-ified
// version, finally if nothing is found, nil is returned.
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

// Delete deletes the given key from the Context and follows a similar key
// lookup process to Get() for finding the key to delete and will only delete
// the first matching key format (specific, Camel, kebab) found
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

// DeleteKeys is a batch wrapper around Delete()
func (c Context) DeleteKeys(keys ...string) {
	for _, key := range keys {
		_ = c.Delete(key)
	}
}

// String returns the key's value as a string, returning the given default if
// not found or not actually a string value.
func (c Context) String(key, def string) string {
	if v := c.Get(key); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

// StringOrStrings returns the key's value as a list of strings and if the key's
// actual value is not a list of strings, return that as a list of one string
func (c Context) StringOrStrings(key string) (values []string) {
	if v := c.Get(key); v != nil {
		if s, ok := v.(string); ok {
			values = []string{s}
			return
		}
		if vi, ok := v.([]interface{}); ok {
			for _, i := range vi {
				if s, ok := i.(string); ok {
					values = append(values, s)
				}
			}
			return
		}
	}
	return
}

// Strings returns the key's value as a list of strings, returning an empty list
// if not found or not actually a list of strings
func (c Context) Strings(key string) (values []string) {
	if v := c.Get(key); v != nil {
		if vs, ok := v.([]string); ok {
			values = vs
			return
		}
	}
	return
}

// DefaultStrings is a wrapper around Strings() and returns the given default
// list of strings if the key is not found
func (c Context) DefaultStrings(key string, def []string) []string {
	if v := c.Get(key); v != nil {
		if s, ok := v.([]string); ok {
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

// AsMap returns this Context, shallowly copied, as a new map[string]interface{}
// instance.
func (c Context) AsMap() (out map[string]interface{}) {
	out = make(map[string]interface{})
	for k, v := range c {
		out[k] = v
	}
	return
}

// AsMapStrings returns this Context as a transformed map[string]string
// structure, where each key's value is checked and if it's a string, use it
// as-is and if it's anything else, run it through fmt.Sprintf("%v") to make it
// a string.
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

// CamelizeKeys transforms all keys within the Context to be of CamelCased form
func (c Context) CamelizeKeys() {
	var remove []string
	for k, v := range c {
		camelized := strcase.ToCamel(k)
		if k != camelized {
			remove = append(remove, k)
			c.SetSpecific(camelized, v)
		}
	}
	c.DeleteKeys(remove...)
}