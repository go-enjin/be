//go:build page_funcmaps || pages || all

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

package strings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/mime"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "pages-funcmaps-strings"

type Feature interface {
	feature.Feature
	feature.FuncMapProvider
}

type MakeFeature interface {
	Make() Feature
}

type CFeature struct {
	feature.CFeature
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	return
}

func (f *CFeature) Make() (feat Feature) {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	return
}

func (f *CFeature) Shutdown() {

}

func (f *CFeature) MakeFuncMap(ctx beContext.Context) (fm feature.FuncMap) {
	fm = feature.FuncMap{
		"toString":         ToString,
		"isEmptyString":    IsEmptyString,
		"mergeClassNames":  MergeClassNames,
		"unescapeHTML":     UnescapeHtml,
		"escapeJsonString": EscapeJsonString,
		"escapeHTML":       EscapeHtml,
		"escapeQuotes":     EscapeQuotes,
		"asMatterValue":    AsMatterValue,
		"escapeUrlPath":    EscapeUrlPath,
		"isUrl":            IsUrl,
		"isPath":           IsPath,
		"parseUrl":         ParseUrl,
		"baseName":         BaseName,
		"pruneCharset":     mime.PruneCharset,
		"trimSpace":        strings.TrimSpace,
		"trimPrefix":       strings.TrimPrefix,
		"trimSuffix":       strings.TrimSuffix,
		"rplString":        ReplaceString,
		"hasPrefix":        strings.HasPrefix,
		"hasSuffix":        strings.HasSuffix,
		"repeatString":     strings.Repeat,
		"centerString":     CenterString,
	}
	return
}

func IsEmptyString(input interface{}) (empty bool) {
	empty = strings.TrimSpace(ToString(input)) == ""
	return
}

func ToString(input interface{}) (output string) {
	switch t := input.(type) {
	case string:
		output = t
	case template.HTML:
		output = string(t)
	case template.CSS:
		output = string(t)
	case template.JS:
		output = string(t)
	case template.HTMLAttr:
		output = string(t)
	case template.JSStr:
		output = string(t)
	case template.URL:
		output = string(t)
	case template.Srcset:
		output = string(t)
	default:
		output = fmt.Sprintf("%v", input)
	}
	return
}

func EscapeUrlPath(input interface{}) (escaped string) {
	switch t := input.(type) {
	case string:
		escaped = url.PathEscape(t)
	case []byte:
		escaped = url.PathEscape(string(t))
	case template.HTML:
		escaped = url.PathEscape(string(t))
	case template.HTMLAttr:
		escaped = url.PathEscape(string(t))
	default:
		escaped = fmt.Sprintf("%v", t)
	}
	return
}

func MergeClassNames(names ...interface{}) (result template.HTML) {
	var accepted []string
	for _, name := range names {
		switch nameTyped := name.(type) {
		case string:
			accepted = beStrings.UniqueFromSpaceSep(nameTyped, accepted)

		case map[string]interface{}:
			if v, ok := nameTyped["Class"]; ok {
				if vString, ok := v.(string); ok {
					accepted = beStrings.UniqueFromSpaceSep(vString, accepted)
				} else if vList, ok := v.([]interface{}); ok {
					for _, vlItem := range vList {
						if vliString, ok := vlItem.(string); ok {
							accepted = beStrings.UniqueFromSpaceSep(vliString, accepted)
						}
					}
				} else {
					log.ErrorF("unsupported class structure: %T %v", v, v)
				}
			}

		default:
			log.ErrorF("unsupported input structure: %T %v", name, name)
		}
	}
	result = template.HTML(strings.Join(accepted, " "))
	return
}

func EscapeJsonString(input interface{}) (value string) {
	var dst bytes.Buffer
	switch t := input.(type) {
	case string:
		json.HTMLEscape(&dst, []byte(t))
	default:
		json.HTMLEscape(&dst, []byte(fmt.Sprintf("%v", t)))
	}
	value = dst.String()
	return
}

func EscapeHtml(input interface{}) (out template.HTML) {
	switch t := input.(type) {
	case string:
		out = template.HTML(html.EscapeString(t))
	case template.HTML:
		out = template.HTML(html.EscapeString(string(t)))
	default:
		out = template.HTML(html.EscapeString(fmt.Sprintf("%v", t)))
	}
	return
}

func EscapeQuotes(input interface{}) (out string) {
	switch t := input.(type) {
	case string:
		out = strings.ReplaceAll(t, `"`, `\"`)
	case template.HTML:
		out = strings.ReplaceAll(string(t), `"`, `\"`)
	default:
		out = strings.ReplaceAll(fmt.Sprintf("%v", t), `"`, `\"`)
	}
	return
}

func AsMatterValue(input interface{}) (out string) {
	switch t := input.(type) {
	case string:
		out = strings.ReplaceAll(html.UnescapeString(t), `"`, `\"`)
	case template.HTML:
		out = strings.ReplaceAll(html.UnescapeString(string(t)), `"`, `\"`)
	default:
		out = strings.ReplaceAll(html.UnescapeString(fmt.Sprintf("%v", t)), `"`, `\"`)
	}
	return
}

func UnescapeHtml(input interface{}) (out template.HTML) {
	switch t := input.(type) {
	case string:
		out = template.HTML(html.UnescapeString(t))
	case template.HTML:
		out = template.HTML(html.UnescapeString(string(t)))
	default:
		out = template.HTML(html.UnescapeString(fmt.Sprintf("%v", t)))
	}
	return
}

func ParseUrl(value string) (u *url.URL) {
	if v, err := url.Parse(value); err != nil {
		log.ErrorF("error parsing url: %v", err)
	} else {
		u = v
	}
	return
}

func IsUrl(value string) (ok bool) {
	if u, err := url.Parse(value); err == nil {
		ok = u.Scheme != "" && u.Host != ""
	}
	return
}

func IsPath(value string) (ok bool) {
	if u, err := url.Parse(value); err == nil {
		ok = u.Scheme == "" && u.Host == "" && u.Path != ""
	}
	return
}

func BaseName(path string) (name string) {
	name = filepath.Base(path)
	return
}

func ReplaceString(input string, argv ...string) (output string, err error) {
	if len(argv)%2 != 0 {
		err = fmt.Errorf("unbalanced argument list: %#+v", argv)
		return
	}
	rpl := map[string]string{}
	for i := 0; i < len(argv); i += 2 {
		rpl[argv[i]] = argv[i+1]
	}
	output = input
	for k, v := range rpl {
		output = strings.ReplaceAll(output, k, v)
	}
	return
}

func CenterString(input string, width int) (centered string) {
	var size int
	if size = len(input); size >= width {
		return input
	}
	delta := width - size
	half := delta / 2
	remainder := delta % 2
	centered = strings.Repeat(" ", half)
	centered += input
	centered += strings.Repeat(" ", half+remainder)
	return
}