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

package deny

import (
	"net/http"

	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net"
)

func DenyWordPressPaths() {
	_ = DenyPath(`/wp-(admin|login\.php|login|includes|content)`)
	_ = DenyPath(`/xmlrpc.php`)
}

func DenyPath(patterns ...string) (err error) {
	for _, pattern := range patterns {
		if err = _manager.Restrict(pattern); err != nil {
			// stop on first error
			return
		}
	}
	return
}

func DenyAddress(address string) {
	_manager.Deny(address)
	log.WarnF(`blocking address: "%v"`)
}

func IsAddressDenied(address string) bool {
	return _manager.Denied(address)
}

func CheckRequestDenied(req *http.Request) (address string, denied bool) {
	var err error
	var addr string
	if addr, err = net.GetIpFromRequest(req); err != nil {
		return err.Error(), true
	}
	if _manager.Denied(addr) {
		return addr, true
	} else if _manager.Restricted(req.URL.Path) {
		_manager.Deny(addr)
		return addr, true
	}
	return addr, false
}

func Serve403(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte("404 - Not Found"))
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if address, denied := CheckRequestDenied(r); denied {
			Serve403(w, r)
			log.DebugF(`denied: "%v" (%v)`, r.URL.Path, address)
			return
		}
		next.ServeHTTP(w, r)
	})
}