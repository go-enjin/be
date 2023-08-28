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

package context

import (
	"net/http"
	"net/url"
	"time"

	"github.com/go-enjin/be/pkg/feature"
)

var (
	_ feature.LoggerContext = (*CLoggerContext)(nil)
)

type CLoggerContext struct {
	r        *http.Request
	url      *url.URL
	size     int
	status   int
	when     time.Time
	duration time.Duration
}

func New(r *http.Request, size, status int, when time.Time, duration time.Duration) (c feature.LoggerContext, err error) {
	ctx := &CLoggerContext{
		r:        r,
		size:     size,
		status:   status,
		when:     when,
		duration: duration,
	}
	if ctx.url, err = url.ParseRequestURI(r.RequestURI); err != nil {
		return
	}
	c = ctx
	return
}

func (c *CLoggerContext) URL() (parsed *url.URL) {
	parsed = c.url
	return
}

func (c *CLoggerContext) Size() (size int) {
	size = c.size
	return
}

func (c *CLoggerContext) StatusCode() (status int) {
	status = c.status
	return
}

func (c *CLoggerContext) TimeStamp() (when time.Time) {
	when = c.when
	return
}

func (c *CLoggerContext) Duration() (duration time.Duration) {
	duration = c.duration
	return
}

func (c *CLoggerContext) Request() (r *http.Request) {
	r = c.r
	return
}