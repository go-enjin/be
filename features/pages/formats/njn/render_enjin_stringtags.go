//go:build !exclude_pages_formats && !exclude_pages_format_njn

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

package njn

import (
	"html/template"
)

func (re *RenderEnjin) PrepareStringTags(text string) (data []interface{}, err error) {
	/*
		 BUG: html.Parse is eating the prefixing space between `</a> section!`
		 TODO: perform shortcode translation only on actual HTML tag contents, not the whole thing
				 - html.Parse is used because we want to parse the contents of tags for shortcodes
				 - how else can this be done?
				 - regexp to extract the content within tags, but this is a nightmare
				 - thus html.Parse was used in the first place
				 - tried using goquery just now and it has the same space-eating problems
				 - this time around, let's just accept the input as plain text and translate that
	*/
	parsed := re.Enjin.TranslateShortcodes(text, re.ctx)
	data = []interface{}{template.HTML(parsed)}
	return
}
