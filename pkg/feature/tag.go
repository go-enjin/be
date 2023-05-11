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

package feature

import (
	"github.com/iancoleman/strcase"
)

// Tag is the primary identifier type for enjin Feature implementations
type Tag string

// String returns the Tag as a string
func (t Tag) String() string {
	return string(t)
}

// Camel returns the Tag as a CamelCased string
func (t Tag) Camel() string {
	return strcase.ToCamel(string(t))
}

// Kebab returns the Tag as a kebab-cased string
func (t Tag) Kebab() string {
	return strcase.ToKebab(string(t))
}

// ScreamingKebab returns the Tag as a SCREAMING-KEBAB-CASED string
func (t Tag) ScreamingKebab() string {
	return strcase.ToScreamingKebab(string(t))
}

// Snake returns the tag as a snake_cased string
func (t Tag) Snake() string {
	return strcase.ToSnake(string(t))
}

// ScreamingSnake returns the Tag as a SCREAMING_SNAKE_CASED string
func (t Tag) ScreamingSnake() string {
	return strcase.ToScreamingSnake(string(t))
}