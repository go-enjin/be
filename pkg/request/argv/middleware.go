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

package argv

import (
	"net/http"
	"strings"

	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/log"
)

const (
	RouteOneArg  = `:{arg0:[^/]+}`
	RouteTwoArgs = `:{arg0:[^/]+}/:{arg1:[^/]+}`
	RoutePgntn   = `{num-per-page:\d+}/{page-number:\d+}`
)

func ProcessRequest(r *http.Request) (argv *Argv, modified *http.Request) {
	// TODO: figure out more reliable way to modify r.RequestURI and consolidate
	urlPath := forms.CleanRequestPath(r.URL.Path)
	reqPath := urlPath
	if argv = DecodeHttpRequest(r); argv != nil {
		r = argv.Set(r)
		reqPath = argv.Path
		log.TraceF("parsed request argv: %v", argv)
	}
	if reqPath == "" {
		reqPath = "/"
	}
	r.URL.Path = reqPath
	r.RequestURI = strings.Replace(r.RequestURI, urlPath, reqPath, 1)
	modified = r.Clone(r.Context())
	return
}

func Middleware(next http.Handler) (this http.Handler) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, r = ProcessRequest(r)
		next.ServeHTTP(w, r)
	})
}
