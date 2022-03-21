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
	"regexp"
)

func BaseURL(r *http.Request) (baseUrl string) {
	u := r.URL
	if baseUrl = u.Scheme; baseUrl == "" {
		baseUrl = "https"
	}
	baseUrl += "://"
	if u.Host != "" {
		baseUrl += u.Host
	} else {
		baseUrl += r.Host
	}
	if portNum := u.Port(); portNum != "" {
		baseUrl += ":" + portNum
	}
	return
}

var RxQueryParams = regexp.MustCompile(`\?(.+?)\s*$`)

func TrimQueryParams(path string) string {
	if RxQueryParams.MatchString(path) {
		path = RxQueryParams.ReplaceAllString(path, "")
	}
	return path
}

func TrimTrailingSlash(url string) string {
	if lu := len(url); lu > 0 {
		if url[lu-1] == '/' {
			url = url[0 : lu-1]
		}
	}
	return url
}