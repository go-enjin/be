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

package net

import (
	"net/http"
)

func BaseURL(r *http.Request) (baseUrl string) {
	if r.URL.Scheme != "" {
		baseUrl = r.URL.Scheme
	} else if r.TLS != nil {
		baseUrl = "https"
	} else {
		baseUrl = "http"
	}
	baseUrl += "://"
	if r.URL.Host != "" {
		baseUrl += r.URL.Host
		if portNum := r.URL.Port(); portNum != "" {
			baseUrl += ":" + portNum
		}
	} else {
		baseUrl += r.Host
	}
	return
}

func MakeUrl(path string, r *http.Request) (url string) {
	pathLen := len(path)
	fullpath := "/"
	if pathLen > 0 && path[0] != '/' {
		fullpath = "/" + path
	}
	url = BaseURL(r) + fullpath
	return
}
