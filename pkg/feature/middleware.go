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

package feature

import (
	"net/http"
)

type MiddlewareFn = func(next http.Handler) (this http.Handler)

type UseMiddleware interface {
	Feature
	Use(s System) MiddlewareFn
}

type ApplyMiddleware interface {
	Feature
	Apply(s System) (err error)
}

type ServePathFeature interface {
	Feature
	ServePath(path string, s System, w http.ResponseWriter, r *http.Request) (err error)
}
