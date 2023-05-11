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
	"fmt"
	"net/http"

	"github.com/go-enjin/be/pkg/log"
)

type MiddlewareFn = func(next http.Handler) (this http.Handler)

type Middleware interface {
	Feature

	Apply(s System) (err error)
	Use(s System) MiddlewareFn
	Middleware(next http.Handler) http.Handler
	Serve(w http.ResponseWriter, r *http.Request) (err error)
	ServePath(path string, s System, w http.ResponseWriter, r *http.Request) (err error)
}

type UseMiddleware interface {
	Use(s System) MiddlewareFn
}

type ApplyMiddleware interface {
	Apply(s System) (err error)
}

type ServePathFeature interface {
	ServePath(path string, s System, w http.ResponseWriter, r *http.Request) (err error)
}

var _ Middleware = (*CMiddleware)(nil)

type CMiddleware struct {
	CFeature
}

func (f *CMiddleware) Apply(s System) (err error) {
	return
}

func (f *CMiddleware) Use(s System) MiddlewareFn {
	log.DebugF("using %v middleware", f.Self().Tag())
	return f.Middleware
}

func (f *CMiddleware) Middleware(next http.Handler) (this http.Handler) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := f.Serve(w, r); err != nil {
			next.ServeHTTP(w, r)
		}
	})
}

func (f *CMiddleware) Serve(w http.ResponseWriter, r *http.Request) (err error) {
	return fmt.Errorf("feature.Middleware - Serve() not implemented")
}

func (f *CMiddleware) ServePath(path string, s System, w http.ResponseWriter, r *http.Request) (err error) {
	return fmt.Errorf("feature.Middleware - ServePath() not implemented")
}