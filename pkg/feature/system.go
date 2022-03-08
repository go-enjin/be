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

package feature

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/page"
	"github.com/go-enjin/be/pkg/theme"
)

type Service interface {
	Prefix() (prefix string)
	Context() (ctx context.Context)
	GetTheme() (t *theme.Theme, err error)
	ThemeNames() (names []string)
	ServerName() (name string)
	Serve403(w http.ResponseWriter, _ *http.Request)
	Serve404(w http.ResponseWriter, _ *http.Request)
	ServePage(p *page.Page, w http.ResponseWriter, request *http.Request) (err error)
	ServeData(data []byte, mime string, w http.ResponseWriter, _ *http.Request)
	Notify(tag string)
	NotifyF(tag, format string, argv ...interface{})
}

type System interface {
	Service

	Router() (router *chi.Mux)
}