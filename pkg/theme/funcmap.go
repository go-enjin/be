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

package theme

import (
	"html/template"

	"github.com/iancoleman/strcase"
	"github.com/leekchan/gtf"

	"github.com/go-enjin/be/pkg/theme/funcs"
)

func DefaultFuncMap() (funcMap template.FuncMap) {
	funcMap = template.FuncMap{
		"toCamel":              strcase.ToCamel,
		"toLowerCamel":         strcase.ToLowerCamel,
		"toDelimited":          strcase.ToDelimited,
		"toScreamingDelimited": strcase.ToScreamingDelimited,
		"toKebab":              strcase.ToKebab,
		"toScreamingKebab":     strcase.ToScreamingKebab,
		"toSnake":              strcase.ToSnake,
		"toScreamingSnake":     strcase.ToScreamingSnake,

		"asHTML":     funcs.AsHTML,
		"asHTMLAttr": funcs.AsHTMLAttr,
		"asCSS":      funcs.AsCSS,
		"asJS":       funcs.AsJS,
		"fsHash":     funcs.FsHash,
		"fsUrl":      funcs.FsUrl,
		"fsMime":     funcs.FsMime,

		"add":      funcs.Add,
		"sub":      funcs.Sub,
		"mul":      funcs.Mul,
		"div":      funcs.Div,
		"mod":      funcs.Mod,
		"addFloat": funcs.AddFloat,
		"subFloat": funcs.SubFloat,
		"mulFloat": funcs.MulFloat,
		"divFloat": funcs.DivFloat,

		"mergeClassNames": funcs.MergeClassNames,

		"unescapeHTML":     funcs.UnescapeHtml,
		"escapeJsonString": funcs.EscapeJsonString,

		"element":           funcs.Element,
		"elementOpen":       funcs.ElementOpen,
		"elementClose":      funcs.ElementClose,
		"elementAttributes": funcs.ElementAttributes,

		"Nonce": funcs.Nonce,

		"isUrl":    funcs.IsUrl,
		"isPath":   funcs.IsPath,
		"parseUrl": funcs.ParseUrl,

		"sortedKeys": funcs.SortedKeys,

		"DebugF": funcs.LogDebug,
		"WarnF":  funcs.LogWarn,
		"ErrorF": funcs.LogError,
	}
	for k, v := range gtf.GtfFuncMap {
		funcMap[k] = v
	}
	return
}