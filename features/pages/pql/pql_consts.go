//go:build page_pql || pages || all

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

package pql

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
// indexed for Page Query Language purposes; these keys are currently:
// "Content" and "FrontMatter"
func MustExcludeContextKeys() []string {
	return []string{"content", "front-matter"}
}

// BaseIncludeContextKeys returns the default list of included context keys for
// Page Query Language purposes; these keys are currently: "Type", "Language",
// "Url", "Title" and "Description"
func BaseIncludeContextKeys() []string {
	return []string{"type", "language", "url", "title", "description", "archetype"}
}

const (
	gAllUrlsBucketName           = "page_urls_all"
	gPageUrlsBucketName          = "page_urls"
	gPageStubsBucketName         = "page_stubs"
	gPermalinksBucketName        = "page_permalinks"
	gPageRedirectionsBucketName  = "page_redirections"
	gPageTranslationsBucketName  = "page_translations"
	gPageTranslatedByBucketName  = "page_translated_by"
	gPageContextValuesBucketName = "page_context_values"
)
