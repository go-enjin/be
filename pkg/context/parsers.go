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
	"net/mail"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-curses/cdk/lib/math"
	"github.com/gofrs/uuid"
	"github.com/iancoleman/strcase"

	"github.com/go-enjin/be/pkg/forms"
	bePath "github.com/go-enjin/be/pkg/path"
	"github.com/go-enjin/be/pkg/slices"
	beStrings "github.com/go-enjin/be/pkg/strings"
	"github.com/go-enjin/be/pkg/values"
)

var (
	DefaultParsers = Parsers{
		"bool":            BoolParser,
		"html":            HtmlParser,
		"string":          StringParser,
		"string-map":      StringMapParser,
		"string-option":   StringOptionParser,
		"number":          NumberParser,
		"number-percent":  NumberPercentParser,
		"decimal":         DecimalParser,
		"decimal-percent": DecimalPercentParser,
		"email":           EmailParser,
		"path":            PathParser,
		"url":             UrlParser,
		"relative-url":    RelativeUrlParser,
		"time":            TimeParser,
		"date":            DateParser,
		"date-time":       DateTimeParser,
		"time-struct":     TimeStructParser,
		"language":        LanguageParser,
		"language-tag":    LanguageTagParser,
		"kebab":           KebabParser,
		"kebab-option":    KebabOptionParser,
		"uuid":            UuidParser,
	}
)

type Parsers map[string]Parser

type Parser func(spec *Field, input interface{}) (parsed interface{}, err error)

func ToKebab(input string) (kebab string) {
	input = strings.TrimSpace(input)
	v := strcase.ToKebab(input)
	parts := strings.Split(v, " ")
	kebab = strings.Join(parts, "--")
	return
}

func BoolParser(spec *Field, input interface{}) (parsed interface{}, err error) {
	if list, ok := input.([]string); ok && len(list) > 0 {
		input = list[0]
	}
	switch t := input.(type) {
	case string:
		if t = strings.TrimSpace(t); t != "" {
			if isTrue := beStrings.IsTrue(t); isTrue {
				parsed = true
			} else if isFalse := beStrings.IsFalse(t); isFalse {
				parsed = false
			} else {
				err = errors.New(spec.Printer.Sprintf("not a boolean string value"))
			}
		}
	case bool:
		parsed = t
	case float64:
		parsed = t > 0.0
	case int:
		parsed = t > 0
	default:
		err = errors.New(spec.Printer.Sprintf("unsupported type: %[1]s", values.TypeOf(input)))
	}
	return
}

func HtmlParser(spec *Field, input interface{}) (parsed interface{}, err error) {
	if list, ok := input.([]string); ok && len(list) > 0 {
		input = list[0]
	}
	switch t := input.(type) {
	case string:
		if t = strings.TrimSpace(t); t != "" {
			parsed = forms.Sanitize(t)
		}
	default:
		err = errors.New(spec.Printer.Sprintf("unsupported type: %[1]s", values.TypeOf(input)))
	}
	return
}

func StringParser(spec *Field, input interface{}) (parsed interface{}, err error) {
	if list, ok := input.([]string); ok && len(list) > 0 {
		input = list[0]
	}
	switch t := input.(type) {
	case string:
		if t = strings.TrimSpace(t); t != "" {
			parsed = forms.StrictSanitize(t)
		}
	default:
		err = errors.New(spec.Printer.Sprintf("unsupported type: %[1]s", values.TypeOf(input)))
	}
	return
}

func StringOptionParser(spec *Field, input interface{}) (parsed interface{}, err error) {
	if list, ok := input.([]string); ok && len(list) > 0 {
		input = list[0]
	}
	switch t := input.(type) {
	case string:
		if t = strings.TrimSpace(t); t != "" {
			if t = forms.StrictSanitize(t); t != "" {
				if len(spec.ValueOptions) > 0 {
					if slices.Within(t, spec.ValueOptions) {
						parsed = t
						return
					}
				}
			}
		}
		if spec.DefaultValue != nil {
			if dv, ok := spec.DefaultValue.(string); ok {
				parsed = dv
				return
			}
		}
	default:
		err = errors.New(spec.Printer.Sprintf("unsupported type: %[1]s", values.TypeOf(input)))
	}
	return
}

func NumberParser(spec *Field, input interface{}) (parsed interface{}, err error) {
	if list, ok := input.([]string); ok && len(list) > 0 {
		input = list[0]
	}
	switch t := input.(type) {
	case string:
		if t = strings.TrimSpace(t); t != "" {
			parsed, err = strconv.Atoi(t)
		}
	case int:
		parsed = t
	case float64:
		parsed = int(t)
	default:
		err = errors.New(spec.Printer.Sprintf("unsupported type: %[1]s", values.TypeOf(input)))
	}
	return
}

func NumberPercentParser(spec *Field, input interface{}) (parsed interface{}, err error) {
	if list, ok := input.([]string); ok && len(list) > 0 {
		input = list[0]
	}
	switch t := input.(type) {
	case string:
		if t = strings.TrimSpace(t); t != "" {
			var v int
			if v, err = strconv.Atoi(t); err == nil {
				parsed = math.ClampI(v, 0, 100)
			}
		}
	case int:
		parsed = math.ClampI(t, 0, 100)
	case float64:
		parsed = math.ClampI(int(t), 0, 100)
	default:
		err = errors.New(spec.Printer.Sprintf("unsupported type: %[1]s", values.TypeOf(input)))
	}
	return
}

func DecimalParser(spec *Field, input interface{}) (parsed interface{}, err error) {
	if list, ok := input.([]string); ok && len(list) > 0 {
		input = list[0]
	}
	switch t := input.(type) {
	case string:
		if t = strings.TrimSpace(t); t != "" {
			parsed, err = strconv.ParseFloat(t, 64)
		}
	case int:
		parsed = float64(t)
	case float64:
		parsed = t
	default:
		err = errors.New(spec.Printer.Sprintf("unsupported type: %[1]s", values.TypeOf(input)))
	}
	return
}

func DecimalPercentParser(spec *Field, input interface{}) (parsed interface{}, err error) {
	if list, ok := input.([]string); ok && len(list) > 0 {
		input = list[0]
	}
	switch t := input.(type) {
	case string:
		if t = strings.TrimSpace(t); t != "" {
			var v float64
			if v, err = strconv.ParseFloat(t, 64); err == nil {
				parsed = math.ClampF(v, 0.0, 1.0)
			}
		}
	case int:
		parsed = math.ClampF(float64(t), 0.0, 1.0)
	case float64:
		parsed = math.ClampF(t, 0.0, 1.0)
	default:
		err = errors.New(spec.Printer.Sprintf("unsupported type: %[1]s", values.TypeOf(input)))
	}

	return
}

func EmailParser(spec *Field, input interface{}) (parsed interface{}, err error) {
	if list, ok := input.([]string); ok && len(list) > 0 {
		input = list[0]
	}
	switch t := input.(type) {
	case string:
		if t = strings.TrimSpace(t); t != "" {
			var v *mail.Address
			if v, err = mail.ParseAddress(t); err == nil {
				parsed = v.Address
			}
		}
	default:
		err = errors.New(spec.Printer.Sprintf("unsupported type: %[1]s", values.TypeOf(input)))
	}
	return
}

func PathParser(spec *Field, input interface{}) (parsed interface{}, err error) {
	if list, ok := input.([]string); ok && len(list) > 0 {
		input = list[0]
	}
	switch t := input.(type) {
	case string:
		if t = strings.TrimSpace(t); t != "" {
			parsed = filepath.Clean(t)
		} else {
			parsed = ""
		}
	default:
		err = errors.New(spec.Printer.Sprintf("unsupported type: %[1]s", values.TypeOf(input)))
	}
	return
}

func UrlParser(spec *Field, input interface{}) (parsed interface{}, err error) {
	if list, ok := input.([]string); ok && len(list) > 0 {
		input = list[0]
	}
	switch t := input.(type) {
	case string:
		if t = strings.TrimSpace(t); t != "" {
			var v *url.URL
			if v, err = url.Parse(t); err == nil {
				parsed = v.String()
			}
		} else {
			parsed = ""
		}
	default:
		err = errors.New(spec.Printer.Sprintf("unsupported type: %[1]s", values.TypeOf(input)))
	}
	return
}

func RelativeUrlParser(spec *Field, input interface{}) (parsed interface{}, err error) {
	if list, ok := input.([]string); ok && len(list) > 0 {
		input = list[0]
	}
	switch t := input.(type) {
	case string:
		if t = strings.TrimSpace(t); t != "" {
			var v *url.URL
			if v, err = url.Parse(t); err == nil {
				if p := bePath.TrimSlashes(v.Path); p != "" {
					parsed = "/" + p
				} else {
					parsed = "/"
				}
			}
		}
	default:
		err = errors.New(spec.Printer.Sprintf("unsupported type: %[1]s", values.TypeOf(input)))
	}
	return
}

func KebabParser(spec *Field, input interface{}) (parsed interface{}, err error) {
	if list, ok := input.([]string); ok && len(list) > 0 {
		input = list[0]
	}
	switch t := input.(type) {
	case string:
		parsed = ToKebab(t)
	default:
		err = errors.New(spec.Printer.Sprintf("unsupported type: %[1]s", values.TypeOf(input)))
	}
	return
}

func KebabOptionParser(spec *Field, input interface{}) (parsed interface{}, err error) {
	if list, ok := input.([]string); ok && len(list) > 0 {
		input = list[0]
	}
	switch t := input.(type) {
	case string:
		if t = ToKebab(t); t != "" {
			if len(spec.ValueOptions) > 0 {
				if slices.Within(t, spec.ValueOptions) {
					parsed = t
					return
				}
			}
		}
		if spec.DefaultValue != nil {
			if dv, ok := spec.DefaultValue.(string); ok {
				parsed = dv
				return
			}
		}
	default:
		err = errors.New(spec.Printer.Sprintf("unsupported type: %[1]s", values.TypeOf(input)))
	}
	return
}

func UuidParser(spec *Field, input interface{}) (parsed interface{}, err error) {
	switch t := input.(type) {
	case string:
		if t = strings.TrimSpace(t); t != "" {
			var id uuid.UUID
			if id, err = uuid.FromString(t); err == nil {
				parsed = id
			}
		}
	default:
		err = errors.New(spec.Printer.Sprintf("unsupported type: %[1]s", values.TypeOf(input)))
	}
	return
}