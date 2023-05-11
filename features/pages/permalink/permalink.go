// disabled //go:build page_search || pages || all

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

package permalink

import (
	"net/http"
	"regexp"

	"github.com/gofrs/uuid"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/page"
	"github.com/go-enjin/be/pkg/theme"
)

var (
	_ MakeFeature        = (*CFeature)(nil)
	_ feature.Middleware = (*CFeature)(nil)

	rxPermalinkRoot   = regexp.MustCompile(`^/([0-9a-f]{10}|[0-9a-fA-F]{8}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{12})/??$`)
	rxPermalinkedSlug = regexp.MustCompile(`-([0-9a-f]{10})$`)
)

const Tag feature.Tag = "PagesPermalink"

type Feature interface {
	feature.Middleware
}

type CFeature struct {
	feature.CMiddleware

	cli   *cli.Context
	enjin feature.Internals
}

type MakeFeature interface {
	Make() Feature
}

func New() MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.FeatureTag = Tag
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	page.RegisterMatcherFn(f._permalinkMatcher)
	theme.RegisterFuncMap("_permalink", f._permalink)
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	return
}

func (f *CFeature) Setup(enjin feature.Internals) {
	f.enjin = enjin
}

func (f *CFeature) _permalinkMatcher(path string, p *page.Page) (found string, ok bool) {
	found = path
	if p.Permalink != uuid.Nil && p.PermalinkSha != "" {
		if parsed, valid := f._parsePath(path); valid {
			switch len(parsed) {
			case 10:
				ok = parsed == p.PermalinkSha
			case 36:
				// e0f7ae8b-85e0-4c3f-b6c7-4c84b59bd3e7
				if parsedUuid := uuid.FromStringOrNil(parsed); parsedUuid != uuid.Nil {
					ok = parsedUuid.String() == parsedUuid.String()
				}
			}
		}
	}
	return
}

func (f *CFeature) _permalink(permalink uuid.UUID) (url string) {
	if permalink != uuid.Nil {
		url = "/" + permalink.String()
		for _, tag := range f.enjin.SiteLocales() {
			if f.enjin.SiteSupportsLanguage(tag) {
				if p := f.enjin.FindPage(tag, url); p != nil {
					url = p.Url + "-" + p.PermalinkSha
					return
				}
			}
		}
	}
	return
}

func (f *CFeature) _parsePath(path string) (permalink string, ok bool) {
	if ok = rxPermalinkRoot.MatchString(path); ok {
		m := rxPermalinkRoot.FindAllStringSubmatch(path, 1)
		permalink = m[0][1]
		log.TraceDF(1, "found permalink root: %v - %v", path, permalink)
	} else if ok = rxPermalinkedSlug.MatchString(path); ok {
		m := rxPermalinkedSlug.FindAllStringSubmatch(path, 1)
		permalink = m[0][1]
		log.TraceDF(1, "found permalink slug: %v - %v", path, permalink)
	}
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	f.cli = ctx
	return
}

func (f *CFeature) FilterPageContext(themeCtx, pageCtx context.Context, r *http.Request) (out context.Context) {
	out = themeCtx
	out.SetSpecific("SitePermalinkable", true)
	return
}

func (f *CFeature) Use(s feature.System) feature.MiddlewareFn {
	log.DebugF("including page permalink middleware")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := forms.TrimQueryParams(r.URL.Path)
			if _, p, ok := lang.ParseLangPath(path); ok {
				path = p
			}

			if permalink, ok := f._parsePath(path); ok {
				permalinkPath := "/" + permalink
				for _, tag := range f.enjin.SiteLocales() {
					if f.enjin.SiteSupportsLanguage(tag) {
						if p := f.enjin.FindPage(tag, permalinkPath); p != nil {
							dst := p.Url + "-" + p.PermalinkSha
							if path == dst {
								if err := f.enjin.ServePage(p, w, r); err == nil {
									return
								} else {
									log.ErrorRF(r, "error serving permalink page: %v - %v", path, err)
								}
							} else {
								http.Redirect(w, r, dst, http.StatusSeeOther)
								return
							}
						}
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}