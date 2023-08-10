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
	"strings"

	"github.com/go-enjin/go-stdlib-text-scanner"

	beStrings "github.com/go-enjin/be/pkg/strings"
)

func slurpTo(scan *scanner.Scanner, to ...string) (text string, stop string) {
	stoppers := map[string]struct{}{}
	for _, stopper := range to {
		stoppers[stopper] = struct{}{}
	}
	for tok := scan.Scan(); tok != scanner.EOF; tok = scan.Scan() {
		token := scan.TokenText()
		if _, stopped := stoppers[token]; stopped {
			stop = token
			return
		} else {
			text += token
		}
	}
	return
}

func slurpToNextTag(scan *scanner.Scanner) (text, raw, maybeTag, maybeTagRaw string, closing bool) {
	var prev, token string

	for tok := scan.Scan(); tok != scanner.EOF; tok = scan.Scan() {
		token = scan.TokenText()

		if prev == "[" {

			prefix := "["
			var keep string
			if token == "/" {
				prefix += "/"
			} else {
				keep = token
			}

			if value, stop := slurpTo(scan, "]", "["); stop == "]" {
				// found closing brace, possibly a tag
				raw += prefix + keep + value + stop
				closing = token == "/"
				maybeTag = strings.ToLower(strings.TrimSpace(keep + value))
				maybeTagRaw = keep + value
				return

			} else if stop == "[" {
				// re-opened
				raw += prefix + keep + value
				text += prefix + keep + value
				prev = "["
				token = ""
				continue

			} else {
				// did not find a closing brace, not a tag
				raw += prefix + keep + value + stop
				text += prefix + keep + value + stop
				token = ""
				prev = ""
				continue
			}

		} else {

			raw += prev
			text += prev

		}

		prev = token
	}

	raw += token
	text += token
	return
}

func parseOpeningTag(input string) (raw, name string, attributes *Attributes, ok bool) {
	attributes = newAttributes()
	raw = "[" + input + "]"

	if ok = rxNameOnly.MatchString(input); ok {
		// [name]
		name = strings.ToLower(input)

	} else if ok = rxNameValue.MatchString(input); ok {
		// [name=value]
		m := rxNameValue.FindAllStringSubmatch(input, 1)
		name = strings.ToLower(m[0][1])
		attributes.Set(name, beStrings.TrimQuotes(m[0][2]))

	} else if ok = rxNameKeyValues.MatchString(input); ok {
		// [name key=value...]
		m := rxNameKeyValues.FindAllStringSubmatch(input, 1)
		name = strings.ToLower(m[0][1])
		if argv, keys := parseArgumentString(m[0][2]); len(keys) > 0 {
			for _, key := range keys {
				attributes.Set(strings.ToLower(key), argv[key])
			}
		}

	}
	return
}

func parseClosingTag(input string) (raw, name string, ok bool) {
	raw = "[/" + input + "]"

	if ok = rxNameOnly.MatchString(input); ok {
		// [/name]
		name = strings.ToLower(input)
	}

	return
}

func parseArgumentString(input string) (argv map[string]string, keys []string) {
	argv = make(map[string]string)

	findingKey := true
	var openQuote uint8
	var key, value string
	for i := 0; i < len(input); i++ {
		if findingKey {
			if input[i] == ' ' {
				continue
			} else if input[i] == '=' {
				findingKey = false
			} else {
				key += string(input[i])
			}
		} else {
			if value == "" {
				if input[i] == '"' || input[i] == '\'' {
					openQuote = input[i]
				} else {
					value += string(input[i])
				}
			} else {
				if openQuote > 0 {
					if input[i] == openQuote {
						openQuote = 0
						findingKey = true
						argv[key] = value
						keys = append(keys, key)
						key = ""
						value = ""
					} else {
						value += string(input[i])
					}
				} else {
					if input[i] == ' ' {
						findingKey = true
						argv[key] = value
						keys = append(keys, key)
						key = ""
						value = ""
					} else {
						value += string(input[i])
					}
				}
			}
		}
	}

	if key != "" {
		argv[key] = value
		keys = append(keys, key)
	}

	return
}