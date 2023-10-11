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
	"errors"
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-enjin/be/pkg/maths"
)

var (
	RxTemplateParseError = regexp.MustCompile(`template: ([^:]+?):(\d+):\s*(.+?)\s*$`)
	RxTemplateExecError  = regexp.MustCompile(`template: ([^:]+?):(\d+):(\d+):\s*executing\s*"[^"]+?"\s*at\s*<[^>]+?>:\s*(.+?)\s*$`)
)

func NewEnjinOffsetError(title, err, content string, offset int64) (ee *EnjinError) {
	ee = NewEnjinOffsetRangeError(title, err, content, offset, offset)
	return
}

func NewEnjinOffsetRangeError(title, err, content string, offset, end int64) (ee *EnjinError) {
	var found bool
	var pos int64
	var lines []string
	isRange := offset < end

	for _, line := range strings.Split(content, "\n") {
		length := int64(len(line)) + 1
		if !found && pos+length >= offset {
			found = true
			delta := maths.Floor(offset-pos-1, 0)
			var escaped string

			if isRange {

				if length > 1 { // always a newline
					escaped = template.HTMLEscapeString(line[:delta])
					if delta < length {
						escaped += `<span class="enjin-error pos">`
						escaped += template.HTMLEscapeString(line[delta:])
						escaped += `</span>`
					} else {
						// last character in the line
						escaped += `<span class="enjin-error pos">&nbsp;</span>`
					}
				} else {
					// last character in the line
					escaped += `<span class="enjin-error pos">&nbsp;</span>`
				}

			} else {
				// single character point
				if length > 1 { // always a newline
					escaped = template.HTMLEscapeString(line[:delta])
					if delta < length {
						escaped += `<span class="enjin-error pos">`
						if length > 0 {
							escaped += template.HTMLEscapeString(string(line[delta]))
							escaped += `</span>`
							escaped += template.HTMLEscapeString(line[delta+1:])
						} else {
							escaped += `&nbsp;`
							escaped += `</span>`
						}
					} else {
						// last character in the line
						escaped += `<span class="enjin-error pos">&nbsp;</span>`
					}
				} else {
					// last character in the line
					escaped += `<span class="enjin-error pos">&nbsp;</span>`
				}
			}

			lines = append(lines, escaped)
			var padding string
			if delta >= 1 {
				padding = strings.Repeat("&nbsp;", int(delta)-1)
			}
			if isRange {
				lines = append(lines, padding+`<span id="json-error" class="enjin-error offset line"><i class="fa-solid fa-arrow-up"></i></span>`)
			} else {
				lines = append(lines, padding+`<span id="json-error" class="enjin-error offset"><i class="fa-solid fa-arrow-up"></i></span>`)
			}
		} else {
			lines = append(lines, template.HTMLEscapeString(line))
		}
		pos += length
	}

	ee = NewEnjinError(
		title,
		fmt.Sprintf(`<a class="enjin-error" href="#json-error">[%d] %v</a>`, offset, err),
		strings.Join(lines, "\n"),
	)
	return
}

func ParseTemplateError(message, content string) (err error) {
	if RxTemplateExecError.MatchString(message) {
		m := RxTemplateExecError.FindAllStringSubmatch(message, 1)
		text := m[0][4]
		lino, _ := strconv.ParseInt(m[0][2], 10, 64)
		colno, _ := strconv.ParseInt(m[0][3], 10, 64)
		var offset int64
		lines := strings.Split(content, "\n")
		for idx, line := range lines {
			if int64(idx) < lino-1 {
				offset += int64(len(line)) + 1
			} else if int64(idx) == lino-1 {
				offset += 1 + colno
				break
			}
		}
		err = NewEnjinOffsetError("template error", text, content, offset)
	} else if RxTemplateParseError.MatchString(message) {
		m := RxTemplateParseError.FindAllStringSubmatch(message, 1)
		text := m[0][3]
		lino, _ := strconv.ParseInt(m[0][2], 10, 64)
		var offset, end int64
		lines := strings.Split(content, "\n")
		for idx, line := range lines {
			if int64(idx) < lino-1 {
				offset += int64(len(line)) + 1
			} else if int64(idx) == lino-1 {
				offset += 1
				end = offset + int64(len(line))
				break
			}
		}
		err = NewEnjinOffsetRangeError("template error", text, content, offset, end)
	} else {
		err = errors.New(message)
	}
	return
}