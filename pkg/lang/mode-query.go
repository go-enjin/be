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

package lang

import (
	"fmt"
	"net/http"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/log"
)

var _ Mode = (*QueryMode)(nil)

type QueryMode struct {
	name string
}

type QueryModeBuilder interface {
	SetQueryParameter(name string) QueryModeBuilder

	Make() Mode
}

func NewQueryMode() (q QueryModeBuilder) {
	q = &QueryMode{
		name: "lang",
	}
	return
}

func (q *QueryMode) SetQueryParameter(name string) QueryModeBuilder {
	q.name = name
	return q
}

func (q *QueryMode) Make() Mode {
	return q
}

func (q *QueryMode) ToUrl(defaultTag, tag language.Tag, path string) (translated string) {
	translated = path
	if !language.Compare(defaultTag, tag) {
		translated += fmt.Sprintf("?%v=%v", q.name, tag.String())
	}
	return
}

func (q *QueryMode) FromRequest(defaultTag language.Tag, r *http.Request) (tag language.Tag, path string, ok bool) {
	ok = true
	path = forms.SanitizeRequestPath(r.URL.Path)
	if langValues, found := r.URL.Query()[q.name]; found && len(langValues) >= 1 {
		var err error
		if tag, err = language.Parse(langValues[0]); err != nil {
			log.ErrorF("error parsing language tag: %v", langValues[0])
			tag = language.Und
			ok = false
		}
	} else {
		tag = defaultTag
	}
	return
}