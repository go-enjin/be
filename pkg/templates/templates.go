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

package templates

import (
	"fmt"
	htmlTemplate "html/template"
	textTemplate "text/template"
)

func AddHtmlParseTree(src, dst *htmlTemplate.Template) (err error) {
	for _, srcTmpl := range src.Templates() {
		if _, err = dst.AddParseTree(srcTmpl.Name(), srcTmpl.Tree); err != nil {
			err = fmt.Errorf("error adding %v parse tree to %v template: %v", srcTmpl.Name(), dst.Name(), err)
			return
		} else {
			// log.TraceF("added %v parse tree to %v template", srcTmpl.Name(), dst.Name())
		}
	}
	return
}

func LookupHtmlTemplate(tt *htmlTemplate.Template, names ...string) (tmpl *htmlTemplate.Template) {
	tt.Templates()
	for _, name := range names {
		if tmpl = tt.Lookup(name); tmpl != nil {
			return
		}
	}
	return
}

func AddTextParseTree(src, dst *textTemplate.Template) (err error) {
	for _, srcTmpl := range src.Templates() {
		if _, err = dst.AddParseTree(srcTmpl.Name(), srcTmpl.Tree); err != nil {
			err = fmt.Errorf("error adding %v parse tree to %v template: %v", srcTmpl.Name(), dst.Name(), err)
			return
		} else {
			// log.TraceF("added %v parse tree to %v template", srcTmpl.Name(), dst.Name())
		}
	}
	return
}

func LookupTextTemplate(tt *textTemplate.Template, names ...string) (tmpl *textTemplate.Template) {
	tt.Templates()
	for _, name := range names {
		if tmpl = tt.Lookup(name); tmpl != nil {
			return
		}
	}
	return
}