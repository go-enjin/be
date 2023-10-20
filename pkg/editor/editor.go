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

package editor

import (
	"regexp"
	"strings"

	"github.com/go-enjin/golang-org-x-text/language"
)

func MakeLangCodePath(code, dirs string) (dirsPath string) {
	if code != "" {
		var parts []string
		if code != language.Und.String() {
			parts = append(parts, code)
		}
		if dirs != "" {
			parts = append(parts, dirs)
		}
		dirsPath = strings.Join(parts, "/")
	} else {
		dirsPath = "."
	}
	return
}

var (
	rxMenuKeyParts = regexp.MustCompile(`\.([a-zA-Z][-a-zA-Z0-9]*[a-zA-Z0-9]+)(?:\[(\d+)\])?`)
)

func MakeScrollToKey(changeID string) (key string) {
	// from ".menu[0].sub-menu[2].sub-menu[0].expand" -> "menu--0--2--0"
	if rxMenuKeyParts.MatchString(changeID) {
		m := rxMenuKeyParts.FindAllStringSubmatch(changeID, -1)
		//log.DebugF("matched: %#+v\n", m)
		for idx, mm := range m {
			if idx == 0 {
				key = mm[1]
			}
			if mm[2] != "" {
				key += "--" + mm[2]
			}
		}
	}
	return
}