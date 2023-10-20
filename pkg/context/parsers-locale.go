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

package context

import (
	"errors"
	"strings"

	"github.com/go-enjin/be/pkg/values"
	"github.com/go-enjin/golang-org-x-text/language"
)

func LanguageParser(spec *Field, input interface{}) (parsed interface{}, err error) {
	if list, ok := input.([]string); ok && len(list) > 0 {
		input = list[0]
	}
	switch t := input.(type) {
	case string:
		if t = strings.TrimSpace(t); t != "" {
			var tag language.Tag
			if tag, err = language.Parse(t); err == nil {
				parsed = tag.String()
			}
		}
	default:
		err = errors.New(spec.Printer.Sprintf("unsupported type: %[1]s", values.TypeOf(input)))
	}
	return
}

func LanguageTagParser(spec *Field, input interface{}) (parsed interface{}, err error) {
	if list, ok := input.([]string); ok && len(list) > 0 {
		input = list[0]
	}
	switch t := input.(type) {
	case string:
		if t = strings.TrimSpace(t); t != "" {
			var tag language.Tag
			if tag, err = language.Parse(t); err == nil {
				parsed = tag
			}
		}
	default:
		err = errors.New(spec.Printer.Sprintf("unsupported type: %[1]s", values.TypeOf(input)))
	}
	return
}