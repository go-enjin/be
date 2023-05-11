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
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

var (
	_ MakeFeature             = (*CFeature)(nil)
	_ feature.Middleware      = (*CFeature)(nil)
	_ feature.RequestModifier = (*CFeature)(nil)
)

const Tag feature.Tag = "pages-robots"

type Feature interface {
	feature.Middleware
}

type CFeature struct {
	feature.CMiddleware

	cliCtx *cli.Context
	enjin  feature.Internals

	rules    []RuleGroup
	sitemaps []string

	siteHeader  string
	siteMetaTag string
}

type MakeFeature interface {
	AddSitemap(sitemap string) MakeFeature
	AddRuleGroup(rule RuleGroup) MakeFeature

	SiteRobotsHeader(content string) MakeFeature
	SiteRobotsMetaTag(content string) MakeFeature

	Make() Feature
}

func New() MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.FeatureTag = Tag
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CMiddleware.Init(this)
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

func (f *CFeature) SiteRobotsHeader(content string) MakeFeature {
	f.siteHeader = content
	return f
}

func (f *CFeature) SiteRobotsMetaTag(content string) MakeFeature {
	f.siteMetaTag = content
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	if f.siteMetaTag != "" {
		b.AddHtmlHeadTag("meta", map[string]string{
			"name":    "robots",
			"content": f.siteMetaTag,
		})
	}
	b.AddFlags(
		&cli.StringFlag{
			Name:    "meta-robots",
			Usage:   "set a site-wide <meta name=\"robots\"/> head tag",
			EnvVars: b.MakeEnvKeys("META_ROBOTS"),
		},
		&cli.StringFlag{
			Name:    "x-robots-tag",
			Usage:   "set a site-wide X-Robots-Tag response header",
			EnvVars: b.MakeEnvKeys("X_ROBOTS_TAG"),
		},
	)
	return
}

func (f *CFeature) Setup(enjin feature.Internals) {
	f.enjin = enjin
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	f.cliCtx = ctx
	if xrt := f.cliCtx.Value("x-robots-tag"); xrt != nil {
		if xRobotsTag, ok := xrt.(string); ok && xRobotsTag != "" {
			f.siteHeader = xRobotsTag
		}
	}
	return
}

func pruneRobotsFromHtmlHeadTags(existing []template.HTML) (found []template.HTML) {
	for _, mt := range existing {
		if !strings.Contains(string(mt), `name="robots"`) {
			found = append(found, mt)
		}
	}
	return
}

func (f *CFeature) FilterPageContext(themeCtx, pageCtx context.Context, r *http.Request) (out context.Context) {
	out = themeCtx
	if mr := f.cliCtx.Value("meta-robots"); mr != nil {
		if metaRobots, ok := mr.(string); ok && metaRobots != "" {
			var found []template.HTML
			if existing, ok := out.Get("HtmlHeadTags").([]template.HTML); ok {
				found = pruneRobotsFromHtmlHeadTags(existing)
			}
			found = append(found, template.HTML(fmt.Sprintf(`<meta name="robots" content="%s"/>`, metaRobots)))
			out.SetSpecific("HtmlHeadTags", found)
			return
		}
	}
	if pr := out.Get("Robots"); pr != nil {
		if pgRobots, ok := pr.(string); ok {
			var found []template.HTML
			if existing, ok := out.Get("HtmlHeadTags").([]template.HTML); ok {
				found = pruneRobotsFromHtmlHeadTags(existing)
			}
			found = append(found, template.HTML(fmt.Sprintf(`<meta name="robots" content="%s"/>`, pgRobots)))
			out.SetSpecific("HtmlHeadTags", found)
		}
	}
	return
}

func (f *CFeature) ModifyRequest(w http.ResponseWriter, r *http.Request) {
	if f.siteHeader != "" {
		w.Header().Set("X-Robots-Tag", f.siteHeader)
	}
}

func (f *CFeature) Use(s feature.System) feature.MiddlewareFn {
	log.DebugF("including page robots middleware")

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