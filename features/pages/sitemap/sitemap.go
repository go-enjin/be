//go:build page_sitemap || pages || all

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

package sitemap

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-enjin/golang-org-x-text/language"
	"github.com/urfave/cli/v2"
	"golang.org/x/net/html"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/page"
)

var (
	_ MakeFeature        = (*CFeature)(nil)
	_ feature.Middleware = (*CFeature)(nil)
)

const Tag feature.Tag = "PagesSitemap"

var (
	DefaultSiteScheme = "https"
)

type Feature interface {
	feature.Middleware
}

type CFeature struct {
	feature.CMiddleware

	cli   *cli.Context
	enjin feature.Internals

	domain string
}

type MakeFeature interface {
	SetDomain(domain string) MakeFeature

	Make() Feature
}

func New() MakeFeature {
	f := new(CFeature)
	f.Init(f)
	return f
}

func (f *CFeature) SetDomain(domain string) MakeFeature {
	f.domain = domain
	return f
}

func (f *CFeature) Make() Feature {
	if f.domain != "" && !strings.HasPrefix(f.domain, "http://") && !strings.HasPrefix(f.domain, "https://") {
		log.FatalDF(1, "http:// or https:// required for sitemap domain setting")
	}
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CMiddleware.Init(this)
}

func (f *CFeature) Tag() (tag feature.Tag) {
	tag = Tag
	return
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	return
}

func (f *CFeature) Setup(enjin feature.Internals) {
	f.enjin = enjin
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	f.cli = ctx
	return
}

// func (f *CFeature) FilterPageContext(themeCtx, pageCtx context.Context, r *http.Request) (out context.Context) {
// 	out = themeCtx
// 	out.SetSpecific("SiteSearchable", true)
// 	return
// }

func (f *CFeature) Use(s feature.System) feature.MiddlewareFn {
	log.DebugF("including page sitemap middleware")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			langMode := f.enjin.SiteLanguageMode()
			defaultTag := f.enjin.SiteDefaultLanguage()

			path := forms.SanitizeRequestPath(r.URL.Path)
			if _, p, ok := lang.ParseLangPath(path); ok {
				path = p
			}

			if path == "/sitemap.xml" {

				var domain string
				if domain = f.domain; domain == "" {
					domain = DefaultSiteScheme + "://" + r.Host
				}

				pages := make(map[string]*page.Page)
				for _, found := range f.enjin.FindPages("/") {
					if ignored := found.Context.String("SitemapIgnored", "false"); ignored != "true" {
						priority := found.Context.Float64("SitemapPriority", 0.5)
						found.Context.SetSpecific("SitemapPriority", priority)
						if changeFreq := found.Context.String("SitemapChangeFreq", ""); changeFreq != "" {
							switch changeFreq {
							case "always", "hourly", "daily", "weekly", "monthly", "yearly", "never":
								found.Context.SetSpecific("SitemapChangeFreq", changeFreq)
							default:
								log.ErrorF("error: page has invalid sitemap-change-freq: %v", changeFreq)
								found.Context.Delete("SitemapChangeFreq")
							}
						}

						tag := found.LanguageTag
						if language.Compare(tag, language.Und) {
							tag = defaultTag
						}

						fullUrl := langMode.ToUrl(defaultTag, tag, found.Url)
						if !strings.HasPrefix(fullUrl, "http") && domain != "" {
							fullUrl = domain + fullUrl
						}

						pages[fullUrl] = found
					}
				}

				var contents string
				contents += `<?xml version="1.0" encoding="UTF-8"?>` + "\n"
				contents += `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n"

				for _, fullUrl := range maps.SortedKeys(pages) {
					pg := pages[fullUrl]
					contents += "\t<url>\n"
					contents += "\t\t<loc>" + html.EscapeString(fullUrl) + "</loc>\n"
					contents += "\t\t<lastmod>" + pg.UpdatedAt.Format("2006-01-02") + "</lastmod>\n"
					if priority := pg.Context.Float64("SitemapPriority", -1.0); priority >= 0.0 {
						contents += fmt.Sprintf("\t\t<priority>%0.1f</priority>\n", priority)
					}
					if changeFreq := pg.Context.String("SitemapChangeFreq", ""); changeFreq != "" {
						contents += fmt.Sprintf("\t\t<changefreq>%s</changefreq>\n", changeFreq)
					}
					contents += "\t</url>\n"
				}

				contents += `</urlset>`

				f.enjin.ServeData([]byte(contents), "application/xml", w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}