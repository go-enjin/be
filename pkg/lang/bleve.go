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

package lang

import (
	"github.com/go-corelibs/x-text/language"

	"github.com/go-corelibs/slices"
)

var bleveLocaleAnalyzers = []string{
	"ar", "bg", "ca", "cjk", "ckb", "cs", "da", "de", "el", "en", "es", "eu",
	"fa", "fi", "fr", "ga", "gl", "hi", "hu", "hy", "id", "in", "it", "nl",
	"no", "pt", "ro", "ru", "sv", "tr",
}

func BleveSupportedAnalyzer(tag language.Tag) (ok bool) {
	ok = slices.Present(tag.String(), bleveLocaleAnalyzers...)
	return
}
