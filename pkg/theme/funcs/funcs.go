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
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/maruel/natural"
	"golang.org/x/net/html"

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

func FsExists(path string) (exists bool) {
	exists = fs.FileExists(path)
	return
}

func FsListFiles(path string) (files []string) {
	if found, err := fs.ListFiles(path); err == nil {
		files = found
	}
	return
}

func FsListAllFiles(path string) (files []string) {
	if found, err := fs.ListAllFiles(path); err == nil {
		files = found
	}
	return
}

func FsListDirs(path string) (files []string) {
	if found, err := fs.ListDirs(path); err == nil {
		files = found
	}
	return
}

func FsListAllDirs(path string) (files []string) {
	if found, err := fs.ListAllDirs(path); err == nil {
		files = found
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

func SortedKeys(v interface{}) (keys []string) {
	t := reflect.TypeOf(v)
	switch t.Kind() {
	case reflect.Map:
		if kt := t.Key(); kt.Kind() == reflect.String {
			value := reflect.ValueOf(v)
			mapKeys := value.MapKeys()
			for _, k := range mapKeys {
				keys = append(keys, k.String())
			}
			sort.Sort(natural.StringSlice(keys))
			return
		}
		log.WarnF("unsupported sortedKeys map key type: %T", v)
	default:
		log.WarnF("unsupported sortedKeys type: %T", v)
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

// func MergeData(datasets ...interface{}) (merged map[string]interface{}) {
// 	merged = make(map[string]interface{})
// 	for _, data := range datasets {
// 		switch typed := data.(type) {
// 		case menu.Item:
// 			merged["Href"] = typed.Href
// 			merged["Text"] = typed.Text
// 			merged["Attributes"] = typed.Attributes
// 			merged["SubMenu"] = typed.SubMenu
// 		case map[string]string:
// 			for k, v := range typed {
// 				merged[k] = v
// 			}
// 		case map[string]template.HTML:
// 			for k, v := range typed {
// 				merged[k] = v
// 			}
// 		case map[string]interface{}:
// 			for k, v := range typed {
// 				merged[k] = v
// 			}
// 		default:
// 			log.ErrorF("merge data - unsupported type: %T", typed)
// 		}
// 	}
// 	return
// }

func CompareDateFormats(format string, a, b time.Time) (same bool) {
	same = a.Format(format) == b.Format(format)
	return
}