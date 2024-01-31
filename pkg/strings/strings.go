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

package strings

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/go-corelibs/slices"
	clStrings "github.com/go-corelibs/strings"
	"github.com/go-enjin/be/pkg/regexps"
)

//func TitleCase(input string) (output string) {
//	first := true
//	output = regexps.RxWord.ReplaceAllStringFunc(
//		strings.ToLower(input),
//		func(word string) string {
//			if !first {
//				switch word {
//				case "with", "in", "of", "at", "a", "the":
//					return word
//				}
//			}
//			first = false
//			return strcase.ToCamel(word)
//		},
//	)
//	return
//}

func ParseHtmlTagAttributes(input interface{}) (attributes map[string]interface{}, err error) {
	attributes = make(map[string]interface{})

	parseAndUpdate := func(raw string) (e error) {
		parts := regexps.RxSplitHtmlTagAttributes.Split(raw, -1)
		for _, part := range parts {
			if regexps.RxParseHtmlTagKeyOnly.MatchString(part) {
				if m := regexps.RxParseHtmlTagKeyOnly.FindAllStringSubmatch(part, -1); m != nil {
					key := m[0][1]
					attributes[key] = nil
				}
			} else if regexps.RxParseHtmlTagKeyValue.MatchString(part) {
				if m := regexps.RxParseHtmlTagKeyValue.FindAllStringSubmatch(part, -1); m != nil {
					key, quoted := m[0][1], m[0][2]
					unquoted := clStrings.TrimQuotes(quoted)
					attributes[key] = unquoted
				}
			} else {
				e = fmt.Errorf(`unsupported HTMLAttr format: %v`, part)
				return
			}
		}
		return
	}

	switch v := input.(type) {

	case string:
		err = parseAndUpdate(v)
	case template.HTML:
		err = parseAndUpdate(string(v))
	case template.HTMLAttr:
		err = parseAndUpdate(string(v))
	case []string:
		for _, tha := range v {
			if err = parseAndUpdate(tha); err != nil {
				return
			}
		}
	case []template.HTML:
		for _, tha := range v {
			if err = parseAndUpdate(string(tha)); err != nil {
				return
			}
		}
	case []template.HTMLAttr:
		for _, tha := range v {
			if err = parseAndUpdate(string(tha)); err != nil {
				return
			}
		}

	default:
		err = fmt.Errorf("unknown input type: (%T) %+v", v, v)
	}
	return
}

func UniqueFromSpaceSep(value string, original []string) (updated []string) {
	updated = original
	parts := regexps.RxSplitHtmlTagAttributes.Split(value, -1)
	for _, part := range parts {
		if !slices.Present(part, updated...) {
			updated = append(updated, part)
		}
	}
	return
}

func AddClassNamesToNjnBlock(data map[string]interface{}, classes ...string) map[string]interface{} {
	if v, ok := data["class"]; ok {
		var unique []string
		switch t := v.(type) {
		case string:
			parts := regexps.RxSplitHtmlTagAttributes.Split(t, -1)
			for _, p := range parts {
				unique = UniqueFromSpaceSep(p, unique)
			}
			for _, c := range classes {
				unique = UniqueFromSpaceSep(c, unique)
			}
		case []interface{}:
			for _, iface := range t {
				if s, ok := iface.(string); ok {
					unique = UniqueFromSpaceSep(s, unique)
				}
			}
		}
		data["class"] = strings.Join(unique, " ")
	} else {
		data["class"] = strings.Join(classes, " ")
	}
	return data
}
