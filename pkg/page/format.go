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

package page

type Format string

const (
	Markdown Format = "md"
	OrgMode  Format = "org"
	Template Format = "tmpl"
	Html     Format = "html"
	HtmlTmpl Format = "html.tmpl"
	Semantic Format = "njn"
)

var Extensions = []string{"org", "md", "html.tmpl", "tmpl", "html", "njn"}

func (pf Format) String() string {
	switch pf {
	case OrgMode:
		return string(OrgMode)
	case Html:
		return string(Html)
	case HtmlTmpl:
		return string(Html) + "." + string(Template)
	case Template:
		return string(Template)
	case Markdown:
		return string(Markdown)
	case Semantic:
		return string(Semantic)
	}
	return "nil"
}