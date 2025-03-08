//go:build !exclude_pages_formats && !exclude_pages_format_njn

// Copyright (c) 2025  The Go-Enjin Authors
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

package video

import (
	"fmt"

	"github.com/iancoleman/strcase"

	"github.com/go-corelibs/values"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/pages"
)

const (
	Tag feature.Tag = "njn-fields-video"
)

var (
	_ Field     = (*CField)(nil)
	_ MakeField = (*CField)(nil)
)

type Field interface {
	feature.EnjinField
}

type MakeField interface {
	Make() Field
}

type CField struct {
	feature.CEnjinField
}

func New() (field MakeField) {
	f := new(CField)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = Tag
	return f
}

func (f *CField) Init(this interface{}) {
	f.CEnjinField.Init(this)
}

func (f *CField) Make() Field {
	return f
}

func (f *CField) NjnClass() (tagClass feature.NjnClass) {
	tagClass = feature.InlineNjnClass
	return
}

func (f *CField) NjnFieldNames() (name []string) {
	name = append(name, "video")
	return
}

func (f *CField) PrepareNjnData(re feature.EnjinRenderer, tagName string, field map[string]interface{}) (data map[string]interface{}, err error) {
	if tagName != "video" {
		err = fmt.Errorf(`%v feature does not support tags named: "%v"`, Tag, tagName)
		return
	}

	data = make(map[string]interface{})

	/*
	   <video
	   loop
	   muted
	   autoplay
	   playsinline
	   disableremoteplayback
	   disablepictureinpicture
	   controls

	   preload=<none|metadata|auto>
	   crossorigin=<anonymous|use-credentials>
	   controlslist=[nodownload,nofullscreen,noremoteplayback]

	   poster=<image>
	   src=<video>

	   width=<px>
	   height=<px>
	  >
	  <source src="<video.webm>" type"video/webm" />
	  <source src="<video.mp4>" type"video/mp4" />
	  <p>Please use a browser that supports HTML video in order to watch!</p>
	  </video>
	*/

	data["Type"] = "video"
	present := make(map[string]bool)

	data["Src"], present["src"] = field["src"].(string)
	data["Poster"], present["poster"] = field["poster"].(string)
	if values, ok := field["sources"].([]interface{}); ok {
		var sources []*Source
		for _, value := range values {
			if src, ok := value.(string); ok {
				sources = append(sources, &Source{Src: src})
			} else if attrs, ok := value.(map[string]string); ok {
				item := &Source{}
				if v, ok := attrs["src"]; ok {
					item.Src = v
				}
				if v, ok := attrs["type"]; ok {
					item.Type = v
				}
				if item.Src != "" {
					sources = append(sources, item)
				}
			} else if attrs, ok := value.(map[string]interface{}); ok {
				item := &Source{}
				if v, ok := attrs["src"].(string); ok {
					item.Src = v
				}
				if v, ok := attrs["type"].(string); ok {
					item.Type = v
				}
				if item.Src != "" {
					sources = append(sources, item)
				}
			}
		}
		if present["sources"] = len(sources) > 0; present["sources"] {
			data["Sources"] = sources
		}
	}

	data["Preload"], present["preload"] = field["preload"].(string)
	data["CrossOrigin"], present["crossorigin"] = field["crossorigin"].(string)
	data["ControlsList"], present["controlslist"] = field["controlslist"].(string)

	var width, height int
	if width, present["width"] = values.ExtractIntValue("width", field); present["width"] {
		data["Width"] = width
	}
	if height, present["height"] = values.ExtractIntValue("height", field); present["height"] {
		data["Height"] = height
	}

	boolKeys := []string{"loop", "muted", "autoplay", "playsinline", "controls", "disableremoteplayback", "disablepictureinpicture"}

	for _, key := range boolKeys {
		name := strcase.ToCamel(key)
		if value, ok := field[key].(string); ok {
			data[name] = values.IsTrue(value)
		} else if value, ok := field[key].(bool); ok {
			data[name] = value
		} else if value, ok := field[key].(float64); ok {
			data[name] = value > 0.0
		} else if value, ok := field[key].(int); ok {
			data[name] = value > 0
		}
		_, present[name] = data[name]
	}

	var attrs map[string]interface{}
	if attrs, _, _, err = pages.ParseNjnFieldAttributes(field); err != nil {
		err = fmt.Errorf("error parsing njn field attributes: %v", err)
		return
	} else {
		if present["width"] {
			delete(attrs, "width")
		}
		if present["height"] {
			delete(attrs, "height")
		}
		for _, key := range boolKeys {
			delete(attrs, key)
		}
		if data["Attributes"], err = pages.FinalizeNjnFieldAttributes(attrs); err != nil {
			err = fmt.Errorf("error finalizing njn field attributes: %v", err)
			return
		}
	}

	if text, ok := field["text"].([]interface{}); ok && len(text) > 0 {
		var combined []interface{}
		if combined, err = re.PrepareInlineFields(text); err != nil {
			return
		} else if len(combined) > 0 {
			data["NoSupport"] = combined
		}
	}

	err = pages.FinalizeNjnFieldData(data, field, "src", "poster", "attributes")
	return
}
