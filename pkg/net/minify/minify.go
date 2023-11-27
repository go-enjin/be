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

package minify

import (
	"net/http"
	"regexp"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/json"
	"github.com/tdewolff/minify/v2/svg"
	"github.com/tdewolff/minify/v2/xml"
)

var (
	excluded = make(map[string]*regexp.Regexp)
)

func ExcludeRegexp(pattern string) (err error) {
	if excluded[pattern], err = regexp.Compile(pattern); err != nil {
		delete(excluded, pattern)
	}
	return
}

func IsExcluded(path string) (skip bool) {
	for _, rx := range excluded {
		if rx.MatchString(path) {
			return true
		}
	}
	return false
}

func NewInstance() *minify.M {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)
	m.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), json.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), xml.Minify)
	return m
}

func Middleware(next http.Handler) http.Handler {
	instance := NewInstance()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if IsExcluded(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		mw := instance.ResponseWriter(w, r)
		defer mw.Close()
		next.ServeHTTP(mw, r)
	})
}
