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

	"github.com/go-corelibs/slices"
	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
)

var (
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "pages-robots"

type Feature interface {
	feature.Feature
	feature.RequestModifier
	feature.ApplyMiddleware
}

type MakeFeature interface {
	AddSitemap(sitemap string) MakeFeature
	AddRuleGroup(rule RuleGroup) MakeFeature

	SiteRobotsHeader(content string) MakeFeature
	SiteRobotsMetaTag(content string) MakeFeature

	Make() Feature
}

type CFeature struct {
	feature.CFeature

	rules    []RuleGroup
	sitemaps []string

	siteHeader  string
	siteMetaTag string

	metaRobots string
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
}

func (f *CFeature) AddSitemap(sitemap string) MakeFeature {
	sitemap = strings.TrimSpace(sitemap)
	if !slices.Within(sitemap, f.sitemaps) {
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
	category := f.Tag().String()
	b.AddFlags(
		&cli.StringFlag{
			Name:     "meta-robots",
			Usage:    "set a site-wide <meta name=\"robots\"/> head tag",
			EnvVars:  b.MakeEnvKeys("META_ROBOTS"),
			Category: category,
		},
		&cli.StringFlag{
			Name:     "x-robots-tag",
			Usage:    "set a site-wide X-Robots-Tag response header",
			EnvVars:  b.MakeEnvKeys("X_ROBOTS_TAG"),
			Category: category,
		},
	)
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	if xrt := ctx.Value("x-robots-tag"); xrt != nil {
		if xRobotsTag, ok := xrt.(string); ok && xRobotsTag != "" {
			f.siteHeader = xRobotsTag
		}
	}
	if mr := ctx.Value("meta-robots"); mr != nil {
		f.metaRobots, _ = mr.(string)
	}
	return
}

func (f *CFeature) FilterPageContext(themeCtx, pageCtx context.Context, r *http.Request) (out context.Context) {
	out = themeCtx

	if f.metaRobots != "" {
		var found []template.HTML
		if existing, ok := out.Get("HtmlHeadTags").([]template.HTML); ok {
			found = pruneRobotsFromHtmlHeadTags(existing)
		}
		found = append(found, template.HTML(fmt.Sprintf(`<meta name="robots" content="%s"/>`, f.metaRobots)))
		out.SetSpecific("HtmlHeadTags", found)
		return
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

func (f *CFeature) Apply(s feature.System) (err error) {
	s.Router().Get("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
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
		f.Enjin.ServeData([]byte(contents), "text/plain", w, r)
	})
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
