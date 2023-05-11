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

package details

import (
	"fmt"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/maps"
)

const (
	Tag feature.Tag = "njn-fields-details"
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
	name = append(name, "details")
	return
}

func (f *CField) PrepareNjnData(re feature.EnjinRenderer, tagName string, field map[string]interface{}) (data map[string]interface{}, err error) {
	if tagName != "details" {
		err = fmt.Errorf(`%v feature does not support tags named: "%v"`, Tag, tagName)
		return
	}

	data = make(map[string]interface{})

	data["Type"] = "details"
	data["Summary"], _ = field["summary"]
	if data["Text"], err = re.PrepareContainerFieldText(field); err != nil {
		return
	}

	err = maps.FinalizeNjnFieldData(data, field, "type", "summary", "text")
	return
}