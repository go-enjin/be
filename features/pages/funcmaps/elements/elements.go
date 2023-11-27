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

package elements

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/urfave/cli/v2"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/menu"
)

var _ Feature = (*CFeature)(nil)

const Tag feature.Tag = "pages-funcmaps-elements"

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
		"element":           Element,
		"elementOpen":       ElementOpen,
		"elementClose":      ElementClose,
		"elementAttributes": ElementAttributes,
	}
	return
}

func elementTextWork(v interface{}) (html template.HTML, err error) {
	switch t := v.(type) {
	case string:
		html += template.HTML(t)
	case template.HTML:
		html += t
	case []interface{}:
		for _, tt := range t {
			if h, e := elementTextWork(tt); e != nil {
				err = e
				return
			} else {
				html += h
			}
		}
	default:
		err = fmt.Errorf("unknown element text type: %T %+v", t, t)
		return
	}
	return
}

func Element(data map[string]interface{}) (html template.HTML, err error) {
	if eo, e := ElementOpen(data); e != nil {
		err = e
		return
	} else {
		html += eo
	}
	if v, ok := data["Text"]; ok {
		if h, e := elementTextWork(v); e != nil {
			err = e
			return
		} else {
			html += h
		}
	} else {
		err = fmt.Errorf("element data missing Text property: %+v", data)
		return
	}
	if ec, _ := ElementClose(data); ec != "" {
		html += ec
	} else {
		err = fmt.Errorf("element failed to close, yet was able to be opened: %+v", data)
	}
	return
}

func ElementAttributes(value interface{}) (html template.HTMLAttr) {
	var parts []string
	switch data := value.(type) {
	case *menu.Item:
		if target := strings.ToLower(data.Target); target != "" {
			switch target {
			case "_self", "_blank", "_parent", "_top":
				parts = append(parts, `target="`+target+`"`)
			}
		}
		if data.Context != nil {
			if attributes, ok := data.Context.Get("attributes").(map[string]interface{}); ok {
				for _, k := range maps.SortedKeys(attributes) {
					parts = append(parts, fmt.Sprintf(`%s=%q`, k, fmt.Sprintf("%v", attributes[k])))
				}
			}
		}
	case []template.HTMLAttr:
		for _, v := range data {
			parts = append(parts, string(v))
		}
	case map[string]interface{}:
		if ai, found := data["Attributes"]; found && ai != nil {
			switch t := ai.(type) {
			case string:
				parts = append(parts, t)
			case []interface{}:
				for _, i := range t {
					if v, ok := i.(string); ok {
						parts = append(parts, v)
					}
				}
			case map[string]interface{}:
				for k, i := range t {
					if v, ok := i.(string); ok {
						parts = append(parts, fmt.Sprintf(`%v="%v"`, k, v))
					}
				}
			case []template.HTMLAttr:
				for idx, i := range t {
					if idx > 0 {
						html += " "
					}
					html += i
				}
				return
			default:
				log.ErrorF("unknown attributes type: %T %+v", t, ai)
			}
		}
	case nil: // nop
	default:
		log.ErrorF("unknown attributes data type: %T %+v", data, data)
	}
	if len(parts) > 0 {
		html += template.HTMLAttr(strings.Join(parts, " "))
	}
	return
}

func elementOpenWork(data map[string]interface{}, dataType interface{}) (html template.HTML, err error) {
	switch dt := dataType.(type) {
	case string:
		html = "<"
		html += template.HTML(dt)
		if attrs := ElementAttributes(data); len(attrs) > 0 {
			html += " "
			html += template.HTML(attrs)
		}
		html += ">"
	case template.HTML:
		html = "<"
		html += dt
		if attrs := ElementAttributes(data); len(attrs) > 0 {
			html += " "
			html += template.HTML(attrs)
		}
		html += ">"
	case template.HTMLAttr:
		html = "<"
		html += template.HTML(dt)
		if attrs := ElementAttributes(data); len(attrs) > 0 {
			html += " "
			html += template.HTML(attrs)
		}
		html += ">"
	default:
		err = fmt.Errorf("element open invalid type property: %T %+v", dt, dt)
	}
	return
}

func ElementOpen(data map[string]interface{}) (html template.HTML, err error) {
	if dataType, ok := data["Type"]; ok {
		switch typedData := dataType.(type) {
		case []interface{}:
			for _, item := range typedData {
				if h, e := elementOpenWork(data, item); e != nil {
					err = e
					return
				} else {
					html += h
				}
			}
		default:
			html, err = elementOpenWork(data, typedData)
		}
	} else {
		err = fmt.Errorf("element open missing type property: %+v", data)
	}
	return
}

func elementCloseWork(dataType interface{}) (html template.HTML, err error) {
	switch dt := dataType.(type) {
	case string:
		html = "</"
		html += template.HTML(dt)
		html += ">"
	case template.HTML:
		html = "</" + dt + ">"
	case template.HTMLAttr:
		html = "</"
		html += template.HTML(dt)
		html += ">"
	default:
		err = fmt.Errorf("element close unsupported dataType structure: %T", dt)
	}
	return
}

func ElementClose(data map[string]interface{}) (html template.HTML, err error) {
	if dataType, ok := data["Type"]; ok {
		switch typedData := dataType.(type) {
		case []interface{}:
			for _, item := range typedData {
				if h, e := elementCloseWork(item); e != nil {
					err = e
					return
				} else {
					html += h
				}
			}
		default:
			html, err = elementCloseWork(typedData)
		}
	} else {
		err = fmt.Errorf("element open missing type property: %+v", data)
	}
	return
}
