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

package policies

import (
	"net/http"

	"github.com/urfave/cli/v2"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/net/headers/policy/csp"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "requests-pages-policies"

type Feature interface {
	feature.Feature
	feature.PrepareServePagesFeature
}

type MakeFeature interface {
	Make() Feature
}

type CFeature struct {
	feature.CFeature
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
	f.CFeature.Init(this)
	return
}

func (f *CFeature) Make() (feat Feature) {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	if err = f.CFeature.Build(b); err != nil {
		return
	}
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	return
}

func (f *CFeature) Shutdown() {
	f.CFeature.Shutdown()
}

func (f *CFeature) PrepareServePage(ctx beContext.Context, t feature.Theme, p feature.Page, w http.ResponseWriter, r *http.Request) (out beContext.Context, modified *http.Request, stop bool) {
	var pccs *csp.PageContextContentSecurity
	pccs, r = f.Enjin.ContentSecurityPolicy().PreparePageContext(t.GetConfig().ContentSecurityPolicy, ctx, r)
	ctx.SetSpecific("RequestPolicy", map[string]interface{}{
		"Permissions":     f.Enjin.PermissionsPolicy().GetRequestPolicy(r),
		"ContentSecurity": pccs,
	})

	out = ctx
	modified = r
	return
}

func (f *CFeature) FinalizeServeRequest(w http.ResponseWriter, r *http.Request) (modified *http.Request) {
	f.Enjin.PermissionsPolicy().FinalizeRequest(w, r)
	f.Enjin.ContentSecurityPolicy().FinalizeRequest(w, r)
	return
}