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
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"

	"github.com/fvbommel/sortorder"
	"github.com/iancoleman/strcase"
	"github.com/maruel/natural"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"

	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
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

var rxOsEnvironKV = regexp.MustCompile(`^([^=]+?)=(.+?)$`)

// NewFromOsEnviron constructs a new Context from os.Environ() string K=V slices
func NewFromOsEnviron(slices ...[]string) (c Context) {
	c = New()
	for _, slice := range slices {
		for _, pair := range slice {
			if rxOsEnvironKV.MatchString(pair) {
				m := rxOsEnvironKV.FindAllStringSubmatch(pair, 1)
				key, value := m[0][1], m[0][2]
				c.SetSpecific(key, value)
			}
		}
	}
	return
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

// TODO: make context.Copy a deeper Copy() of maybe literal types only

// Copy makes a duplicate of this Context
func (c Context) Copy() (ctx Context) {
	ctx = New()
	for k, v := range c {
		switch t := v.(type) {
		case map[string]bool:
			m := make(map[string]bool)
			for tk, tv := range t {
				m[tk] = tv
			}
			ctx.SetSpecific(k, m)
		case map[string]int:
			m := make(map[string]int)
			for tk, tv := range t {
				m[tk] = tv
			}
			ctx.SetSpecific(k, m)
		case map[string]string:
			m := make(map[string]string)
			for tk, tv := range t {
				m[tk] = tv
			}
			ctx.SetSpecific(k, m)
		case map[string]interface{}:
			dst := make(map[string]interface{})
			if encoded, err := json.Marshal(t); err != nil {
				log.ErrorF("error marshalling map[string]interface{}: %v", err)
			} else if err = json.Unmarshal(encoded, &dst); err != nil {
				log.ErrorF("error unmarshalling map[string]interface{}: %v", err)
			} else {
				ctx.SetSpecific(k, Context(dst))
			}
		case []bool:
			ctx.SetSpecific(k, t[:])
		case []string:
			ctx.SetSpecific(k, t[:])
		case []int:
			ctx.SetSpecific(k, t[:])
		case []int8:
			ctx.SetSpecific(k, t[:])
		case []int16:
			ctx.SetSpecific(k, t[:])
		case []int32:
			ctx.SetSpecific(k, t[:])
		case []int64:
			ctx.SetSpecific(k, t[:])
		case []uint:
			ctx.SetSpecific(k, t[:])
		case []uint8:
			ctx.SetSpecific(k, t[:])
		case []uint16:
			ctx.SetSpecific(k, t[:])
		case []uint32:
			ctx.SetSpecific(k, t[:])
		case []uint64:
			ctx.SetSpecific(k, t[:])
		case []float32:
			ctx.SetSpecific(k, t[:])
		case []float64:
			ctx.SetSpecific(k, t[:])
		case bool, string,
			int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64:
			ctx.SetSpecific(k, t)
		case float32, float64:
			ctx.SetSpecific(k, t)
		default:
			ctx.SetSpecific(k, t)
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

// SetKV allows for .Deep.Variable.Names
func (c Context) SetKV(key string, value interface{}) (err error) {
	if key != "" && key[0] == '.' {
		err = maps.Set(key, value, c)
		return
	}
	k := strcase.ToCamel(key)
	c[k] = value
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
	if key != "" && key[0] == '.' {
		k = key
		v = maps.Get(key, c)
		return
	}

	k = key
	var ok bool
	if v, ok = c[k]; ok {
		return
	}
	k = strcase.ToCamel(key)
	if v, ok = c[k]; ok {
		return
	}
	k = strcase.ToKebab(key)
	if v, ok = c[k]; ok {
		return
	}
	k = strcase.ToSnake(key)
	if v, ok = c[k]; ok {
		return
	}
	return
}

// Delete deletes the given key from the Context and follows a similar key
// lookup process to Get() for finding the key to delete and will only delete
// the first matching key format (specific, Camel, kebab) found
func (c Context) Delete(key string) (deleted bool) {
	if key != "" && key[0] == '.' {
		maps.Delete(key, c)
		return
	}
	if k, v := c.GetKV(key); v != nil {
		delete(c, k)
		return true
	}
	return false
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

func (c Context) Uint(key string, def uint) uint {
	if v := c.Get(key); v != nil {
		if i, ok := v.(uint); ok {
			return i
		}
	}
	return def
}

func (c Context) Uint64(key string, def uint64) uint64 {
	if v := c.Get(key); v != nil {
		if i, ok := v.(uint64); ok {
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
		if camelized := strcase.ToCamel(k); k != camelized {
			remove = append(remove, k)
			c.SetSpecific(camelized, v)
		}
	}
	c.DeleteKeys(remove...)
}

// LowerCamelizeKeys transforms all keys within the Context to be of lowerCamelCased form
func (c Context) LowerCamelizeKeys() {
	var remove []string
	for k, v := range c {
		if camelized := strcase.ToLowerCamel(k); k != camelized {
			remove = append(remove, k)
			c.SetSpecific(camelized, v)
		}
	}
	c.DeleteKeys(remove...)
}

// KebabKeys transforms all keys within the Context to be of kebab-cased form
func (c Context) KebabKeys() {
	var remove []string
	for k, v := range c {
		if camelized := strcase.ToKebab(k); k != camelized {
			remove = append(remove, k)
			c.SetSpecific(camelized, v)
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

func (c Context) AsJSON() (data []byte, err error) {
	data, err = json.Marshal(c)
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