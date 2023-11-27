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

package request

import (
	"net/http"
	"net/mail"
	"net/url"
	"strings"

	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/regexps"
)

func QueryFormValue(r *http.Request, key string) (value string) {
	if r.Method == http.MethodPost {
		value = r.FormValue(key)
	} else {
		value = r.URL.Query().Get(key)
	}
	return
}

// SafeParseForm sanitizes all r.Form values and returns a flat form context.Context with only the first value of each
// key included (similar to r.FormValue)
func SafeParseForm(r *http.Request) (form map[string]interface{}) {
	_ = r.ParseForm()
	sanitized := make(url.Values)
	form = make(map[string]interface{})
	for key, values := range r.Form {
		for _, v := range values {
			sanitized[key] = append(sanitized[key], forms.StrictSanitize(v))
		}
		if len(sanitized[key]) > 0 {
			form[key] = sanitized[key][0]
		}
	}
	r.Form = sanitized
	return
}

func SafeQueryFormValue(r *http.Request, key string) (sanitized string) {
	sanitized = forms.StrictSanitize(QueryFormValue(r, key))
	return
}

func SafeQueryFormEmail(r *http.Request, key string) (address string) {
	sanitized := forms.StrictSanitize(QueryFormValue(r, key))
	if parsed, err := mail.ParseAddress(sanitized); err == nil {
		address = strings.ToLower(parsed.Address)
	}
	return
}

func SafeQueryFormHash10(r *http.Request, key string) (shasum string) {
	sanitized := forms.StrictSanitize(QueryFormValue(r, key))
	if regexps.RxHash10.MatchString(sanitized) {
		m := regexps.RxHash10.FindAllStringSubmatch(sanitized, 1)
		shasum = m[0][1]
	}
	return
}

func SafeQueryFormSixDigits(r *http.Request, key string) (digits string) {
	sanitized := forms.StrictSanitize(QueryFormValue(r, key))
	if regexps.RxSixDigits.MatchString(sanitized) {
		m := regexps.RxSixDigits.FindAllStringSubmatch(sanitized, 1)
		digits = m[0][1]
	}
	return
}
