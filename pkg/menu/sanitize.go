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

package menu

import (
	"strings"

	"github.com/mrz1836/go-sanitize"

	"github.com/go-corelibs/x-text/language"

	"github.com/go-enjin/be/pkg/forms"
)

func cleanInPlace(text, href, lang, icon, image, imgAlt, target *string) {
	*text = forms.StrictSanitize(*text)
	*href = sanitize.URL(*href)
	if *lang != "" {
		if parsed, e := language.Parse(*lang); e == nil {
			*lang = parsed.String()
		} else {
			*lang = ""
		}
	}
	if *icon = forms.StrictSanitize(*icon); *icon != "" {
		sanitize.Custom(*icon, `[^-_ a-zA-Z0-9]`)
	}
	*image = sanitize.URL(*image)
	*imgAlt = forms.StrictSanitize(*imgAlt)

	if clean := strings.ToLower(*target); clean != "" {
		switch clean {
		case "_self", "_blank", "_parent", "_top":
			*target = clean
		default:
			*target = ""
		}
	}
}

func SanitizeMenu[T Menu | EditMenu](input T) {

	if m, ok := interface{}(input).(Menu); ok {
		for _, i := range m {
			cleanInPlace(&i.Text, &i.Href, &i.Lang, &i.Icon, &i.Image, &i.ImgAlt, &i.Target)
			SanitizeMenu(i.SubMenu)
		}
		return
	}

	if m, ok := interface{}(input).(EditMenu); ok {
		for _, i := range m {
			cleanInPlace(&i.Text, &i.Href, &i.Lang, &i.Icon, &i.Image, &i.ImgAlt, &i.Target)
			SanitizeMenu(i.SubMenu)
		}
	}

	return
}
