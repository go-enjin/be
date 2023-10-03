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

	Step         float64     `json:"step"`
	Minimum      float64     `json:"minimum"`
	Maximum      float64     `json:"maximum"`
	Placeholder  string      `json:"placeholder"`
	DefaultValue interface{} `json:"default-value"`
	ValueOptions []string    `json:"value-options"`

	LockNonEmpty bool `json:"lock-non-empty"`
	NoResetValue bool `json:"no-reset-value"`

	Printer *message.Printer `json:"-"`
	Parse   Parser           `json:"-"`
}