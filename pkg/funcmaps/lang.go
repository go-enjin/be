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

package funcmaps

import (
	"fmt"

	"github.com/go-enjin/golang-org-x-text/language"
)

func CmpLang(a interface{}, other ...interface{}) (equal bool, err error) {
	var aTag language.Tag
	var oTags []language.Tag

	parse := func(v interface{}) (tag language.Tag, err error) {
		switch t := v.(type) {
		case string:
			tag, err = language.Parse(t)
		case language.Tag:
			tag = t
		default:
			err = fmt.Errorf("all arguments must be of either string or language.Tag type")
		}
		return
	}

	if aTag, err = parse(a); err != nil {
		return
	}

	for _, o := range other {
		var oTag language.Tag
		if oTag, err = parse(o); err != nil {
			return
		}
		oTags = append(oTags, oTag)
	}

	equal = language.Compare(aTag, oTags...)
	return
}