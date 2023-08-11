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

package inline

import (
	"fmt"
	"strings"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/slices"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

const (
	Tag feature.Tag = "njn-fields-inline"
)

var (
	TagNames = []string{
		"abbr", "b", "cite", "del", "dfn", "em", "i", "ins", "kbd", "mark",
		"meter", "progress", "q", "s", "small", "strong", "sub", "sup", "u",
		"var", "span", "button", "code",
	}
)

var (
	_ Field     = (*CField)(nil)
	_ MakeField = (*CField)(nil)
)

type Field interface {
	feature.EnjinField
}

type MakeField interface {
	AddTag(name string) MakeField
	RemoveTag(name string) MakeField

	Defaults() MakeField

	Make() Field
}

type CField struct {
	feature.CEnjinField

	supported []string
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

func (f *CField) Defaults() MakeField {
	f.supported = append(
		f.supported,
		TagNames...,
	)
	return f
}

func (f *CField) AddTag(name string) MakeField {
	name = strings.ToLower(name)
	if slices.Present(name, TagNames...) {
		if !slices.Present(name, f.supported...) {
			f.supported = append(f.supported, name)
		}
	} else {
		log.FatalDF(1, `%v feature does not support tags named: "%v"`, Tag, name)
	}
	return f
}

func (f *CField) RemoveTag(name string) MakeField {
	if idx := beStrings.StringIndexInStrings(name, f.supported...); idx >= 0 {
		f.supported = slices.Remove(f.supported, idx)
	}
	return f
}

func (f *CField) Make() Field {
	return f
}

func (f *CField) NjnClass() (tagClass feature.NjnClass) {
	tagClass = feature.InlineNjnClass
	return
}

func (f *CField) NjnFieldNames() (names []string) {
	names = append(names, f.supported...)
	return
}

func (f *CField) PrepareNjnData(re feature.EnjinRenderer, tagName string, field map[string]interface{}) (data map[string]interface{}, err error) {
	if !slices.Present(tagName, f.supported...) {
		err = fmt.Errorf(`%v feature does not support tags named "%v"`, Tag, tagName)
		return
	}

	data = make(map[string]interface{})
	data["Type"] = tagName
	if data["Text"], err = re.PrepareInlineFieldText(field); err != nil {
		return
	}

	err = maps.FinalizeNjnFieldData(data, field, "type", "text")
	return
}