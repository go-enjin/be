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
	"github.com/urfave/cli/v2"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/types/site"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "site-dashboard"

type Feature interface {
	feature.SiteFeature
}

type MakeFeature interface {
	feature.SiteMakeFeature[MakeFeature]

	Make() Feature
}

type CFeature struct {
	site.CSiteFeature[feature.Feature, MakeFeature]
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.SetSiteFeatureName(tag.String())
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CSiteFeature.Init(this)
	return
}

func (f *CFeature) Make() (feat Feature) {
	return f
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CSiteFeature.Startup(ctx); err != nil {
		return
	}
	return
}

func (f *CFeature) RouteSiteFeature(r chi.Router) {
	r.Get("/", f.RenderDashboard)
}

func (f *CFeature) RenderDashboard(w http.ResponseWriter, r *http.Request) {
	var err error
	var pg feature.Page
	var ctx beContext.Context
	t := f.SiteFeatureTheme()

	if pg, ctx, err = f.Site().PreparePage("site", "dashboard", f.SiteFeaturePath(), t); err != nil {
		log.ErrorRF(r, "error preparing %v dashboard page: %v", f.Tag(), err)
		f.Enjin.ServeNotFound(w, r)
		return
	}
	printer := lang.GetPrinterFromRequest(r)

	ctx.SetSpecific("EnjinContext", f.Enjin.Context().Copy())
	ctx.SetSpecific("EnjinFeatures", f.Enjin.Features())

	pg.SetTitle(printer.Sprintf("Dashboard"))
	f.Site().ServePreparedPage(pg, ctx, t, w, r)
}