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

package fmtsubs

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"

	"github.com/go-enjin/be/pkg/convert"
)

type subst struct {
	pos  int
	verb string

	width     int
	precision int

	plus  bool
	minus bool
	hash  bool
	space bool

	zero bool

	source string
	value  string
}

func ParseFmtString(format string, argv ...string) (replaced, labelled string, subs FmtSubs, err error) {

	var list FmtSubs
	var state *subst
	var posOpen bool
	currentPos := 1

	built := map[int][]*subst{}

	last := len(format) - 1
	for i := 0; i <= last; i++ {
		char := format[i]

		if state == nil {
			if char == '%' {
				state = &subst{
					source: "%",
					pos:    currentPos,
				}
			}
			continue
		} else if char == '%' {
			state = nil
			continue
		}

		state.source += string(char)

		switch char {
		case 'b', 'c', 'd', 'e', 'E', 'f', 'F', 'g', 'G', 'o', 'O', 'p', 'q', 's', 't', 'T', 'U', 'v', 'x', 'X':
			state.verb = string(char)
			state.value += state.verb

			subPos := -1
			if found, ok := built[state.pos]; ok {
				for idx, os := range found {
					if os.value == state.value {
						subPos = idx
						break
					}
				}
				if subPos == -1 {
					subPos = len(found)
				}
			}

			var label, valueType string

			switch state.verb {
			case "s":
				valueType = "string"
			case "d":
				valueType = "int"
			case "f", "F":
				valueType = "float"
			default:
				valueType = "-"
			}

			if len(argv) > (state.pos - 1) {
				label = strcase.ToCamel(argv[state.pos-1])
				if label != "" && subPos > 0 {
					label += fmt.Sprintf("%d%s", state.pos, convert.ToLetters(subPos))
				}
			}

			if label == "" {

				switch state.verb {
				case "s":
					label = "Txt"
				case "d":
					label = "Num"
				case "f", "F":
					label = "Num"
				default:
					label = "Arg"
				}

				if subPos <= 0 {
					label += fmt.Sprintf("%d", state.pos)
				} else {
					label += fmt.Sprintf("%d%s", state.pos, convert.ToLetters(subPos))
				}
			}

			list = append(list, &FmtSub{
				Type:   valueType,
				Label:  label,
				Source: state.source,
				Pos:    state.pos,
				Verb:   state.verb,
				Value:  state.value,
			})

			built[state.pos] = append(built[state.pos], state)
			state = nil
			posOpen = false
			currentPos += 1

		case '+':
			state.value += string(char)
			state.plus = true
		case '-':
			state.value += string(char)
			state.minus = true
		case '#':
			state.value += string(char)
			state.hash = true
		case ' ':
			state.value += string(char)
			state.space = true

		case '*':
			state.pos = currentPos

		case '[':
			posOpen = true

		case ']':
			posOpen = false

		default:
			if posOpen {
				number := -1
				if v, e := strconv.Atoi(string(char)); e == nil {
					number = v
				}
				if number > -1 {
					state.pos = number
					currentPos = number
				}
				continue
			}

			switch char {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.':
				state.value += string(char)
			default:
				// not actually a substitution
				state = nil
			}

		}

	}

	replaced = format
	labelled = format

	unique := map[int]*FmtSub{}
	for _, sub := range list {
		if orig, present := unique[sub.Pos]; present {
			if orig.Type != sub.Type {
				err = fmt.Errorf(`conflicting substitution types: %v != %v`, orig, sub)
			}
		} else {
			unique[sub.Pos] = sub
			subs = append(subs, sub)
		}
		replaced = strings.Replace(replaced, sub.Source, sub.String(), 1)
		labelled = strings.Replace(labelled, sub.Source, "{"+sub.Label+"}", 1)
	}

	subs = subs.Sort()
	return
}