//go:build page_shortcodes || pages || all

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

package shortcodes

import (
	"fmt"
	"strings"
	"unicode"

	clStrings "github.com/go-corelibs/strings"
)

// IsKeyword ranges over the input string and returns true if the first
// rune is a letter and the remainder are all digits, letters, or underscores
//
// IsKeyword is equivalent to matching the regexp pattern:
// `[a-zA-Z][_a-zA-Z0-9]*[a-zA-Z0-9]*`
func IsKeyword(input string) (ok bool) {
	for idx, r := range input {
		if idx == 0 {
			if ok = unicode.IsLetter(r); !ok {
				return
			}
		} else if ok = r == '_' || unicode.IsDigit(r) || unicode.IsLetter(r); !ok {
			return
		}
	}
	return
}

// IsToken ranges over the input string and returns true if the first
// rune is a letter and the remainder are all digits, letters, hyphens or
// underscores
//
// IsToken is equivalent to matching the regexp pattern:
// `[a-zA-Z][-_a-zA-Z0-9]*[a-zA-Z0-9]*`
func IsToken(input string) (ok bool) {
	for idx, r := range input {
		if idx == 0 {
			if ok = unicode.IsLetter(r); !ok {
				return
			}
		} else if ok = r == '-' || r == '_' || unicode.IsDigit(r) || unicode.IsLetter(r); !ok {
			return
		}
	}
	return
}

// ParseKeyValue parses the input for the equivalent regexp pattern:
// `^([a-zA-Z][-_a-zA-Z0-9]*[a-zA-Z0-9]*)\s*=\s*["'](.+?)["']\s*$`
func ParseKeyValue(input string) (key, value string, ok bool) {
	if before, after, found := strings.Cut(input, "="); found {
		before = strings.TrimSpace(before)
		if ok = IsToken(before); ok {
			key = before
			after = strings.TrimSpace(after)
			if ok = after != ""; ok {
				value = clStrings.TrimQuotes(after)
			}
		}
	}
	return
}

func searchForSpaceUntilEquals(input []rune) (ok bool) {
	for _, next := range input {
		if unicode.IsSpace(next) {
		} else if ok = next == '='; ok {
			return
		}
		return
	}
	return
}

func SplitTokenAttributes(input string) (pairs []string) {
	s := &struct {
		open   bool
		closer rune
		key    string
		value  string
		skip   bool
	}{}
	actual := []rune(input)
	last := len(actual) - 1
	for idx, r := range actual {
		if s.skip {
			s.skip = false
			continue
		}
		if s.open { // looking for value
			if (s.closer > 0 && s.closer != r) || (s.closer == 0 && !unicode.IsSpace(r)) {
				s.value += string(r)
				continue
			}
			pairs = append(pairs, fmt.Sprintf(`%s=%q`, s.key, s.value))
			s.open = false
			s.closer = 0
			s.key = ""
			s.value = ""
			continue
		}
		if unicode.IsSpace(r) {
			if !searchForSpaceUntilEquals(actual[idx:]) && s.key != "" {
				pairs = append(pairs, s.key)
				s.key = ""
			}
			continue
		}
		if s.open = r == '='; s.open {
			if idx < last {
				if s.skip = clStrings.IsQuote(actual[idx+1]); s.skip {
					s.closer = actual[idx+1]
				} else if p, ok := clStrings.GetFancyQuote(actual[idx+1]); ok {
					s.skip = true
					s.closer = p.End
				} else {
					s.closer = 0
				}
			}
			continue
		}
		// looking for key
		if IsToken(s.key + string(r)) {
			s.key += string(r)
			continue
		}
	}

	if s.key != "" && s.value != "" {
		pairs = append(pairs, fmt.Sprintf(`%s=%q`, s.key, s.value))
	}
	return
}

func ParseTokenWithAttributes(input string) (token string, attributes *Attributes, ok bool) {
	parts := SplitTokenAttributes(input)
	if ok = len(parts) > 0; ok {
		if ok = IsToken(parts[0]); ok {
			token = strings.ToLower(parts[0])
			attributes = newAttributes()
			for _, pair := range parts[1:] {
				var k, v string
				if k, v, ok = ParseKeyValue(pair); ok {
					attributes.Set(k, v)
				} else if ok = IsToken(pair); ok {
					attributes.Set(pair, "true")
				}
			}
		}
	}
	return
}
