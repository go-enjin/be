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

package regexps

import "regexp"

var RxEmptySpace = regexp.MustCompile(`\s+`)
var RxLanguageKey = regexp.MustCompile(`language:(\*|[a-z][-a-zA-Z]+)\s*`)

var RxKeywordPunctuation = regexp.MustCompile(`([-'])`)

const KeywordPattern = `([a-zA-Z0-9]+?[-'a-zA-Z0-9]*[a-zA-Z0-9]+?|[a-zA-Z0-9]+)`

var RxKeyword = regexp.MustCompile(`^` + KeywordPattern + `$`)
var RxKeywords = regexp.MustCompile(`\b` + KeywordPattern + `\b`)

const PathVarPattern = `\$\{([^}]+)\}`

var RxPathVar = regexp.MustCompile(`^` + PathVarPattern + `$`)
var RxPathVars = regexp.MustCompile(`\b` + PathVarPattern + `\b`)

var RxNonWord = regexp.MustCompile(`[^a-zA-Z0-9]`)