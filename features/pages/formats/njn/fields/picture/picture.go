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

package picture

import (
	"fmt"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/maps"
)

const (
	Tag feature.Tag = "NjnPictureField"
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
	return f
}

func (f *CField) Tag() feature.Tag {
	return Tag
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
	name = append(name, "picture")
	return
}

func (f *CField) PrepareNjnData(re feature.EnjinRenderer, tagName string, field map[string]interface{}) (data map[string]interface{}, err error) {
	if tagName != "picture" {
		err = fmt.Errorf(`%v feature does not support tags named: "%v"`, Tag, tagName)
		return
	}

	data = make(map[string]interface{})

	data["Type"] = "picture"
	dataDefault := make(map[string]interface{})
	if defaultMap, ok := field["default"].(map[string]interface{}); !ok {
		err = fmt.Errorf("picture field missing default image: %v", field)
		return
	} else {
		dataDefault["Type"] = "img"
		if dataDefault["Src"], ok = defaultMap["src"]; !ok {
			err = fmt.Errorf("picture field missing default img src: %v", defaultMap)
			return
		}
		if err = maps.FinalizeNjnFieldData(dataDefault, defaultMap, "type", "src"); err != nil {
			err = fmt.Errorf("error finalizing njn field data: %v", err)
			return
		}
		data["Default"] = dataDefault
	}
	var dataSources []map[string]interface{}
	if sources, ok := field["sources"].([]interface{}); ok {
		for _, si := range sources {
			if source, ok := si.(map[string]interface{}); ok {
				src := make(map[string]interface{})
				src["Type"] = "source"
				if src["Srcset"], ok = source["srcset"]; !ok {
					err = fmt.Errorf("picture field source missing srcset: %v", field)
					return
				}
				if src["Media"], ok = source["media"]; !ok {
					err = fmt.Errorf("picture field source missing media: %v", field)
					return
				}
				dataSources = append(dataSources, src)
			}
		}
	}
	data["Sources"] = dataSources
	err = maps.FinalizeNjnFieldData(data, field, "type", "sources", "default")
	return
}