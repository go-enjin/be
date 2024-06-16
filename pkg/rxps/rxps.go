// Copyright (c) 2024  The Go-Enjin Authors
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

package rxps

import (
	"github.com/go-corelibs/rxp"
)

var (
	RxFieldKey = rxp.Pattern{
		rxp.IsFieldKey("c"),
	}
	RxKeywords = rxp.Pattern{
		rxp.IsKeyword("c"),
	}

	RxLanguageKey = rxp.Pattern{
		rxp.Text("language:"),
		IsLanguageKey("+", "c"),
		rxp.S("*"),
	}
)

// IsLanguageKey creates a Matcher equivalent to:
//
//	(?:\*|[a-z][-a-zA-Z]+)
func IsLanguageKey(flags ...string) rxp.Matcher {
	_, cfg := rxp.ParseFlags(flags...)
	return func(scope rxp.Flags, reps rxp.Reps, input *rxp.InputReader, index int, sm [][2]int) (scoped rxp.Flags, consumed int, proceed bool) {
		scoped = scope | cfg

		if inputLen := input.Len(); 0 <= index && index < inputLen {

			if this, size, ok := input.Get(index); ok {
				if this == '*' {
					if proceed = inputLen == index+size; proceed {
						// matched wildcard
						consumed = size
						proceed = !scoped.Negated()
						return
					}
					// has an asterisk but also has more
					proceed = scoped.Negated()
					return
				}

				if proceed = 'a' <= this && this <= 'z'; proceed {
					consumed += size

					for idx := index + size; idx < inputLen; {
						if next, sz, present := input.Get(idx); present {
							if next == '-' || ('a' <= next && next <= 'z') || ('A' <= next && next <= 'Z') {
								idx += sz
								consumed += sz
								continue
							}
						}

						break
					}

					// matched at least one character
					proceed = !scoped.Negated()
					return
				}
			}

		}

		// did not match
		consumed = 0
		proceed = scoped.Negated()
		return
	}
}
