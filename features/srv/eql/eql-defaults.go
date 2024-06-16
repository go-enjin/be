// Copyright (c) 2024  The Go-Enjin Authors
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

//go:build srv_eql || eql || all

package eql

import (
	"github.com/go-corelibs/slices"
)

var (
	// AlwaysExcludeContextKeys is a pages-pql package-wide setting for a list
	// of keys to always exclude, in addition to the MustExcludeContextKeys
	AlwaysExcludeContextKeys []string

	// AlwaysIncludeContextKeys is a pages-pql package-wide setting for a list
	// of keys to always include, in addition to those specified during the
	// feature build phase with IncludeContextKeys or SetIncludedContextKeys
	AlwaysIncludeContextKeys []string
)

// MustExcludeContextKeys returns the list of context keys that should never be
// indexed; these keys: "Content" and "FrontMatter", plus any package-level
// AlwaysExcludeContextKeys
func MustExcludeContextKeys() []string {
	return slices.Unique(append([]string{"Content", "FrontMatter"}, AlwaysExcludeContextKeys...))
}

// BaseIncludeContextKeys returns the default list of included context keys.
// These keys are currently: "Title" and "Description"
func BaseIncludeContextKeys() []string {
	return []string{"Title", "Description"}
}

// RequiredContextKeys returns the list of context keys that must be indexed
// for basic page request to function at all. The following are the required
// keys: "Shasum", "Language", "Type", "Archetype", "CreatedAt", "UpdatedAt",
// and "Url"
func RequiredContextKeys() []string {
	return []string{"Shasum", "Language", "Type", "Archetype", "CreatedAt", "UpdatedAt", "Url"}
}
