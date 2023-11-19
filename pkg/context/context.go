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
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/maruel/natural"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"

	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/maths"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

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

// NewFromOsEnviron constructs a new Context from os.Environ() string K=V slices
func NewFromOsEnviron(slices ...[]string) (c Context) {
	c = New()
	for _, slice := range slices {
		for _, pair := range slice {
			if key, value, ok := strings.Cut(pair, "="); ok {
				c.SetSpecific(key, beStrings.TrimQuotes(value))
			}
		}
	}
	return
}

// Len returns the number of keys in the Context
func (c Context) Len() (count int) {
	count = len(c)
	return
}

// Empty returns true if there is nothing stored in the Context
func (c Context) Empty() (empty bool) {
	empty = c.Len() == 0
	return
}

// Keys returns a list of all the map keys in the Context, sorted in natural
// order for consistency
func (c Context) Keys() (keys []string) {
	for k, _ := range c {
		keys = append(keys, k)
	}
	sort.Sort(natural.StringSlice(keys))
	return
}

// Copy makes a deep-copy of this Context
func (c Context) Copy() (ctx Context) {
	ctx = maps.DeepCopy(c)
	return
}

// DeepKeys returns a list of .deep.keys for the entire context structure
func (c Context) DeepKeys() (keys []string) {
	for k, v := range c {
		dk := "." + k
		keys = append(keys, dk)
		switch t := v.(type) {
		case Context:
			for _, deeper := range t.DeepKeys() {
				keys = append(keys, dk+deeper)
			}
		case map[string]interface{}:
			for _, deeper := range Context(t).DeepKeys() {
				keys = append(keys, dk+deeper)
			}
		}
	}
	sort.Sort(natural.StringSlice(keys))
	return
}

// AsDeepKeyed returns a deep-key flattened version of this context
// For example:
//
//	map[string]interface{}{"one": map[string]interface{}{"two": "deep"}}
//
// becomes:
//
//	map[string]interface{}{".one.two": "deep"}
func (c Context) AsDeepKeyed() (ctx Context) {
	ctx = Context{}
	for _, k := range c.Keys() {
		dk := "." + k
		switch t := c[k].(type) {
		case Context:
			for deeperKey, deeperValue := range t.AsDeepKeyed() {
				ctx[dk+deeperKey] = deeperValue
			}
		case map[string]interface{}:
			for deeperKey, deeperValue := range Context(t).AsDeepKeyed() {
				ctx[dk+deeperKey] = deeperValue
			}
		default:
			ctx[dk] = t
		}
	}
	return
}

// Apply takes a list of contexts and merges their contents into this one
func (c Context) Apply(contexts ...Context) {
	for _, cc := range contexts {
		if cc != nil {
			for k, v := range cc {
				c.Set(k, v)
			}
		}
	}
	return
}

// ApplySpecific takes a list of contexts and merges their contents into this one, keeping the keys specifically
func (c Context) ApplySpecific(contexts ...Context) {
	for _, cc := range contexts {
		if cc != nil {
			for k, v := range cc {
				c.SetSpecific(k, v)
			}
		}
	}
	return
}

// Has returns true if the given Context key exists and is not nil
func (c Context) Has(key string) (present bool) {
	present = c.Get(key) != nil
	return
}

// HasExact returns true if the specific Context key given exists and is not nil
func (c Context) HasExact(key string) (present bool) {
	if v, ok := c[key]; ok {
		present = v != nil
	}
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

// SetKV allows for .Deep.Variable.Names
func (c Context) SetKV(key string, value interface{}) (err error) {
	err = SetKV(c, key, value)
	return
}

// Get is a convenience wrapper around GetKV
func (c Context) Get(key string) (value interface{}) {
	_, value = c.GetKV(key)
	return
}

// GetKV looks for the key as given first and if not found looks for CamelCased, kebab-case and snake_cased variations;
// returning the actual key found and the generic value; returns an empty key and nil value if nothing found at all
func (c Context) GetKV(key string) (k string, v interface{}) {
	k, v = GetKV(c, key)
	return
}

// Delete deletes the given key from the Context and follows a similar key
// lookup process to Get() for finding the key to delete and will only delete
// the first matching key format (specific, Camel, kebab) found
func (c Context) Delete(key string) (deleted bool) {
	return DeleteKV(c, key)
}

// DeleteKeys is a batch wrapper around Delete()
func (c Context) DeleteKeys(keys ...string) {
	for _, key := range keys {
		_ = c.Delete(key)
	}
}

// Bytes returns the key's value as a byte slice, returning the given default if
// not found or not actually a byte slice value.
func (c Context) Bytes(key string, def []byte) []byte {
	if v := c.Get(key); v != nil {
		if s, ok := v.([]byte); ok {
			return s
		}
	}
	return def
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
		} else if vi, ok := v.([]interface{}); ok {
			for _, vii := range vi {
				if viis, ok := vii.(string); ok {
					values = append(values, viis)
				}
			}
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

func (c Context) Boolean(key string) (value, ok bool) {
	v := c.Get(key)
	if ok = v != nil; ok {
		value = maps.ExtractBoolValue(v)
	}
	return
}

func (c Context) ValueAsInt(key string, def int) int {
	if v := c.Get(key); v != nil {
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
		case float32:
			return int(t)
		case float64:
			return int(t)
		case string:
			i, _ := strconv.Atoi(t)
			return i
		case []byte:
			i, _ := strconv.Atoi(string(t))
			return i
		}
	}
	return def
}

func (c Context) Int(key string, def int) int {
	if v := c.Get(key); v != nil {
		return maths.ToInt(v, def)
	}
	return def
}

func (c Context) Int64(key string, def int64) int64 {
	if v := c.Get(key); v != nil {
		return maths.ToInt64(v, def)
	}
	return def
}

func (c Context) Uint(key string, def uint) uint {
	if v := c.Get(key); v != nil {
		return maths.ToUint(v, def)
	}
	return def
}

func (c Context) Uint64(key string, def uint64) uint64 {
	if v := c.Get(key); v != nil {
		return maths.ToUint64(v, def)
	}
	return def
}

func (c Context) Float64(key string, def float64) float64 {
	if v := c.Get(key); v != nil {
		return maths.ToFloat64(v, def)
	}
	return def
}

// Context looks for the given key and if the value is of Context type, returns it
func (c Context) Context(key string) (ctx Context) {
	if v := c.Get(key); v != nil {
		var ok bool
		if ctx, ok = v.(Context); !ok {
			ctx, _ = v.(map[string]interface{})
		}
	}
	if ctx == nil {
		ctx = New()
	}
	return
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

// AsOsEnviron returns this Context as a transformed []string slice where each
// key is converted to SCREAMING_SNAKE_CASE and the value is converted to a
// string (similarly to AsMapStrings) and the key/value pair is concatenated
// into a single "K=V" string and appended to the output slice, sorted by key in
// natural order, suitable for use in os.Environ cases.
func (c Context) AsOsEnviron() (out []string) {
	data := c.AsMapStrings()
	var keys []string
	for k, _ := range data {
		key := strcase.ToScreamingSnake(k)
		keys = append(keys, key)
	}
	sort.Sort(natural.StringSlice(keys))
	for _, k := range keys {
		v := data[k]
		out = append(out, k+"="+v)
	}
	return
}

// CamelizeKeys transforms all keys within the Context to be of CamelCased form
func (c Context) CamelizeKeys() {
	var remove []string
	for k, v := range c {
		if vc, ok := v.(Context); ok {
			vc.CamelizeKeys()
		} else if vm, ok := v.(map[string]interface{}); ok {
			Context(vm).CamelizeKeys()
		}
		if modified := strcase.ToCamel(k); k != modified {
			remove = append(remove, k)
			c.SetSpecific(modified, v)
		}
	}
	c.DeleteKeys(remove...)
}

// LowerCamelizeKeys transforms all keys within the Context to be of lowerCamelCased form
func (c Context) LowerCamelizeKeys() {
	var remove []string
	for k, v := range c {
		if vc, ok := v.(Context); ok {
			vc.LowerCamelizeKeys()
		} else if vm, ok := v.(map[string]interface{}); ok {
			Context(vm).LowerCamelizeKeys()
		}
		if modified := strcase.ToLowerCamel(k); k != modified {
			remove = append(remove, k)
			c.SetSpecific(modified, v)
		}
	}
	c.DeleteKeys(remove...)
}

// KebabKeys transforms all keys within the Context to be of kebab-cased form
func (c Context) KebabKeys() {
	var remove []string
	for k, v := range c {
		if vc, ok := v.(Context); ok {
			vc.KebabKeys()
		} else if vm, ok := v.(map[string]interface{}); ok {
			Context(vm).KebabKeys()
		}
		if modified := strcase.ToKebab(k); k != modified {
			remove = append(remove, k)
			c.SetSpecific(modified, v)
		}
	}
	c.DeleteKeys(remove...)
}

func (c Context) Select(keys ...string) (selected map[string]interface{}) {
	selected = make(map[string]interface{})
	for _, key := range keys {
		if v, ok := c[key]; ok {
			selected[key] = v
		}
	}
	return
}

func (c Context) PruneEmpty() (pruned Context) {
	pruned = c.Copy()
	for k, v := range c {
		if v == nil || k == "" {
			delete(pruned, k)
			continue
		}
		switch t := v.(type) {
		case bool:
			if !t {
				delete(pruned, k)
			}
		case string:
			if t == "" {
				delete(pruned, k)
			}
		case []byte:
			if len(t) == 0 {
				delete(pruned, k)
			}
		case sql.NullTime:
			if !t.Valid {
				delete(pruned, k)
			}
		}
	}
	return
}

func (c Context) AsJSON() (data []byte, err error) {
	data, err = json.MarshalIndent(c, "", "  ")
	return
}

func (c Context) AsTOML() (data []byte, err error) {
	data, err = toml.Marshal(c)
	return
}

func (c Context) AsYAML() (data []byte, err error) {
	data, err = yaml.Marshal(c)
	return
}

func (c Context) FirstString(key string) (value string, ok bool) {
	var v interface{}
	var list []string
	if v, ok = c[key]; v != nil {
		if value, ok = v.(string); ok {
		} else if list, ok = v.([]string); ok && len(list) > 0 {
			value = list[0]
		}
	}
	return
}