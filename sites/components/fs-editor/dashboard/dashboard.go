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

package dashboard

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/types/editor"
)

var (
	DefaultEditorType = "dashboard"
	DefaultEditorName = "dashboard"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "editor-dashboard"

type Feature interface {
	feature.EditorFeature
}

type MakeFeature interface {
	feature.EditorMakeFeature[MakeFeature]

	Make() Feature
}

type CFeature struct {
	editor.CEditorFeature[MakeFeature]
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CEditorFeature.Init(this)
	f.CEditorFeature.EditorName = DefaultEditorName
	f.CEditorFeature.EditorType = DefaultEditorType
	return
}

func (f *CFeature) Make() (feat Feature) {
	return f
}

func (f *CFeature) SetupEditorRoute(r chi.Router) {
	r.Get("/", f.RenderDashboard)
}

func (f *CFeature) RenderDashboard(w http.ResponseWriter, r *http.Request) {

	var err error
	var pg feature.Page
	var ctx beContext.Context

	if pg, ctx, err = f.SelfEditor().PrepareEditPage("fs-editor--dashboard", f.EditorType, ""); err != nil {
		log.ErrorRF(r, "error preparing %v editor page: %v", f.Tag(), err)
		f.Enjin.ServeNotFound(w, r)
		return
	}
	printer := lang.GetPrinterFromRequest(r)

	ctx.SetSpecific("EnjinContext", f.Enjin.Context().Copy())
	ctx.SetSpecific("EnjinFeatures", f.Enjin.Features())

	pg.SetTitle(printer.Sprintf("Dashboard"))
	f.SelfEditor().ServePreparedEditPage(pg, ctx, w, r)
}