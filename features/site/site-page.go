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

package site

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/userbase"
	"github.com/go-enjin/be/types/page"
)

func (f *CFeature) PreparePage(layout, pageType, pagePath string, t feature.Theme, r *http.Request) (pg feature.Page, ctx context.Context, err error) {
	content := feature.MakeRawPage(context.Context{
		"type":   pageType,
		"layout": layout,
	}, "")

	ctx = f.Enjin.Context()
	now := time.Now().Unix()

	if pg, err = page.New(f.Tag().String(), pagePath, content, now, now, t, ctx); err != nil {
		err = fmt.Errorf("error making new page instance: %w", err)
		return
	}

	ctx.SetSpecific("SiteMenu", f.SiteMenu(r))
	return
}

func (f *CFeature) ServePreparedPage(pg feature.Page, ctx context.Context, t feature.Theme, w http.ResponseWriter, r *http.Request) {
	handler := f.Enjin.GetServePagesHandler()
	eid := userbase.GetCurrentEID(r)
	r = feature.AddUserNotices(r, f.PullNotices(eid)...)
	if err := handler.ServePage(pg, t, ctx, w, r); err != nil {
		log.ErrorRF(r, "error serving %v prepared page: %v", f.Tag(), err)
		f.Enjin.ServeInternalServerError(w, r)
	}
}

func (f *CFeature) PrepareAndServePage(layout, pageType, pagePath string, t feature.Theme, w http.ResponseWriter, r *http.Request, custom context.Context) (err error) {
	var pg feature.Page
	var ctx context.Context
	if pg, ctx, err = f.PreparePage(layout, pageType, pagePath, t, r); err != nil {
		return
	}
	ctx.ApplySpecific(custom)
	f.ServePreparedPage(pg, ctx, t, w, r)
	return
}

func (f *CFeature) FinalizeServeRequest(w http.ResponseWriter, r *http.Request) (modified *http.Request) {
	if f.sitePath != "/" {
		if f.sitePath == r.URL.Path || strings.HasPrefix(r.URL.Path, f.sitePath+"/") {
			modified = r
			for _, uap := range feature.FilterTyped[feature.SiteAuthRequestHandler](f.siteFeatures.Features.AsFeatures()) {
				if changed := uap.FinalizeSiteRequest(w, modified); changed != nil {
					modified = changed
				}
			}
		}
	}
	return
}
