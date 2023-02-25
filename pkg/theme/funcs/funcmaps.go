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

package funcs

import (
	htmlTemplate "html/template"
	textTemplate "text/template"

	"github.com/iancoleman/strcase"
	"github.com/leekchan/gtf"
)

func TextFuncMap() (funcMap textTemplate.FuncMap) {
	funcMap = HtmlFuncMap()
	return
}

func HtmlFuncMap() (funcMap htmlTemplate.FuncMap) {
	funcMap = htmlTemplate.FuncMap{
		"toCamel":              strcase.ToCamel,
		"toLowerCamel":         strcase.ToLowerCamel,
		"toDelimited":          strcase.ToDelimited,
		"toScreamingDelimited": strcase.ToScreamingDelimited,
		"toKebab":              strcase.ToKebab,
		"toScreamingKebab":     strcase.ToScreamingKebab,
		"toSnake":              strcase.ToSnake,
		"toScreamingSnake":     strcase.ToScreamingSnake,

		"asHTML":     AsHTML,
		"asHTMLAttr": AsHTMLAttr,
		"asCSS":      AsCSS,
		"asJS":       AsJS,

		"fsHash":   FsHash,
		"fsUrl":    FsUrl,
		"fsMime":   FsMime,
		"fsExists": FsExists,

		"numberAsInt": NumberAsInt,

		"add":      Add,
		"sub":      Sub,
		"mul":      Mul,
		"div":      Div,
		"mod":      Mod,
		"addFloat": AddFloat,
		"subFloat": SubFloat,
		"mulFloat": MulFloat,
		"divFloat": DivFloat,

		"mergeClassNames": MergeClassNames,

		"unescapeHTML":     UnescapeHtml,
		"escapeJsonString": EscapeJsonString,
		"escapeHTML":       EscapeHtml,
		"escapeUrlPath":    EscapeUrlPath,

		"element":           Element,
		"elementOpen":       ElementOpen,
		"elementClose":      ElementClose,
		"elementAttributes": ElementAttributes,

		"Nonce": Nonce,

		"isUrl":    IsUrl,
		"isPath":   IsPath,
		"parseUrl": ParseUrl,

		"stringsAsList":              StringsAsList,
		"reverseStrings":             ReverseStrings,
		"sortedKeys":                 SortedKeys,
		"sortedFirstLetters":         SortedFirstLetters,
		"sortedLastNameFirstLetters": SortedLastNameFirstLetters,

		"cmpDateFmt": CompareDateFormats,

		"DebugF": LogDebug,
		"WarnF":  LogWarn,
		"ErrorF": LogError,

		"CmpLang": CmpLang,
	}
	for k, v := range gtf.GtfFuncMap {
		funcMap[k] = v
	}
	return
}