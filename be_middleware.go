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

package be

import (
	"net/http"

	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	beNet "github.com/go-enjin/be/pkg/net"
	"github.com/go-enjin/be/pkg/net/headers"
	"github.com/go-enjin/be/pkg/slices"
)

func (e *Enjin) requestFiltersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		remoteAddr, _ := beNet.GetIpFromRequest(r)
		for _, rf := range e.eb.fRequestFilters {
			if err := rf.FilterRequest(r); err != nil {
				log.WarnRF(r, "filtering request from: %v - %v", remoteAddr, err)
				e.ServeNotFound(w, r)
				return
			} else {
				log.DebugRF(r, "allowing request from: %v", remoteAddr)
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (e *Enjin) domainsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(e.eb.domains) > 0 {
			if !slices.Present(r.Host, e.eb.domains...) {
				log.WarnRF(r, "rejecting unsupported domain: %v", r.Host)
				e.ServeNotFound(w, r)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (e *Enjin) headersMiddleware(next http.Handler) http.Handler {
	return headers.ModifyMiddleware(e.modifyHeadersFn)(next)
}

func (e *Enjin) modifyHeadersFn(request *http.Request, headers map[string]string) map[string]string {
	for _, fn := range e.eb.headers {
		headers = fn(request, headers)
	}
	return headers
}

func (e *Enjin) redirectionMiddleware(next http.Handler) http.Handler {
	log.DebugF("including redirection middleware")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := forms.CleanRequestPath(r.URL.Path)

		if rp := e.FindRedirection(path); rp != nil {
			langMode := e.SiteLanguageMode()
			reqLang := lang.GetTag(r)
			dst := langMode.ToUrl(e.SiteDefaultLanguage(), reqLang, rp.Url())
			log.DebugRF(r, "redirecting from %v to %v", path, dst)
			e.ServeRedirect(dst, w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}
