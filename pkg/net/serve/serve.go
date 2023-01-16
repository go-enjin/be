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

package serve

import (
	"net/http"

	"github.com/go-enjin/be/pkg/lang"
)

func Redirect(destination string, w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, destination, http.StatusSeeOther)
}

func Serve204(w http.ResponseWriter, r *http.Request) {
	printer := lang.GetPrinterFromRequest(r)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNoContent)
	// The request was successful and there is no content in the response
	_, _ = w.Write([]byte("204 - " + printer.Sprintf("No Content")))
}

func Serve401(w http.ResponseWriter, r *http.Request) {
	printer := lang.GetPrinterFromRequest(r)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	// The request was rejected due to being an Unauthorized user (or anonymous guest)
	_, _ = w.Write([]byte("401 - " + printer.Sprintf("Unauthorized")))
}

func ServeBasic401(w http.ResponseWriter, r *http.Request) {
	printer := lang.GetPrinterFromRequest(r)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("WWW-Authenticate", "Basic")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte("401 - " + printer.Sprintf("Unauthorized")))
}

func Serve403(w http.ResponseWriter, r *http.Request) {
	printer := lang.GetPrinterFromRequest(r)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)
	// The request was rejected due to being Forbidden to the user (or anonymous guest)
	_, _ = w.Write([]byte("403 - " + printer.Sprintf("Forbidden")))
}

func Serve404(w http.ResponseWriter, r *http.Request) {
	printer := lang.GetPrinterFromRequest(r)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	// The request was rejected due to the page requested not existing
	_, _ = w.Write([]byte("404 - " + printer.Sprintf("Not Found")))
}

func Serve405(w http.ResponseWriter, r *http.Request) {
	printer := lang.GetPrinterFromRequest(r)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusMethodNotAllowed)
	// The request was rejected due to the method used bin not allowed
	_, _ = w.Write([]byte("405 - " + printer.Sprintf("Method Not Allowed")))
}

func Serve500(w http.ResponseWriter, r *http.Request) {
	printer := lang.GetPrinterFromRequest(r)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	// The request resulted in an internal server error
	_, _ = w.Write([]byte("500 - " + printer.Sprintf("Internal Server Error")))
}