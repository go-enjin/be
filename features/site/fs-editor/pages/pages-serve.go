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

package pages

import (
	errors2 "errors"
	"net/http"

	"github.com/go-corelibs/x-text/message"
	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
)

func (f *CFeature) ServePreparedEditPage(pg feature.Page, ctx context.Context, w http.ResponseWriter, r *http.Request) {
	handler := f.Enjin.GetServePagesHandler()
	t := f.Enjin.MustGetTheme()
	ctx.SetSpecific("PageArchetypes", t.ListArchetypes())
	if err := handler.ServePage(pg, f.Editor.SiteFeatureTheme(), ctx, w, r); err != nil {
		log.ErrorRF(r, "error serving %v editor generic page: %v", f.Tag(), err)
		f.Enjin.ServeInternalServerError(w, r)
	}
}

func (f *CFeature) ServePreviewEditPage(pg feature.Page, ctx context.Context, w http.ResponseWriter, r *http.Request) {
	printer := message.GetPrinter(r)

	ctx.Delete("SiteMenu")

	if ee := f.PageRenderCheck(pg); ee != nil {
		var contents string
		var enjErr *errors.EnjinError
		if errors2.As(ee, &enjErr) {
			contents = string(enjErr.Html())
		} else {
			contents = "<p>" + printer.Sprintf(`(this page failed to render)`) + "</p>"
		}
		pg.SetContent(contents)
		pg.SetArchetype("")
		pg.SetFormat("html")
	}

	handler := f.Enjin.GetServePagesHandler()
	if err := handler.ServePage(pg, f.Enjin.MustGetTheme(), ctx, w, r); err != nil {
		log.ErrorRF(r, "error serving %v editor preview page: %v", f.Tag(), err)
		f.Enjin.ServeInternalServerError(w, r)
	}
}
