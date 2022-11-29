//go:build page_search || pages || all

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

package search

import (
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/net/html"

	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/forms/nonce"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
)

func (f *CFeature) handleQueryRedirect(r *http.Request) (redirect string, err error) {
	tag := lang.GetTag(r)
	printer := lang.GetPrinterFromRequest(r)
	var query string
	var foundNonce, foundQuery bool
	for k, v := range r.URL.Query() {
		switch k {
		case "nonce":
			value := forms.StrictPolicy(v[0])
			if vv, e := url.QueryUnescape(value); e != nil {
				log.ErrorF("error un-escaping url path: %v", e)
			} else {
				value = vv
			}
			value = forms.Sanitize(value)
			if nonce.Validate("site-search-form", value) {
				foundNonce = true
			} else {
				err = fmt.Errorf(printer.Sprintf("search form expired"))
				break
			}
		case "query":
			query = forms.StrictPolicy(v[0])
			if vv, e := url.QueryUnescape(query); e != nil {
				log.ErrorF("error un-escaping url path: %v", e)
			} else {
				query = vv
			}
			query = forms.Sanitize(query)
			query = html.UnescapeString(query)
			foundQuery = true

		}
		if foundQuery && foundNonce || err != nil {
			break
		}
	}

	langMode := f.enjin.SiteLanguageMode()

	if ok := err == nil && foundQuery && foundNonce; ok {
		query = url.PathEscape(query)
		if query != "" {
			query = "/:" + query
		}
		redirect = langMode.ToUrl(f.enjin.SiteDefaultLanguage(), tag, f.path+query)
		// log.DebugF("search redirecting: %v", dst)
		return
	}

	if err != nil {
		log.TraceF("error handling search redirect: %v", err)
	}
	redirect = langMode.ToUrl(f.enjin.SiteDefaultLanguage(), tag, f.path)
	return
}