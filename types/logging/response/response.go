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

package response

import (
	"bufio"
	"net"
	"net/http"

	"github.com/felixge/httpsnoop"

	"github.com/go-enjin/be/pkg/feature"
)

var (
	_ feature.ServiceResponseLogger = (*CLogger)(nil)
)

type CLogger struct {
	w      http.ResponseWriter
	size   int
	status int
}

func NewLogger(w http.ResponseWriter) (logger feature.ServiceResponseLogger, writer http.ResponseWriter) {
	l := &CLogger{
		w:      w,
		status: http.StatusOK,
	}
	logger = l
	writer = httpsnoop.Wrap(w, httpsnoop.Hooks{
		Write: func(httpsnoop.WriteFunc) httpsnoop.WriteFunc {
			return l.Write
		},
		WriteHeader: func(httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
			return l.WriteHeader
		},
	})
	return
}

func (l *CLogger) Size() int {
	return l.size
}

func (l *CLogger) Status() int {
	return l.status
}

func (l *CLogger) Write(b []byte) (int, error) {
	size, err := l.w.Write(b)
	l.size += size
	return size, err
}

func (l *CLogger) WriteHeader(s int) {
	l.w.WriteHeader(s)
	l.status = s
}

func (l *CLogger) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	conn, rw, err := l.w.(http.Hijacker).Hijack()
	if err == nil && l.status == 0 {
		// The status will be StatusSwitchingProtocols if there was no error and
		// WriteHeader has not been called yet
		l.status = http.StatusSwitchingProtocols
	}
	return conn, rw, err
}