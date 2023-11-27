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

package request

import (
	"context"
	"net/http"
)

// Set is a convenience wrapper for cloning the given request with one or more context key+value pairs added, will panic
// if the number of key+value pairs is not balanced (even)
func Set(r *http.Request, keyValuePairs ...interface{}) (m *http.Request) {
	if r == nil {
		return
	}
	count := len(keyValuePairs)
	if count%2 != 0 {
		panic("unbalanced keyValuePairs argument")
	}
	m = r
	for i := 0; i < count; i += 2 {
		key, value := keyValuePairs[i], keyValuePairs[i+1]
		m = m.Clone(context.WithValue(m.Context(), key, value))
	}
	return
}

// Value is a convenience wrapper for getting a generic value type from a given request context
func Value[T interface{}](r *http.Request, key interface{}) (value T, ok bool) {
	if r == nil {
		return
	}
	value, ok = r.Context().Value(key).(T)
	return
}

// String is a convenience wrapper around Value[string]()
func String(r *http.Request, key interface{}) (value string, ok bool) {
	value, ok = Value[string](r, key)
	return
}

// Int is a convenience wrapper around Value[int]()
func Int(r *http.Request, key interface{}) (value int, ok bool) {
	value, ok = Value[int](r, key)
	return
}

// Float64 is a convenience wrapper around Value[float64]()
func Float64(r *http.Request, key interface{}) (value float64, ok bool) {
	value, ok = Value[float64](r, key)
	return
}
