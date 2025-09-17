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

package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-corelibs/maths"
)

var (
	rxTemplateSummaryError = regexp.MustCompile(`template error: \[(\d+):(\d+):(\d+)\]\s*(.+?)\s*$`)
	rxTemplateParseError   = regexp.MustCompile(`template: ([^:]+?):(\d+):\s*(.+?)\s*$`)
	rxTemplateExecError    = regexp.MustCompile(`template: ([^:]+?):(\d+):(\d+):\s*executing\s*"[^"]+?"\s*at\s*<[^>]+?>:\s*(.+?)\s*$`)
)

func makeEnjinErrorSpan(unescaped string) (span string) {
	return fmt.Sprintf(
		`<span id="error" class="enjin-error pos">%s</span>`,
		template.HTMLEscapeString(unescaped),
	)
}

func NewEnjinOffsetError(title, err, content string, offset int64) (ee *EnjinError) {
	ee = NewEnjinOffsetRangeError(title, err, content, offset, offset)
	return
}

func makePointError(content string, offset int64) (lines []string, row, column int) {

	var pos int64 = 0
	for idx, line := range strings.Split(content, "\n") {
		var escaped string
		eol := int64(len(line))
		posEol := pos + eol

		switch {

		case offset == pos:
			// start of line or empty line
			if eol == 1 {
				// empty line
				escaped += makeEnjinErrorSpan(" ")
			} else {
				if ll := len(line); ll > 0 {
					escaped += makeEnjinErrorSpan(string(line[0]))
					if ll > 1 {
						escaped += template.HTMLEscapeString(line[1:])
					}
				}
			}

		case offset >= pos && offset < posEol:
			// middle of line
			delta := posEol - offset // -1 for the implied newline
			row = idx + 1
			column = int(delta)
			escaped += template.HTMLEscapeString(line[:delta])
			escaped += makeEnjinErrorSpan(string(line[delta]))
			escaped += template.HTMLEscapeString(line[delta+1:])

		case offset == posEol:
			// end of line
			escaped += template.HTMLEscapeString(line)
			escaped += makeEnjinErrorSpan(" ")

		default:
			// not the error line
			escaped = template.HTMLEscapeString(line)
		}

		lines = append(lines, escaped)
		pos += eol // line plus \n
	}

	return
}

func makeRangeError(content string, offset, end int64) (lines []string, row, column int) {
	lines, row, column = makePointError(content, offset)
	// TODO: implement makeRangeError for template error cases, makePointError is for json errors

	// all non-error lines must be html escaped
	//lines = append(lines, template.HTMLEscapeString(line))

	//if length > 1 { // always a newline
	//	escaped = template.HTMLEscapeString(line[:delta])
	//	if delta < length {
	//		escaped += `<span class="enjin-error pos">`
	//		escaped += template.HTMLEscapeString(line[delta:])
	//		escaped += `</span>`
	//	} else {
	//		// last character in the line
	//		escaped += `<span class="enjin-error pos">&nbsp;</span>`
	//	}
	//} else {
	//	// last character in the line
	//	escaped += `<span class="enjin-error pos">&nbsp;</span>`
	//}

	return
}

func NewEnjinOffsetRangeError(title, err, content string, offset, end int64) (ee *EnjinError) {
	var lines []string
	var row, column int

	if offset < end {
		// ranged error
		lines, row, column = makeRangeError(content, offset, end)
	} else {
		// position error
		lines, row, column = makePointError(content, offset)
	}

	count := len(lines)
	digits := strconv.Itoa(maths.IntegerLen(count))
	format := `%` + digits + `d %s`
	for i := 0; i < count; i++ {
		lines[i] = fmt.Sprintf(format, i+1, lines[i])
	}

	ee = NewEnjinError(
		title,
		fmt.Sprintf(`<a class="enjin-error" href="#error">[%d:%d:%d] %v</a>`, offset, row, column, err),
		strings.Join(lines, "\n"),
	)
	return
}

func ParseTemplateError(message, content string) (err error) {
	if m := rxTemplateSummaryError.FindAllStringSubmatch(message, 1); len(m) > 0 {
		text := m[0][4]
		lino, _ := strconv.ParseInt(m[0][2], 10, 64)
		colno, _ := strconv.ParseInt(m[0][3], 10, 64)
		var offset int64
		lines := strings.Split(content, "\n")
		for idx, line := range lines {
			if int64(idx) < lino-1 {
				offset += int64(len(line))
			} else if int64(idx) == lino-1 {
				offset += 1 + colno
				break
			}
		}
		err = NewEnjinOffsetError("template error", text, content, offset)
	} else if m := rxTemplateExecError.FindAllStringSubmatch(message, 1); len(m) > 0 {
		text := m[0][4]
		lino, _ := strconv.ParseInt(m[0][2], 10, 64)
		colno, _ := strconv.ParseInt(m[0][3], 10, 64)
		var offset int64
		lines := strings.Split(content, "\n")
		for idx, line := range lines {
			if int64(idx) < lino-1 {
				offset += int64(len(line))
			} else if int64(idx) == lino-1 {
				offset += 1 + colno
				break
			}
		}
		err = NewEnjinOffsetError("template error", text, content, offset)
	} else if m := rxTemplateParseError.FindAllStringSubmatch(message, 1); len(m) > 0 {
		text := m[0][3]
		lino, _ := strconv.ParseInt(m[0][2], 10, 64)
		var offset, end int64
		lines := strings.Split(content, "\n")
		for idx, line := range lines {
			if int64(idx) < lino-1 {
				offset += int64(len(line))
			} else if int64(idx) == lino-1 {
				offset += 1
				end = offset + int64(len(line))
				break
			}
		}
		err = NewEnjinOffsetRangeError("template error", text, content, offset, end)
	} else {
		err = errors.New(ExtractErrSummary(message))
	}
	return
}

func ParseJsonError(e error, content string) (err error) {
	var jse *json.SyntaxError
	var jute *json.UnmarshalTypeError

	if errors.As(e, &jse) {
		err = NewEnjinOffsetError("json syntax error", jse.Error(), content, jse.Offset)
	} else if errors.As(e, &jute) {
		err = NewEnjinOffsetError("json decode error", jute.Error(), content, jute.Offset)
	} else {
		err = NewEnjinOffsetError("json error", e.Error(), content, -1)
	}

	return
}
