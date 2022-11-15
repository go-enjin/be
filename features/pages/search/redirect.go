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

	"github.com/go-enjin/golang-org-x-text/language"
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
			value := forms.StripTags(v[0])
			if vv, e := url.QueryUnescape(value); e != nil {
				log.ErrorF("error un-escaping url path: %v", e)
			} else {
				value = vv
			}
			value = forms.Sanitize(value)
			if !nonce.Validate("site-search-form", value) {
				// Let the visitor know that their search form expired
				err = fmt.Errorf(printer.Sprintf("search form expired"))
				return
			}
			foundNonce = true
			if foundQuery {
				break
			}
		case "query":
			// trap random page ?query= requests?
			query = forms.StripTags(v[0])
			query = forms.Sanitize(query)
			if vv, e := url.QueryUnescape(query); e != nil {
				log.ErrorF("error un-escaping url path: %v", e)
			} else {
				query = vv
			}
			query = forms.Sanitize(query)
			query = html.UnescapeString(query)
			// query = html.EscapeString(query)
			// query = url.PathEscape(query)
			foundQuery = true
			if foundNonce {
				break
			}
		}
	}
	if foundQuery && !foundNonce {
		// Let the user know that the search form submission resulted in a (generic) error
		err = fmt.Errorf(printer.Sprintf("search form error"))
		return
	}
	if ok := foundQuery && foundNonce; ok {
		var dst string
		query = url.PathEscape(query)
		if !language.Compare(tag, f.enjin.SiteDefaultLanguage()) {
			dst = "/" + tag.String() + f.path + "/:" + query
		} else {
			dst = f.path + "/:" + query
		}
		// log.DebugF("search redirecting: %v", dst)
		// http.Redirect(w, r, dst, http.StatusSeeOther)
		redirect = dst
	}
	return
}