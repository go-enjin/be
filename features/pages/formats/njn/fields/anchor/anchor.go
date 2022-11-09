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

package anchor

import (
	"fmt"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/maps"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

const (
	Tag feature.Tag = "NjnAnchorField"
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
	name = append(name, "a")
	return
}

func (f *CField) PrepareNjnData(re feature.EnjinRenderer, tagName string, field map[string]interface{}) (data map[string]interface{}, err error) {
	if tagName != "a" {
		err = fmt.Errorf(`%v feature does not support tags named: "%v"`, Tag, tagName)
		return
	}

	data = make(map[string]interface{})
	data["Type"] = "a"
	if href, ok := field["href"].(string); ok {
		data["Href"] = href
	} else {
		data["Href"] = "#"
	}
	if dataText, e := re.PrepareInlineFieldText(field); e != nil {
		return
	} else {
		if len(dataText) > 0 {
			data["Text"] = dataText
		} else {
			data["Text"] = []interface{}{data["Href"]}
		}
	}

	decorated := false
	if v, ok := field["decorated"].(string); ok {
		if beStrings.IsTrue(v) {
			decorated = true
		}
	}
	data["Decorated"] = decorated
	if attrs, classes, _, e := maps.ParseNjnFieldAttributes(field); e == nil {
		if decorated {
			classes = append(classes, "decorated")
		}
		if data["Attributes"], e = maps.FinalizeNjnFieldAttributes(attrs); e != nil {
			err = fmt.Errorf("error finalizing njn field attributes: %v", e)
			return
		}
	} else if e != nil {
		err = fmt.Errorf("error parsing njn field attributes: %v", e)
		return
	} else if decorated {
		if data["Attributes"], e = maps.FinalizeNjnFieldAttributes(map[string]interface{}{
			"class": "decorated",
		}); e != nil {
			err = fmt.Errorf("error finalizing njn field attributes: %v", e)
			return
		}
	}

	err = maps.FinalizeNjnFieldData(data, field, "type", "href", "text", "decorated", "attributes")
	return
}