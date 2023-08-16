//go:build srv_fastcgi || all

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

package fastcgi

import (
	"net/http"

	"github.com/go-enjin/be/pkg/log"
	handlers "github.com/go-enjin/be/pkg/net/gorilla-handlers"
)

type logger struct {
	next http.Handler
}

func (l *logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h := handlers.LoggingHandler(log.InfoWriter(), l.next)
	h.ServeHTTP(w, r)
}