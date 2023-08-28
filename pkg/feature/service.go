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

package feature

import (
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
)

type ServiceListener interface {
	Feature
	ServiceInfo() (scheme, listen string, port int)
	StopListening() (err error)
	StartListening(router *chi.Mux, e EnjinRunner) (err error)
}

// LoggerContext is a per-request context used with ServiceLogger implementations
type LoggerContext interface {
	// URL is the parsed request URL
	URL() (parsed *url.URL)
	// Size is the number of response bytes
	Size() (size int)
	// StatusCode is the HTTP status code of the response
	StatusCode() (status int)
	// TimeStamp returns when the request was first received
	TimeStamp() (when time.Time)
	// Duration returns the total time taken to complete the request
	Duration() (duration time.Duration)
	// Request returns the request object
	Request() (r *http.Request)
}

type ServiceLogger interface {
	Feature

	RequestLogger(ctx LoggerContext) (err error)
}

type ServiceLogHandler interface {
	Feature

	LogHandler(next http.Handler) (this http.Handler)
}

type ServiceResponseLogger interface {
	Size() int
	Status() int
}