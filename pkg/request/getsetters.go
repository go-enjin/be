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

package request

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

const (
	KeyEnjinID  Key = "enjin-id"
	KeyHomePath Key = "home-path"
)

func GetRequestID(r *http.Request) (id string) {
	id, _ = String(r, middleware.RequestIDKey)
	return
}

func GetEnjinID(r *http.Request) (name string) {
	name, _ = String(r, KeyEnjinID)
	return
}

func SetHomePath(r *http.Request, path string) (modified *http.Request) {
	if path == "" {
		path = "/"
	}
	modified = Set(r, KeyHomePath, path)
	return
}

func GetHomePath(r *http.Request) (path string) {
	path, _ = String(r, KeyHomePath)
	return
}