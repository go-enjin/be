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

package errors

import (
	"fmt"
	"regexp"

	"github.com/go-enjin/be/pkg/log"
)

var (
	rxTmplErr0 = regexp.MustCompile(`(?msi)<a class=\\?"enjin-error\\?" href=\\?"#error\\?">(.+?)</a>`)
	rxTmplErr1 = regexp.MustCompile(`(?msi)at <(.{80,}?)>:`)
)

func Must(err error) {
	if err != nil {
		log.ErrorDF(1, "%v", err)
		panic(err)
	}
}

func ExtractErrSummary(text string) (summary string) {
	/*
	   content.html.tmpl:1:3: executing \"content.html.tmpl\" at {{ ...... }}

	   html/template error [content.html.tmpl]: "
	   template: content.html.tmpl:1:3: executing \"content.html.tmpl\" at
	   <......>: error calling njn: json syntax error: <a class=\"enjin-error\" href=\"#error\">
	   [5380:174:2] invalid character '[' after object key
	   </a>"
	*/

	if m := rxTmplErr0.FindAllStringSubmatch(text, -1); len(m) > 0 {
		if len(m) == 1 {
			summary = m[0][1]
			return
		}
		summary = fmt.Sprintf("[ERR=%d]:", len(m))
		for idx, mm := range m {
			if idx > 0 {
				summary += ";"
			}
			summary += " " + mm[1]
		}
		return
	} else if rxTmplErr1.MatchString(text) {
		return rxTmplErr1.ReplaceAllString(text, "at <...>:")
	}

	return text
}
