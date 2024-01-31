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
	"strings"

	"github.com/iancoleman/strcase"

	clStrings "github.com/go-corelibs/strings"
)

const (
	NilTag Tag = ""
)

// Tag is the primary identifier type for enjin Feature implementations
type Tag string

// IsNil returns true if the tag is empty
func (t Tag) IsNil() (empty bool) {
	empty = t == NilTag
	return
}

// Equal returns true if the kebab-case of this Tag is the same as the kebab-case of the other Tag
func (t Tag) Equal(other Tag) (same bool) {
	same = t.Kebab() == other.Kebab()
	return
}

// String returns the Tag as a string
func (t Tag) String() string {
	return string(t)
}

// Camel returns the Tag as a CamelCased string
func (t Tag) Camel() string {
	return strcase.ToCamel(string(t))
}

// SpacedCamel returns the tag as a Spaced Camel Cased string (first letters capitalized, separated by spaces)
func (t Tag) SpacedCamel() string {
	return clStrings.ToSpacedCamel(t.Kebab())
}

// Kebab returns the Tag as a kebab-cased string
func (t Tag) Kebab() string {
	return strcase.ToKebab(string(t))
}

// Spaced returns the tag as a space cased string (all lowercase, separated by spaces)
func (t Tag) Spaced() string {
	return strings.ReplaceAll(t.Kebab(), "-", " ")
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
