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

package catalog

import (
	"strings"
	"text/scanner"
	"unicode"

	"github.com/go-enjin/be/pkg/strings/fmtsubs"
)

func parseTmplSubStatements(statement string) (list []string) {

	var s scanner.Scanner
	s.Init(strings.NewReader(statement))
	s.Error = func(_ *scanner.Scanner, _ string) {}
	s.Filename = "input.tmpl"
	s.Mode ^= scanner.SkipComments
	s.Whitespace ^= 1<<'\t' | 1<<'\n' | 1<<' '

	var isOpen bool
	var current string

	var stack []string

	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		token := s.TokenText()

		stackSize := len(stack)
		if token == "(" {
			if isOpen {
				stack = append(stack, strings.TrimSpace(current))
				current = ""
			} else {
				isOpen = true
			}
			continue
		}
		if token == ")" {
			list = append(list, strings.TrimSpace(current))
			if stackSize > 0 {
				current = stack[stackSize-1]
				stack = stack[:stackSize-1]
			} else {
				current = ""
				isOpen = false
			}
			continue
		}

		if isOpen {
			current += token
		}
	}

	if len(stack) > 0 {
		list = append(list, stack...)
	}

	return
}

func parseTmplStatements(input string) (list []string) {

	var s scanner.Scanner
	s.Init(strings.NewReader(input))
	s.Error = func(_ *scanner.Scanner, _ string) {}
	s.Filename = "input.tmpl"
	s.Mode ^= scanner.SkipComments
	s.Whitespace ^= 1<<'\t' | 1<<'\n' | 1<<' '

	var foundOpen, isOpen, foundClose bool
	var current string

	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		token := s.TokenText()

		if token == "{" {
			if foundOpen {
				// found second opening curly-brace
				isOpen = true
				continue
			}
			// found first opening curly-brace
			foundOpen = true
			continue
		} else if foundOpen {
			foundOpen = false
		}

		if token == "}" {
			if foundClose {
				// found second closing curly-brace
				list = append(list, strings.TrimSpace(strings.Trim(current, "-")))
				if extras := parseTmplSubStatements(current); len(extras) > 0 {
					list = append(list, extras...)
				}
				current = ""
				foundOpen, isOpen, foundClose = false, false, false
				continue
			}
			// found first closing curly-brace
			foundClose = true
			continue
		} else if foundClose {
			foundClose = false
		}

		if isOpen {
			current += token
		}
	}

	return
}

func ParseMessagePlaceholders(key string, argv ...string) (replaced, labelled string, placeholders Placeholders) {
	var subs fmtsubs.FmtSubs
	replaced, labelled, subs, _ = fmtsubs.ParseFmtString(key, argv...)
	for _, sub := range subs {
		placeholders = append(placeholders, &Placeholder{
			ID:             sub.Label,
			String:         sub.String(),
			Type:           sub.Type,
			UnderlyingType: sub.Type,
			ArgNum:         sub.Pos,
			Expr:           "-",
		})
	}
	return
}

func MakeMessageFromKey(key, comment string, argv ...string) (m *Message) {
	replaced, labelled, placeholders := ParseMessagePlaceholders(key, argv...)
	m = &Message{
		ID:                labelled,
		Key:               key,
		Message:           replaced,
		Translation:       &Translation{String: replaced},
		TranslatorComment: comment,
		Placeholders:      placeholders,
		Fuzzy:             true,
	}
	return
}

type parseMessageState struct {
	format  string
	argv    []string
	comment string
}

func ParseTemplateMessages(input string) (msgs []*Message, err error) {

	var pruned []string
	for _, item := range parseTmplStatements(input) {
		if strings.HasPrefix(item, "_ ") {
			item = item[2:]
			if pIdx := strings.Index(item, "|"); pIdx > -1 {
				item = strings.TrimSpace(item[:pIdx])
			}
			pruned = append(pruned, item)
		}
	}

	var list []*parseMessageState

	for _, item := range pruned {
		var s scanner.Scanner
		s.Init(strings.NewReader(item))
		s.Error = func(_ *scanner.Scanner, _ string) {}
		s.Filename = "input.tmpl"
		s.Mode ^= scanner.SkipComments
		//s.Whitespace ^= 1<<'\t' | 1<<'\n' | 1<<' '
		s.IsIdentRune = func(ch rune, i int) bool {
			if i == 0 {
				if ch == '$' || ch == '.' {
					return true
				}
				// all template identifiers start with $ or .
				return false
			}
			return ch == '.' || ch == '_' || unicode.IsLetter(ch) || (unicode.IsDigit(ch) && i > 1)
		}

		state := &parseMessageState{}

		for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
			token := s.TokenText()
			if state.format == "" {
				if size := len(token); size > 2 {
					if beStrings.IsQuoted(token) {
						state.format = beStrings.TrimQuotes(token)
					} else {
						// variable translation
						state = nil
						break
					}
				}
			} else if strings.HasPrefix(token, "/*") {
				if state.comment != "" {
					state.comment += "\n"
				}
				state.comment += token
			} else if beStrings.IsQuoted(token) {
				// support quoted string arguments
				state.argv = append(state.argv, beStrings.TrimQuotes(token))
			} else {
				if argc := len(state.argv); argc > 0 {
					switch state.argv[argc-1] {
					case "$", ".":
						state.argv[argc-1] += token
					default:
						state.argv = append(state.argv, token)
					}
				} else {
					state.argv = append(state.argv, token)
				}
			}
		}

		if state != nil {
			list = append(list, state)
		}
	}

	var order []string
	unique := make(map[string][]*parseMessageState)
	for _, item := range list {
		if _, present := unique[item.format]; !present {
			order = append(order, item.format)
		}
		unique[item.format] = append(unique[item.format], item)
	}

	for _, key := range order {
		items, _ := unique[key]
		item := items[0]
		comment := item.comment
		if count := len(items); count > 1 {
			for idx, itm := range items {
				if idx == 0 {
					continue
				}
				if itm.comment != "" {
					var dupe bool
					for jdx, other := range items {
						if dupe = idx != jdx && other.comment == itm.comment; dupe {
							break
						}
					}
					if !dupe {
						if comment != "" {
							comment += "\n"
						}
						comment += itm.comment
					}
				}
			}
		}
		msg := MakeMessageFromKey(item.format, comment, item.argv...)
		if argc := len(item.argv); argc > 0 {
			for idx, placeholder := range msg.Placeholders {
				index := placeholder.ArgNum - 1
				if argc > index {
					if msg.Placeholders[idx].Expr == "-" {
						msg.Placeholders[idx].Expr = item.argv[index]
					} else {
						msg.Placeholders[idx].Expr += ", " + item.argv[index]
					}
				}
			}
		}
		msgs = append(msgs, msg)
	}

	return
}