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

package forms

import (
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/net/html"

	bePath "github.com/go-enjin/be/pkg/path"
)

// TrimQueryParams truncates the given string at the first question mark character
func TrimQueryParams(path string) (trimmed string) {
	trimmed, _, _ = strings.Cut(path, "?")
	return
}

// Sanitize uses bluemonday.UGCPolicy to sanitize the given input
func Sanitize(input string) (sanitized string) {
	p := bluemonday.UGCPolicy()
	sanitized = p.Sanitize(input)
	return
}

// StrictSanitize uses bluemonday.StrictPolicy to sanitize the given input
func StrictSanitize(input string) (sanitized string) {
	p := bluemonday.StrictPolicy()
	sanitized = p.Sanitize(input)
	return
}

// Clean unescapes any HTML entities after using Sanitize on the given input
func Clean(input string) (cleaned string) {
	cleaned = html.UnescapeString(Sanitize(input))
	return
}

// StrictClean is the same as Clean except it uses StrictSanitize on the given input
func StrictClean(input string) (cleaned string) {
	cleaned = html.UnescapeString(StrictSanitize(input))
	return
}

// CleanRequestPath splits the given path into segments, using StrictClean on each segment while reassembling the
// cleaned output as an absolute file path (has a leading slash character)
func CleanRequestPath(path string) (cleaned string) {
	if path = bePath.TrimSlashes(TrimQueryParams(path)); path == "" {
		return "/"
	}
	for _, segment := range strings.Split(path, "/") {
		cleaned += "/" + StrictClean(segment)
	}
	return
}

// CleanRelativePath is the same as CleanRequestPath except the cleaned result is a relative file path (has no leading
// slash character)
func CleanRelativePath(path string) (cleaned string) {
	cleaned = CleanRequestPath(path)
	cleaned = strings.TrimPrefix(cleaned, "/")
	return
}

// KebabValue uses StrictClean on the given string, replaces all slashes with dashes and renders the string in
// kebab-cased format
func KebabValue(name string) (cleaned string) {
	cleaned = strings.ReplaceAll(StrictClean(name), "/", "-")
	cleaned = strcase.ToKebab(name)
	return
}

// KebabRelativePath uses CleanRelativePath on the given string, splits it into path segments and renders each segment
// in kebab-cased format
//
// ie: KebabRelativePath("lowerCamel/CamelCased/snake_cased") == "lower-camel/camel-cased/snake-cased"
func KebabRelativePath(path string) (kebab string) {
	path = CleanRelativePath(path)
	for idx, segment := range strings.Split(path, "/") {
		if idx > 0 {
			kebab += "/"
		}
		kebab += strcase.ToKebab(segment)
	}
	return
}