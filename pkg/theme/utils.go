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

package theme

import (
	"fmt"
	"html/template"
)

func AddParseTree(src, dst *template.Template) (err error) {
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

func LookupTemplate(tt *template.Template, names ...string) (tmpl *template.Template) {
	for _, name := range names {
		if tmpl = tt.Lookup(name); tmpl != nil {
			return
		}
	}
	return
}