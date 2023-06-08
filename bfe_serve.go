//go:build fastcgi || all

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

package be

import (
	"net/http"

	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net/serve"
)

func (fe *FastcgiEnjin) ServeNotFound(w http.ResponseWriter, r *http.Request) {
	fe.ServeStatusPage(404, w, r)
}

func (fe *FastcgiEnjin) ServeInternalServerError(w http.ResponseWriter, r *http.Request) {
	fe.ServeStatusPage(500, w, r)
}

func (fe *FastcgiEnjin) ServeStatusPage(status int, w http.ResponseWriter, r *http.Request) {
	r = serve.SetServeStatus(status, r)

	if path, ok := fe.feb.statusPages[status]; ok && path != "" {
		serve.Redirect(path, w, r)
		return
	}

	switch status {
	case 401:
		serve.Serve401(w, r)
	case 403:
		serve.Serve403(w, r)
	case 404:
		serve.Serve404(w, r)
	case 405:
		serve.Serve405(w, r)
	case 500:
		serve.Serve500(w, r)
	default:
		log.WarnF("unsupported status page: %v, serving 404 instead", status)
		fe.ServeStatusPage(404, w, r)
	}
}