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

const (
	NoPermissionsPolicy = "accelerometer=(), ambient-light-sensor=(), autoplay=(), battery=(), camera=(), cross-origin-isolated=(), display-capture=(), document-domain=(), encrypted-media=(), execution-while-not-rendered=(), execution-while-out-of-viewport=(), fullscreen=(), geolocation=(), gyroscope=(), keyboard-map=(), magnetometer=(), microphone=(), midi=(), navigation-override=(), payment=(), picture-in-picture=(), publickey-credentials-get=(), screen-wake-lock=(), sync-xhr=(), usb=(), web-share=(), xr-spatial-tracking=()"
)

type ModifyHeadersFn = func(r *http.Request, headers map[string]string) map[string]string

type ModifyAfterUseHeadersFn = func(w http.ResponseWriter, r *http.Request)

/*
func init() {
	ModifySetDefault = func(r *http.Request, headers map[string]string) map[string]string {
		if hostBaseUrl := r.Context().Value("hostBaseUrl"); hostBaseUrl != "" {
			headers["Content-Security-Policy"] = fmt.Sprintf(
				`default-src 'self' %s https: data: 'unsafe-inline';frame-ancestors %s`,
				be.baseUrl,
				hostBaseUrl,
			)
		} else {
			headers["Content-Security-Policy"] = fmt.Sprintf(
				`default-src 'self' %s https: data: 'unsafe-inline';frame-ancestors 'none'`,
				be.baseUrl,
			)
		}
		return headers
	}
}
*/

func GetDefault() (headers map[string]string) {
	headers = map[string]string{
		"Permissions-Policy":        NoPermissionsPolicy,
		"Referrer-Policy":           "strict-origin-when-cross-origin",
		"X-Xss-Protection":          `1; mode=block`,
		"X-Content-Type-Options":    "nosniff",
		"Strict-Transport-Security": `max-age=31536000; includeSubDomains`,
		"Content-Security-Policy":   `default-src 'self' https: data: 'unsafe-inline';frame-ancestors 'none'`,
		"Cache-Control":             "no-cache",
	}
	headers["X-Content-Security-Policy"] = headers["Content-Security-Policy"]
	return
}

func SetDefault(w http.ResponseWriter, r *http.Request) {
	for k, v := range GetDefault() {
		w.Header().Set(k, v)
	}
}

func Middleware(next http.Handler) http.Handler {
	log.DebugF("including default headers middleware")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		SetDefault(w, r)
		next.ServeHTTP(w, r)
	})
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