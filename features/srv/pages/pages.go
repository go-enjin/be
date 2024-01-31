//go:build srv_pages || srv || all

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
	"context"
	"net/http"

	"github.com/urfave/cli/v2"

	"github.com/go-corelibs/x-text/message"
	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/feature/signaling"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net"
	"github.com/go-enjin/be/pkg/net/serve"
	"github.com/go-enjin/be/pkg/request"
	"github.com/go-enjin/be/pkg/request/argv"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "srv-pages"

type Feature interface {
	feature.Feature
	feature.RoutePagesHandler
	feature.ServePagesHandler
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
	f.CFeature.Construct(f)
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
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	return
}

func (f *CFeature) Shutdown() {

}

func (f *CFeature) RoutePage(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	tag := message.GetTag(r)

	// look for any page provider providing the requested page
	for _, pp := range feature.FilterTyped[feature.PageProvider](f.Enjin.Features().List()) {
		if pg := pp.FindPage(r, tag, path); pg != nil {
			if err := f.Enjin.ServePage(pg, w, r); err == nil {
				log.DebugRF(r, "enjin router served provided page: %v", pg.Url())
				f.Enjin.Emit(signaling.SignalServePage, pp.Tag().String(), pg)
				return
			} else {
				log.ErrorRF(r, "error serving provided page: %v - %v", pg.Url(), err)
			}
		}
	}

	// look for any serve-path feature handling the requested page
	for _, spf := range feature.FilterTyped[feature.ServePathFeature](f.Enjin.Features().List()) {
		if ee := spf.ServePath(path, f.Enjin.(feature.System), w, r); ee == nil {
			log.DebugRF(r, "%v feature served path: %v", spf.Tag(), path)
			f.Enjin.Emit(signaling.SignalServePath, spf.Tag().String(), path)
			return
		}
	}

	// look for any fallback, enjin-built-in, pages
	if pages := f.Enjin.Pages(); len(pages) > 0 {
		if pg, ok := pages[path]; ok {
			if err := f.Enjin.ServePage(pg, w, r); err != nil {
				log.ErrorRF(r, "serve page err: %v", err)
				f.Enjin.ServeInternalServerError(w, r)
			} else {
				log.DebugRF(r, "enjin router served page: %v", path)
				f.Enjin.Emit(signaling.SignalServePage, feature.EnjinTag.String(), pg)
			}
			return
		}
	}

	f.Enjin.ServeNotFound(w, r)
}

func (f *CFeature) ServePage(p feature.Page, t feature.Theme, ctx beContext.Context, w http.ResponseWriter, r *http.Request) (err error) {

	for _, ptp := range f.Enjin.GetPageTypeProcessors() {
		var pg feature.Page
		var redirect string
		var processed bool
		if pg, redirect, processed, err = ptp.ProcessRequestPageType(r, p); err != nil {
			return
		} else if redirect != "" {
			f.Enjin.ServeRedirect(redirect, w, r)
			return
		} else if processed {
			p = pg
			//break
		}
	}

	pUrl := p.Url()

	ctx.SetSpecific("Theme", t)
	ctx.SetSpecific("BaseUrl", net.BaseURL(r))
	ctx.SetSpecific("HomePath", request.GetHomePath(r))
	ctx.SetSpecific("UserNotices", feature.GetUserNotices(r))

	ctx.SetSpecific("R", r)
	ctx.SetSpecific(argv.RequestKey.String(), argv.Get(r))

	for _, pspf := range f.Enjin.GetPrepareServePagesFeatures() {
		if out, modified, handled := pspf.PrepareServePage(ctx, t, p, w, r); handled {
			log.DebugF("%v feature handled serve page early", pspf.Tag())
			return
		} else {
			if len(out) > 0 {
				ctx = out
			}
			if modified != nil {
				r = modified
			}
		}
	}

	var data []byte
	var redirect string

	renderer := f.Enjin.GetThemeRenderer(ctx)

	if data, redirect, err = renderer.RenderPage(t, ctx, p); err != nil {
		log.ErrorRF(r, "error rendering page: %v - %v", pUrl, err)
		return
	} else if redirect != "" {
		log.DebugRF(r, "redirecting from RenderPage: %v - %v", pUrl, redirect)
		f.Enjin.ServeRedirect(redirect, w, r)
		return
	}
	if cacheControl := p.Context().String("CacheControl", ""); cacheControl != "" {
		r = serve.SetCacheControl(cacheControl, w, r)
	}
	mime := ctx.String("ContentType", "text/html; charset=utf-8")
	contentDisposition := ctx.String("ContentDisposition", "inline")
	r = r.Clone(context.WithValue(r.Context(), "Content-Disposition", contentDisposition))

	f.Enjin.ServeData(data, mime, w, r)
	return
}
