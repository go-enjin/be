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

package regexps

import (
	"regexp"
)

var (
	RxWord                   = regexp.MustCompile(`\w+`)
	RxQuoteStringsOnly       = regexp.MustCompile(`(?i)^(true|false|-?\d*[.,]?\d*)$`)
	RxParseHtmlTagKeyOnly    = regexp.MustCompile(`^([a-zA-Z][-a-zA-Z0-9]+)$`)
	RxParseHtmlTagKeyValue   = regexp.MustCompile(`^([a-zA-Z][-a-zA-Z0-9]+)=(.+?)$`)
	RxSplitHtmlTagAttributes = regexp.MustCompile(`\s+`)
	RxEmpty                  = regexp.MustCompile(`(?ms)\A\s*\z`)
	RxTmplTags               = regexp.MustCompile(`\{\{.+?}}`)
	RxHash10                 = regexp.MustCompile(`^\s*([a-fA-F0-9]{10})\s*$`)
	RxAtLeastSixDigits       = regexp.MustCompile(`^\s*([0-9]{6,})\s*$`)
)