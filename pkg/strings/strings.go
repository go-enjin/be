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
	"regexp"
	"strings"
	"unicode"

	"github.com/iancoleman/strcase"

	"github.com/go-enjin/be/pkg/regexps"
)

func StringsToKebabs(in ...string) (out []string) {
	for _, i := range in {
		out = append(out, strcase.ToKebab(i))
	}
	return
}

func LowerStrings(in ...string) (out []string) {
	for _, i := range in {
		out = append(out, strings.ToLower(i))
	}
	return
}

func StringInSlices(src string, dst ...[]string) bool {
	for _, v := range dst {
		for _, vv := range v {
			if src == vv {
				return true
			}
		}
	}
	return false
}

func StringInStrings(src string, dst ...string) bool {
	for _, v := range dst {
		if src == v {
			return true
		}
	}
	return false
}

func StringIndexInSlice(src string, dst []string) int {
	for i, v := range dst {
		if src == v {
			return i
		}
	}
	return -1
}

func StringIndexInStrings(src string, dst ...string) int {
	for i, v := range dst {
		if src == v {
			return i
		}
	}
	return -1
}

func AnyStringsInStrings(src, tgt []string) (found bool) {
	for _, s := range src {
		for _, t := range tgt {
			if found = s == t; found {
				return
			}
		}
	}
	return
}

func RemoveValueFromSlice(value string, slice *[]string) (modified bool) {
	var found []int
	for jdx, v := range *slice {
		if v == value {
			found = append(found, jdx)
		}
	}
	for _, rdx := range found {
		size := len(*slice)
		if rdx < size && rdx >= 0 {
			if rdx == 0 {
				*slice = (*slice)[1:]
			} else if rdx < size-1 {
				*slice = (*slice)[:size-1]
			} else {
				*slice = append((*slice)[:rdx], (*slice)[rdx:]...)
			}
		}
	}
	modified = modified || len(found) > 0
	return
}

func RemoveIndexFromStrings(idx int, slice []string) []string {
	if idx >= 0 && idx < len(slice) {
		if idx == 0 {
			return slice[1:]
		}
		return append(slice[:idx], slice[idx+1:]...)
	}
	return slice
}

func TitleCase(input string) (output string) {
	first := true
	output = regexps.RxWord.ReplaceAllStringFunc(
		strings.ToLower(input),
		func(word string) string {
			if !first {
				switch word {
				case "with", "in", "of", "at", "a", "the":
					return word
				}
			}
			first = false
			return strcase.ToCamel(word)
		},
	)
	return
}

var RxBasicMimeType = regexp.MustCompile(`^\s*([^\s;]*)\s*.+?\s*$`)

func GetBasicMime(mime string) (basic string) {
	if RxBasicMimeType.MatchString(mime) {
		m := RxBasicMimeType.FindAllStringSubmatch(mime, 1)
		basic = m[0][1]
		return
	}
	basic = mime
	return
}

// QuoteJsonValue will quote everything other than numbers or boolean text
func QuoteJsonValue(in string) (out string) {
	if regexps.RxQuoteStringsOnly.MatchString(in) {
		return strings.ToLower(in)
	}
	out = fmt.Sprintf(`"%v"`, strings.ReplaceAll(in, `"`, `\"`))
	return
}

func EscapeHtmlAttribute(unescaped string) (escaped string) {
	var quote uint8
	switch unescaped[0] {
	case '"', '\'':
		quote = unescaped[0]
		last := len(unescaped) - 1
		if unescaped[last] == quote {
			unescaped = unescaped[1 : last-1]
		}
	}
	escaped = strings.ReplaceAll(unescaped, `"`, "&quot;")
	return
}

func IsTrue(text string) bool {
	switch strings.ToLower(text) {
	case "1", "on", "yes", "y", "true":
		return true
	}
	return false
}

func IsFalse(text string) bool {
	switch strings.ToLower(text) {
	case "0", "nil", "off", "no", "n", "false":
		return true
	}
	return false
}

func TrimQuotes(maybeQuoted string) (unquoted string) {
	if total := len(maybeQuoted); total > 2 {
		// there's enough length for quotes to be possible
		if last := total - 1; maybeQuoted[0] == maybeQuoted[last] {
			// the first and last characters are the same
			switch maybeQuoted[0] {
			case '\'', '`', '"':
				// valid quote detected, trim string
				unquoted = maybeQuoted[1:last]
				return
			}
		}
	}
	unquoted = maybeQuoted
	return
}

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
					unquoted := TrimQuotes(quoted)
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
		if !StringInStrings(part, updated...) {
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

func Empty(value string) (empty bool) {
	empty = regexps.RxEmpty.MatchString(value)
	return
}

func StripTmplTags(value string) (clean string) {
	clean = regexps.RxTmplTags.ReplaceAllString(value, "")
	return
}

func AppendWithSpace(src, add string) (combined string) {
	combined = src
	if add == "" {
		return
	}
	srcLen := len(src)
	if srcLen > 0 {
		switch {
		case unicode.IsPunct(rune(add[0])):
		case unicode.IsSpace(rune(add[0])):
		case unicode.IsSpace(rune(src[srcLen-1])):
		default:
			combined += " "
		}
	}
	combined += add
	return
}

func TrimPrefixes(value string, prefixes ...string) (trimmed string) {
	trimmed = value
	for _, prefix := range prefixes {
		trimmed = strings.TrimPrefix(trimmed, "/")
		trimmed = strings.TrimPrefix(trimmed, prefix)
		trimmed = strings.TrimPrefix(trimmed, "/")
		if trimmed != value {
			// stop at the first trim
			return
		}
	}
	return
}