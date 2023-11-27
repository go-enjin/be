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

package headers

import (
	"net/http"

	"github.com/go-enjin/be/pkg/log"
)

type ModifyHeadersFn = func(r *http.Request, headers map[string]string) map[string]string

type ModifyAfterUseHeadersFn = func(w http.ResponseWriter, r *http.Request)

func GetDefault() (headers map[string]string) {
	headers = map[string]string{
		"Referrer-Policy":           "strict-origin-when-cross-origin",
		"X-Xss-Protection":          `1; mode=block`,
		"X-Content-Type-Options":    "nosniff",
		"Strict-Transport-Security": `max-age=31536000; includeSubDomains`,
		"Cache-Control":             "no-cache",
	}
	return
}

func ModifyMiddleware(fn ModifyHeadersFn) (mw func(next http.Handler) http.Handler) {
	return func(next http.Handler) http.Handler {
		log.DebugF("including modify headers middleware")
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headers := GetDefault()
			if fn != nil {
				headers = fn(r, headers)
			}
			for k, v := range headers {
				w.Header().Set(k, v)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func ModifyAfterUseMiddleware(fn ModifyAfterUseHeadersFn) (mw func(next http.Handler) http.Handler) {
	return func(next http.Handler) http.Handler {
		log.DebugF("including modify after-use headers middleware")
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fn(w, r)
			next.ServeHTTP(w, r)
		})
	}
}
