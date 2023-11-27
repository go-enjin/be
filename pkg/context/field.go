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

package context

import (
	"fmt"

	"github.com/go-enjin/golang-org-x-text/message"
)

type Field struct {
	Key      string `json:"key"`
	Tab      string `json:"tab"`
	Label    string `json:"label"`
	Format   string `json:"format"`
	Category string `json:"category"`

	Input    string `json:"input"`
	Weight   int    `json:"weight"`
	Required bool   `json:"required,omitempty"`

	Step         float64           `json:"step"`
	Minimum      float64           `json:"minimum"`
	Maximum      float64           `json:"maximum"`
	Placeholder  string            `json:"placeholder"`
	DefaultValue interface{}       `json:"default-value"`
	ValueLabels  map[string]string `json:"value-labels"`
	ValueOptions []string          `json:"value-options"`

	LockNonEmpty bool `json:"lock-non-empty"`
	NoResetValue bool `json:"no-reset-value"`

	Printer *message.Printer `json:"-"`
	Parse   Parser           `json:"-"`
}

func ParseField(data map[string]interface{}) (field *Field) {
	field = &Field{}
	field.Key, _ = data["key"].(string)
	field.Tab, _ = data["tab"].(string)
	field.Label, _ = data["label"].(string)
	field.Format, _ = data["format"].(string)
	field.Category, _ = data["category"].(string)
	field.Input, _ = data["input"].(string)
	field.Weight, _ = data["weight"].(int)
	field.Required, _ = data["required,omitempty"].(bool)
	field.Step, _ = data["step"].(float64)
	field.Minimum, _ = data["minimum"].(float64)
	field.Maximum, _ = data["maximum"].(float64)
	field.Placeholder, _ = data["placeholder"].(string)
	field.DefaultValue, _ = data["default-value"].(interface{})
	field.ValueOptions, _ = data["value-options"].([]string)
	field.LockNonEmpty, _ = data["lock-non-empty"].(bool)
	field.NoResetValue, _ = data["no-reset-value"].(bool)

	field.ValueLabels = make(map[string]string)
	if v, ok := data["value-labels"]; ok {
		switch t := v.(type) {
		case map[string]string:
			field.ValueLabels = t
		case map[string]interface{}:
			for k, vv := range t {
				field.ValueLabels[k] = fmt.Sprintf("%v", vv)
			}
		}
	}
	return
}

func (f *Field) Copy() (copied *Field) {
	clone := *f
	copied = &clone
	return
}