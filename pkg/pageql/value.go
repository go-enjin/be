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

package pageql

import "fmt"

type Value struct {
	ContextKey *string `parser:"  ( '.' @Ident )" json:"context-key,omitempty"`
	Regexp     *string `parser:"| ( 'm' @Regexp )" json:"regexp,omitempty"`
	String     *string `parser:"| ( @String )" json:"string,omitempty"`

	// Regexp *string `parser:"| ( 'm' '/' @(~'/')+ '/' ) | ( 'm' '!' @(~'!')+ '!' )" json:"regexp,omitempty"`
	// Number        *float64   `parser:"| @Float | @Int" json:"number,omitempty"`
	// Bool          *string     `parser:"| ( @'true' | 'false' )" json:"bool,omitempty"`
	// Nil           bool        `parser:"| @'nil'" json:"nil,omitempty"`
}

func (v *Value) Render() (clone *Value) {
	clone = new(Value)
	if v.ContextKey != nil {
		key := *v.ContextKey
		clone.ContextKey = &key
	}
	if v.Regexp != nil {
		pattern := *v.Regexp
		pattern, _ = UnquoteRegexp(pattern)
		clone.Regexp = &pattern
	}
	if v.String != nil {
		text := *v.String
		text, _ = UnquoteString(text)
		clone.String = &text
	}
	return
}

func UnquoteString(s string) (out string, err error) {
	if s != "" {
		last := len(s) - 1
		for _, quote := range []uint8{'\'', '"'} {
			if s[0] == quote {
				if s[last] == quote {
					out = s[1:last]
				} else {
					err = fmt.Errorf(`expected closing quote "%v"`, string(quote))
				}
				return
			}
		}
	}
	return
}

func UnquoteRegexp(s string) (out string, err error) {
	if s != "" {
		last := len(s) - 1
		for _, quote := range []uint8{'/', '!', '@', '~'} {
			if s[0] == quote {
				if s[last] == quote {
					out = s[1:last]
				} else {
					err = fmt.Errorf(`expected closing character "%v"`, string(quote))
				}
				return
			}
		}
	}
	return
}