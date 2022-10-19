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

package funcs

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/log"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

func AsJS(input interface{}) template.JS {
	switch v := input.(type) {
	case string:
		return template.JS(v)
	case template.JS:
		return v
	default:
		return template.JS(fmt.Sprintf("%v", v))
	}
}

func AsCSS(input interface{}) template.CSS {
	switch v := input.(type) {
	case string:
		return template.CSS(v)
	case template.CSS:
		return v
	default:
		return template.CSS(fmt.Sprintf("%v", v))
	}
}

func AsHTML(input interface{}) template.HTML {
	switch v := input.(type) {
	case string:
		return template.HTML(v)
	case template.HTML:
		return v
	default:
		return template.HTML(fmt.Sprintf("%v", v))
	}
}

func AsHTMLAttr(input interface{}) template.HTMLAttr {
	switch v := input.(type) {
	case string:
		return template.HTMLAttr(v)
	case template.HTML:
		return template.HTMLAttr(v)
	default:
		return template.HTMLAttr(fmt.Sprintf("%v", v))
	}
}

func FsHash(path string) (shasum string) {
	shasum, _ = fs.FindFileShasum(path)
	return
}

func FsUrl(path string) (url string) {
	url = path
	if shasum, err := fs.FindFileShasum(path); err == nil {
		url += "?rev=" + shasum
	}
	return
}

func FsMime(path string) (mime string) {
	mime, _ = fs.FindFileMime(path)
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

func LogDebug(format string, argv ...interface{}) (html template.HTML) {
	log.DebugDF(2, format, argv...)
	return
}

func LogWarn(format string, argv ...interface{}) (html template.HTML) {
	log.WarnDF(2, format, argv...)
	return
}

func LogError(format string, argv ...interface{}) (html template.HTML) {
	log.ErrorDF(2, format, argv...)
	return
}