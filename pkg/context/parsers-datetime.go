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
	"sync"
	"time"

	"github.com/go-corelibs/slices"
	"github.com/go-enjin/be/pkg/values"
)

var (
	DateLayout     = "Jan 02, 2006"
	TimeLayout     = "15:04 -0700"
	DateTimeLayout = "2006-01-02 15:04 MST"

	gDateTime = struct {
		formats []string
		sync.RWMutex
	}{
		formats: []string{
			DateLayout,
			TimeLayout,
			DateTimeLayout,
			time.DateOnly,
			time.DateTime,
			time.TimeOnly,
			time.RFC1123,
			time.RFC1123Z,
			time.RFC3339,
			"2006-01-02T15:04",
		},
	}
)

func AddDateTimeFormat(format string) {
	gDateTime.Lock()
	defer gDateTime.RUnlock()
	if !slices.Within(format, gDateTime.formats) {
		gDateTime.formats = append(gDateTime.formats, format)
	}
}

func ParseTimeStructure(input string) (parsed time.Time, err error) {
	input = strings.TrimSpace(input)
	gDateTime.RLock()
	defer gDateTime.RUnlock()
	for _, format := range gDateTime.formats {
		if v, e := time.Parse(format, input); e == nil {
			parsed = v
			return
		}
	}
	err = errors.New("not a time format value")
	return
}

func TimeStructParser(spec *Field, input interface{}) (parsed interface{}, err error) {
	if list, ok := input.([]string); ok && len(list) > 0 {
		input = list[0]
	}
	switch t := input.(type) {
	case string:
		if t = strings.TrimSpace(t); t != "" {
			parsed, err = ParseTimeStructure(t)
		}
	default:
		err = errors.New(spec.Printer.Sprintf("unsupported type: %[1]s", values.TypeOf(input)))
	}
	return
}

func TimeParser(spec *Field, input interface{}) (parsed interface{}, err error) {
	if list, ok := input.([]string); ok && len(list) > 0 {
		input = list[0]
	}
	switch t := input.(type) {
	case string:
		if t = strings.TrimSpace(t); t != "" {
			gDateTime.RLock()
			defer gDateTime.RUnlock()
			for _, format := range gDateTime.formats {
				if v, e := time.Parse(format, t); e == nil {
					parsed = v.Format(TimeLayout)
					return
				}
			}
			err = errors.New(spec.Printer.Sprintf("not a time value"))
		}
	default:
		err = errors.New(spec.Printer.Sprintf("unsupported type: %[1]s", values.TypeOf(input)))
	}
	return
}

func DateParser(spec *Field, input interface{}) (parsed interface{}, err error) {
	if list, ok := input.([]string); ok && len(list) > 0 {
		input = list[0]
	}
	switch t := input.(type) {
	case string:
		if t = strings.TrimSpace(t); t != "" {
			gDateTime.RLock()
			defer gDateTime.RUnlock()
			for _, format := range gDateTime.formats {
				if v, e := time.Parse(format, t); e == nil {
					parsed = v.Format(DateLayout)
					return
				}
			}
			err = errors.New(spec.Printer.Sprintf("not a date value"))
		}
	default:
		err = errors.New(spec.Printer.Sprintf("unsupported type: %[1]s", values.TypeOf(input)))
	}

	return
}

func DateTimeParser(spec *Field, input interface{}) (parsed interface{}, err error) {
	if list, ok := input.([]string); ok && len(list) > 0 {
		input = list[0]
	}
	switch t := input.(type) {
	case string:
		if t = strings.TrimSpace(t); t != "" {
			gDateTime.RLock()
			defer gDateTime.RUnlock()
			for _, format := range gDateTime.formats {
				if v, e := time.Parse(format, t); e == nil {
					parsed = v.Format(DateTimeLayout)
					return
				}
			}
			err = errors.New(spec.Printer.Sprintf("not a datetime value"))
		}
	default:
		err = errors.New(spec.Printer.Sprintf("unsupported type: %[1]s", values.TypeOf(input)))
	}
	return
}
