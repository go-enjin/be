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

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/page"
)

type PageContextFilterFn = func(ctx context.Context, r *http.Request) (out context.Context)

type PageContextModifier interface {
	FilterPageContext(tCtx, pCtx context.Context, r *http.Request) (out context.Context)
}

type PageRestrictionHandler interface {
	RestrictServePage(ctx context.Context, w http.ResponseWriter, r *http.Request) (co context.Context, ro *http.Request, allow bool)
}

type DataRestrictionHandler interface {
	RestrictServeData(data []byte, mime string, w http.ResponseWriter, r *http.Request) (out *http.Request, allow bool)
}

type PageProvider interface {
	FindRedirection(url string) (p *page.Page)
	FindTranslations(url string) (pages []*page.Page)
	FindPage(tag language.Tag, url string) (p *page.Page)
}