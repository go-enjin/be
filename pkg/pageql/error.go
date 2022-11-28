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

import (
	"fmt"
	"strconv"
	"strings"
)

type ParseError struct {
	err     error
	Query   string
	Column  int
	Message string
}

func newParseError(query string, participleError error) (e *ParseError) {
	if participleError == nil {
		return
	}
	e = &ParseError{
		err:   participleError,
		Query: query,
	}
	if parts := strings.Split(e.err.Error(), ":"); len(parts) == 4 {
		var sCol, sMsg = parts[2], strings.TrimSpace(parts[3])
		if col, err := strconv.Atoi(sCol); err == nil {
			e.Column = col
		}
		e.Message = sMsg
	}
	return
}

func (e *ParseError) Error() (msg string) {
	if e.err != nil {
		msg = e.err.Error()
	}
	return
}

func (e *ParseError) Pretty() (refined string) {
	indent := strings.Repeat(" ", e.Column-1)
	message := fmt.Sprintf("error: %v", e.Message)
	refined = fmt.Sprintf("%v\n%v^- %v\n", e.Query, indent, message)
	return
}