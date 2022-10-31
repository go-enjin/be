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
	"strings"

	"github.com/go-enjin/golang-org-x-text/language"

	bePath "github.com/go-enjin/be/pkg/path"
)

var _ Mode = (*PathMode)(nil)

type PathMode struct {
}

type PathModeBuilder interface {
	Make() Mode
}

func NewPathMode() (p PathModeBuilder) {
	p = &PathMode{}
	return
}

func (p *PathMode) Make() Mode {
	return p
}

func (p *PathMode) ParsePathLang(path string) (tag language.Tag, trimmed string, ok bool) {
	trimmed = path
	tag = language.Und
	parts := strings.Split(bePath.TrimSlashes(path), "/")
	if len(parts) >= 1 {
		if pathTag, err := language.Parse(parts[0]); err == nil {
			tag = pathTag
			trimmed = "/" + strings.Join(parts[1:], "/")
			ok = true
		}
	}
	return
}

func (p *PathMode) ToUrl(defaultTag, tag language.Tag, path string) (translated string) {
	if parsedTag, parsedPath, ok := p.ParsePathLang(path); ok {
		if language.Compare(parsedTag, tag) {
			translated = parsedPath
			return
		}
	}
	translated = path
	if !language.Compare(defaultTag, tag) {
		if len(path) > 0 && path[0] == '/' {
			path = path[1:]
		}
		translated = fmt.Sprintf("/%v/%v", tag.String(), path)
	}
	return
}

func (p *PathMode) FromRequest(defaultTag language.Tag, r *http.Request) (tag language.Tag, path string) {
	var ok bool
	if tag, path, ok = p.ParsePathLang(r.URL.Path); !ok {
		tag = defaultTag
		path = r.URL.Path
	}
	return
}