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

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/types/site"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/net/headers"
	"github.com/go-enjin/be/pkg/page"
	"github.com/go-enjin/be/pkg/theme"
)

type Service interface {
	Prefix() (prefix string)
	Context() (ctx context.Context)
	GetTheme() (t *theme.Theme, err error)
	ThemeNames() (names []string)
	ServerName() (name string)

	ServeRedirect(destination string, w http.ResponseWriter, r *http.Request)

	Serve204(w http.ResponseWriter, r *http.Request)
	Serve401(w http.ResponseWriter, r *http.Request)
	ServeBasic401(w http.ResponseWriter, r *http.Request)
	Serve403(w http.ResponseWriter, r *http.Request)
	Serve404(w http.ResponseWriter, r *http.Request)
	Serve405(w http.ResponseWriter, r *http.Request)
	Serve500(w http.ResponseWriter, r *http.Request)

	ServeNotFound(w http.ResponseWriter, r *http.Request)
	ServeInternalServerError(w http.ResponseWriter, r *http.Request)

	ServeStatusPage(status int, w http.ResponseWriter, r *http.Request)
	ServePage(p *page.Page, w http.ResponseWriter, r *http.Request) (err error)
	ServePath(urlPath string, w http.ResponseWriter, r *http.Request) (err error)
	ServeData(data []byte, mime string, w http.ResponseWriter, r *http.Request)

	FindPage(tag language.Tag, url string) (p *page.Page)

	Notify(tag string)
	NotifyF(tag, format string, argv ...interface{})
}

type System interface {
	Service

	Router() (router *chi.Mux)
}

type Internals interface {
	Service
	site.Enjin

	FindPage(tag language.Tag, url string) (p *page.Page)

	Features() (features []Feature)
	Pages() (pages map[string]*page.Page)
	Theme() (theme string)
	Theming() (theming map[string]*theme.Theme)
	Headers() (headers []headers.ModifyHeadersFn)
	Domains() (domains []string)
	Consoles() (consoles map[Tag]Console)
	Processors() (processors map[string]ReqProcessFn)
	Translators() (translators map[string]TranslateOutputFn)
	Transformers() (transformers map[string]TransformOutputFn)
	Slugsums() (enabled bool)
}