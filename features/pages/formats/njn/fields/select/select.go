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

package _select

import (
	"fmt"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/maps"
)

const (
	Tag feature.Tag = "NjnSelectField"
)

var (
	_ Field     = (*CField)(nil)
	_ MakeField = (*CField)(nil)
)

var _instance *CField

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
	if _instance == nil {
		_instance = new(CField)
		_instance.Init(_instance)
	}
	field = _instance
	return
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

func (f *CField) NjnFieldNames() (name []string) {
	name = append(name, "select")
	return
}

func (f *CField) PrepareNjnData(re feature.EnjinRenderer, tagName string, field map[string]interface{}) (data map[string]interface{}, err error) {
	if tagName != "select" {
		err = fmt.Errorf(`%v feature does not support tags named: "%v"`, Tag, tagName)
		return
	}

	data = make(map[string]interface{})
	data["Type"] = "select"

	if list, ok := field["placeholder"].([]interface{}); ok {
		if html, e := re.RenderInlineFieldList(list); e != nil {
			err = fmt.Errorf("error rendering select placeholder: %v", e)
			return
		} else {
			data["Placeholder"] = html
		}
	} else if text, ok := field["placeholder"].(string); ok {
		data["Placeholder"] = text
	}

	if list, ok := field["options"].([]interface{}); ok {
		if html, e := re.RenderInlineFieldList(list); e != nil {
			err = fmt.Errorf("error rendering select options: %v", e)
			return
		} else {
			data["Options"] = html
		}
	}

	err = maps.FinalizeNjnFieldData(data, field, "type", "placeholder", "options")
	return
}