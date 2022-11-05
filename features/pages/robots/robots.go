//go:build page_robots || pages || all

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

package robots

import (
	"net/http"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

var (
	_ MakeFeature        = (*CFeature)(nil)
	_ feature.Middleware = (*CFeature)(nil)
)

const Tag feature.Tag = "PagesRobots"

type Feature interface {
	feature.Middleware
}

type CFeature struct {
	feature.CMiddleware

	cli   *cli.Context
	enjin feature.Internals

	rules    []RuleGroup
	sitemaps []string

	siteRobots string
}

type MakeFeature interface {
	AddSitemap(sitemap string) MakeFeature
	AddRuleGroup(rule RuleGroup) MakeFeature
	SiteRobotsMetaTag(content string) MakeFeature

	Make() Feature
}

func New() MakeFeature {
	f := new(CFeature)
	f.Init(f)
	return f
}

func (f *CFeature) AddSitemap(sitemap string) MakeFeature {
	sitemap = strings.TrimSpace(sitemap)
	if !beStrings.StringInSlices(sitemap, f.sitemaps) {
		f.sitemaps = append(f.sitemaps, sitemap)
	}
	return f
}

func (f *CFeature) AddRuleGroup(rule RuleGroup) MakeFeature {
	f.rules = append(f.rules, rule)
	return f
}

func (f *CFeature) SiteRobotsMetaTag(content string) MakeFeature {
	f.siteRobots = content
	return f
}

func (f *CFeature) Make() Feature {
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
	if f.siteRobots != "" {
		b.AddHtmlHeadTag("meta", map[string]string{
			"name":    "robots",
			"content": f.siteRobots,
		})
	}
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
	log.DebugF("including page search middleware")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := forms.SanitizeRequestPath(r.URL.Path)
			if _, p, ok := lang.ParseLangPath(path); ok {
				path = p
			}

			if path == "/robots.txt" && len(f.rules) > 0 {
				var contents string
				for idx, rule := range f.rules {
					if idx > 0 {
						contents += "\n"
					}
					contents += rule.String()
				}
				if len(f.sitemaps) > 0 {
					if contents != "" {
						contents += "\n"
					}
					for _, sitemap := range f.sitemaps {
						contents += "Sitemap: " + sitemap + "\n"
					}
				}
				f.enjin.ServeData([]byte(contents), "text/plain", w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}