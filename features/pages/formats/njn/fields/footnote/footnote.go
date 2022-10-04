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

package footnote

import (
	"fmt"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/maps"
)

const (
	Tag feature.Tag = "NjnFootnoteField"
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
	name = append(name, "footnote")
	return
}

func (f *CField) PrepareNjnData(re feature.EnjinRenderer, tagName string, field map[string]interface{}) (data map[string]interface{}, err error) {
	if tagName != "footnote" {
		err = fmt.Errorf(`%v feature does not support tags named: "%v"`, Tag, tagName)
		return
	}

	data = make(map[string]interface{})

	data["Type"] = "footnote"

	data["BlockIndex"] = re.GetBlockIndex()
	data["Index"] = re.AddFootnote(re.GetBlockIndex(), data)

	if text, ok := field["text"].(string); ok {
		data["Text"] = text
	} else {
		err = fmt.Errorf("footnote %v missing text (string only)", data["Index"])
		return
	}

	if note, ok := field["note"].([]interface{}); ok {
		if data["Note"], err = re.RenderInlineFields(note); err != nil {
			return
		}
	} else {
		err = fmt.Errorf("footnote %v missing note", data["Index"])
		return
	}

	err = maps.FinalizeNjnFieldData(data, field, "type", "text", "note")
	return
}